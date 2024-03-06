package nyauro

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

type RunArgument struct {
	DiscordToken         string
	OpenWeatherMapApiKey string
	SpeechToTextWorker   SpeechToTextWorkerArgument
}

func Run(args RunArgument) error {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "speech-to-text",
			Description: "ボイスチャットで話したことをテキストに起こすよ",
		},
		{
			Name:        "newsletter",
			Description: "時報をボイスチャットで話すよ",
		},
		{
			Name:        "clock",
			Description: "時報をボイスチャットで話すよ",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "enable",
					Description: "時報を切り替えるよ",
					Required:    true,
				},
			},
		},
	}
	commandHandlers := map[string]func(d *discordgo.Session, i *discordgo.InteractionCreate){
		"speech-to-text": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			go SpeechToTextWorker(d, i, args)
		},
		"newsletter": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			go NewsletterWorker(d, i, args)
		},
		"clock": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			go ClockWorker(d, i, args)
		},
	}

	d, err := discordgo.New("Bot " + args.DiscordToken)
	if err != nil {
		return err
	}

	d.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	d.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})

	cancel_map := make(map[string]context.CancelFunc)
	defer func() {
		for _, cancel := range cancel_map {
			cancel()
		}
	}()
	d.AddHandler(func(s *discordgo.Session, i *discordgo.VoiceStateUpdate) {
		if (i.BeforeUpdate == nil || i.BeforeUpdate.ChannelID == "") && i.ChannelID != "" {
			log.Println("Join user: ", i.Member.User.Username)

			ctx, cancel := context.WithCancel(context.Background())
			cancel_map[i.UserID] = cancel
			go JoinTimerWorker(s, i.UserID, i.GuildID, 30*time.Minute, ctx)
		}
		if i.BeforeUpdate != nil && i.BeforeUpdate.ChannelID != "" && i.ChannelID == "" {
			log.Println("leave user: ", i.Member.User.Username)

			cancel, ok := cancel_map[i.UserID]
			if ok {
				log.Println("cancel: ", i.Member.User.Username)
				cancel()
			}
		}
	})

	d.Identify.Intents = discordgo.IntentsAll

	err = d.Open()
	if err != nil {
		return err
	}

	registerCommands := make(map[string][]*discordgo.ApplicationCommand)

	for _, g := range d.State.Guilds {
		registerCommands[g.ID] = make([]*discordgo.ApplicationCommand, len(commands))
		for j, v := range commands {
			cmd, err := d.ApplicationCommandCreate(d.State.User.ID, g.ID, v)
			if err != nil {
				return err
			}

			registerCommands[g.ID][j] = cmd
		}
	}

	defer d.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	return nil
}
