package nyauro

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

type VoicevoxV3Response struct {
	Success         bool   `json:"success"`
	IsApiKeyValid   bool   `json:"isApiKeyValid"`
	SpeakerName     string `json:"speakerName"`
	AudioStatusUrl  string `json:"audioStatusUrl"`
	WavDownloadUrl  string `json:"wavDownloadUrl"`
	Mp3DownloadUrl  string `json:"mp3DownloadUrl"`
	Mp3StreamingUrl string `json:"mp3StreamingUrl"`
}

type VoicevoxAudioFileResponse struct {
	Success      bool   `json:"success"`
	IsAudioReady bool   `json:"isAudioReady"`
	IsAudioError bool   `json:"isAudioError"`
	Status       string `json:"status"`
	UpdatedTime  int    `json:"updatedTime"`
}

func PlayText(v *discordgo.VoiceConnection, cache_dir, text string) {
	path := FileBase64(cache_dir, text)

	if !ExistFile(path) {
		log.Println("Not ExistFile: ", path)
		if err := MkVoice(path, text); err != nil {
			log.Println("Error MkVoice: ", err)
		}
	} else {
		log.Println("ExistFile: ", path)
	}

	dgvoice.PlayAudioFile(v, path, make(chan bool))
}

func ExistFile(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func FileBase64(path, text string) string {
	base64 := base64.URLEncoding.EncodeToString([]byte(text))
	s := strings.Split(path, "/")
	if s[len(s)-1] == "" {
		return path + base64 + ".mp3"
	}
	return path + "/" + base64 + ".mp3"
}

func MkVoice(path, text string) error {
	if text == "" {
		return fmt.Errorf("text is empty")
	}

	text = strings.ReplaceAll(text, " ", "")

	url := "https://api.tts.quest/v3/voicevox/synthesis" + "?text=" + text

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var v3_res VoicevoxV3Response

	if err := json.Unmarshal(b, &v3_res); err != nil {
		return err
	}

	t := time.NewTicker(1 * time.Second)
	defer t.Stop()

	for {
		<-t.C
		resp, err := http.Get(v3_res.AudioStatusUrl)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var status_res VoicevoxAudioFileResponse

		if err := json.Unmarshal(b, &status_res); err != nil {
			return err
		}

		log.Println("StatusResponse: ", status_res)

		if status_res.IsAudioReady {
			break
		}
	}

	log.Println("Start downloading...")

	if err := DownloadFile(path, v3_res.Mp3DownloadUrl); err != nil {
		return err
	}
	log.Println("Downloaded.")

	return nil
}

func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	s := strings.Split(filepath, "/")
	dir_path := strings.Join(s[:len(s)-1], "/")

	if err := os.MkdirAll(dir_path, 0777); err != nil {
		return err
	}

	os.Remove(filepath)

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
