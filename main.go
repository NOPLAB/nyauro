package main

import (
	"log"
	"nyauro/nyauro"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Error loading .env file")
		return
	}

	nyauro.Run(nyauro.RunArgument{
		DiscordToken:         os.Getenv("DISCORD_TOKEN"),
		OpenWeatherMapApiKey: os.Getenv("OPEN_WEATHER_MAP_API_KEY"),
		SpeechToTextWorker: nyauro.SpeechToTextWorkerArgument{
			SpeechToTextModel: "short",
			Recognizer:        os.Getenv("SPEECH_TO_TEXT_RECOGNIZER"),
		},
	})
}
