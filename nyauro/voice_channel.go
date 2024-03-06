package nyauro

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func JoinVoiceChannelWithInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.VoiceConnection, error) {
	state, err := s.State.VoiceState(i.GuildID, i.Member.User.ID)

	if err != nil {
		log.Println("Error getting voice state: ", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "ボイスチャンネルに入室していません!",
			},
		})
		return nil, err
	}

	v, err := s.ChannelVoiceJoin(i.GuildID, state.ChannelID, false, false)
	if err != nil {
		log.Println("Error joining voice channel: ", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "入室中にエラーが発生しました!",
			},
		})
		return nil, err
	}

	return v, err
}

func JoinVoiceChannelWithIds(s *discordgo.Session, guildId string, userId string) (*discordgo.VoiceConnection, error) {
	state, err := s.State.VoiceState(guildId, userId)

	if err != nil {
		log.Println("Error getting voice state: ", err)
		return nil, err
	}

	v, err := s.ChannelVoiceJoin(guildId, state.ChannelID, false, false)
	if err != nil {
		return nil, err
	}

	return v, err
}
