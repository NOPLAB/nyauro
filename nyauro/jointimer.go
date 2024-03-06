package nyauro

import (
	"context"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func JoinTimerWorker(s *discordgo.Session, userId string, guildId string, duration time.Duration, ctx context.Context) {
	d := time.Duration(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(duration):
		}

		v, err := JoinVoiceChannelWithIds(s, guildId, userId)
		if err != nil {
			return
		}

		d += duration

		text := "入室してから" + fmt.Sprintf("%.0f", d.Minutes()) + "分たちました！"
		PlayText(v, "./data/join_timer/", text)
	}
}
