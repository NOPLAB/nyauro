package nyauro

import (
	"log"
	"os"
	"os/signal"

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
	}
	commandHandlers := map[string]func(d *discordgo.Session, i *discordgo.InteractionCreate){
		"speech-to-text": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			go SpeechToTextWorker(d, i, args)
		},
		"newsletter": func(d *discordgo.Session, i *discordgo.InteractionCreate) {
			go NewsletterWorker(d, i, args)
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
