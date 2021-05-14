package util

import (
	"github.com/bwmarrin/discordgo"
)

func InteractionErrorResponse(s *discordgo.Session, i *discordgo.Interaction, message string) error {
	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: message,
			Flags: 64,	// set to 64 to make your response ephemeral
		},
	})
	if err != nil {
		return err
	}
	return nil
}
