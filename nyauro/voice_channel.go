package nyauro

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func JoinVoiceChannel(d *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.VoiceConnection, error) {
	state, err := d.State.VoiceState(i.GuildID, i.Member.User.ID)

	if err != nil {
		log.Println("Error getting voice state: ", err)
		d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "ボイスチャンネルに入室していません!",
			},
		})
		return nil, err
	}

	v, err := d.ChannelVoiceJoin(i.GuildID, state.ChannelID, false, false)
	if err != nil {
		log.Println("Error joining voice channel: ", err)
		d.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "入室中にエラーが発生しました!",
			},
		})
		return nil, err
	}

	return v, err
}
