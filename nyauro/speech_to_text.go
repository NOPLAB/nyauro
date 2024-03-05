package nyauro

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"log"
	"runtime"

	speech "cloud.google.com/go/speech/apiv2"
	"cloud.google.com/go/speech/apiv2/speechpb"
	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SpeechToTextWorkerArgument struct {
	Recognizer        string
	SpeechToTextModel string
}

func SpeechToTextWorker(d *discordgo.Session, i *discordgo.InteractionCreate, args RunArgument) {
	v, err := JoinVoiceChannel(d, i)
	if err != nil {
		return
	}
	// defer v.Disconnect()

	d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})

	ctx := context.Background()
	c, err := speech.NewClient(ctx)
	if err != nil {
		log.Println("Error creating speech client: ", err)
		return
	}
	defer c.Close()

	config_req := &speechpb.StreamingRecognizeRequest{
		StreamingRequest: &speechpb.StreamingRecognizeRequest_StreamingConfig{
			StreamingConfig: &speechpb.StreamingRecognitionConfig{
				Config: &speechpb.RecognitionConfig{
					DecodingConfig: &speechpb.RecognitionConfig_ExplicitDecodingConfig{
						ExplicitDecodingConfig: &speechpb.ExplicitDecodingConfig{
							Encoding:          speechpb.ExplicitDecodingConfig_LINEAR16,
							SampleRateHertz:   44100,
							AudioChannelCount: 2,
						},
					},
					LanguageCodes: []string{"ja-JP"},
					Model:         args.SpeechToTextWorker.SpeechToTextModel,
					Features:      &speechpb.RecognitionFeatures{},
				},
				StreamingFeatures: &speechpb.StreamingRecognitionFeatures{
					InterimResults: true,
				},
			},
		},
		Recognizer: args.SpeechToTextWorker.Recognizer,
	}

	recv := make(chan *discordgo.Packet, 2)
	go dgvoice.ReceivePCM(v, recv)

	reader := make(chan string)
	reader_cancel := make(chan struct{})

	go func() {
		done := ""
		for {
			select {
			case <-reader_cancel:
				return
			case text := <-reader:
				if text == done {
					continue
				}

				log.Println("読み上げます", text)
				go PlayText(v, "./data/speech_to_text/", text)
				d.ChannelMessageSend(i.ChannelID, text)
				log.Println("Goroutine Num", runtime.NumGoroutine())
				done = text
			}
		}
	}()

	for {
		stream, err := c.StreamingRecognize(ctx)
		if err != nil {
			log.Println("Error creating streaming recognizer: ", err)
			return
		}

		if err := stream.Send(config_req); err != nil {
			log.Println("Could not sent config: ", err)
		}

		poll_send := make(chan struct{})

		go func() {
			for {
				p := <-recv
				pcm := new(bytes.Buffer)
				binary.Write(pcm, binary.LittleEndian, p.PCM)

				err := stream.Send(&speechpb.StreamingRecognizeRequest{
					StreamingRequest: &speechpb.StreamingRecognizeRequest_Audio{
						Audio: pcm.Bytes(),
					},
					Recognizer: args.SpeechToTextWorker.Recognizer,
				})
				if err == io.EOF {
					log.Println("Send EOF received. Exiting.")
					break
				} else if err != nil {
					log.Println("Error sending audio: ", err)
					break
				}
			}
			poll_send <- struct{}{}
		}()

		go func() {
			for {
				resp, err := stream.Recv()

				if err != nil {
					if err == io.EOF {
						log.Println("Recv EOF received. Exiting.")
						break
					}

					status, ok := status.FromError(err)
					if !ok {
						log.Println("Unknown Error: ", err)
						break
					}
					if status.Code() == codes.Aborted {
						log.Println("Aborted: ", status.Message())
						break
					}
				} else {
					if resp != nil && resp.Results != nil && len(resp.Results) > 0 && len(resp.Results[0].Alternatives) > 0 {
						log.Println(resp.Results[0].Alternatives[0].Transcript)

						reader <- resp.Results[0].Alternatives[0].Transcript
					}
				}
			}
		}()

		log.Println("Waiting...")

		<-poll_send

		log.Println("Polling done.")

		stream.CloseSend()
	}
}
