package nyauro

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/bwmarrin/discordgo"
)

var clock map[string]context.CancelFunc

func ClockWorker(s *discordgo.Session, i *discordgo.InteractionCreate, args RunArgument) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})

	options := i.ApplicationCommandData().Options

	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	enable := optionMap["enable"].BoolValue()

	if clock == nil {
		clock = make(map[string]context.CancelFunc)
	}

	if enable {
		_, ok := clock[i.GuildID]
		if !ok {
			ctx, cancel := context.WithCancel(context.Background())
			go Clock(s, i.GuildID, i.Member.User.ID, ctx, args)
			clock[i.GuildID] = cancel
		}
	} else {
		c, ok := clock[i.GuildID]
		if ok {
			c()
		}
		clock[i.GuildID] = nil
	}
}

func Clock(s *discordgo.Session, guildId string, userId string, ctx context.Context, args RunArgument) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()

	dir := "./data/clock/"

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-t.C:
			if now.Minute() == 0 {
				v, err := JoinVoiceChannelWithIds(s, guildId, userId)
				if err != nil {
					return
				}

				text := time.Now().Format("15時くらいです!")
				PlayText(v, dir, text)

				values := url.Values{}
				values.Add("q", "Tokyo")
				values.Add("APPID", args.OpenWeatherMapApiKey)

				req, err := http.NewRequest("GET", "https://api.openweathermap.org/data/2.5/weather", nil)
				if err != nil {
				}

				req.URL.RawQuery = values.Encode()
				client := &http.Client{}
				res, err := client.Do(req)
				if err != nil {
				}

				defer res.Body.Close()

				w_res := Response{}
				err = json.NewDecoder(res.Body).Decode(&w_res)
				if err != nil {
					panic(err)
				}

				text = "現在の気温は体感、" + fmt.Sprintf("%.1f", w_res.Main.Temp-273.15) + "度くらいな気がします"
				PlayText(v, dir, text)

				if w_res.Main.Temp-273.15 > 30 {
					PlayText(v, dir, "暑いです、アイスが食べたいです")
				} else if w_res.Main.Temp-273.15 < 10 {
					PlayText(v, dir, "寒いです、お布団に入りたいです")
				} else if w_res.Main.Temp-273.15 < 0 {
					PlayText(v, dir, "雪降らないかなです、雪だるま作りたいです")
				}
			}
		}
	}
}
