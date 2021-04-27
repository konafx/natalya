package doc

import (
	"github.com/bwmarrin/discordgo"
)

var Command = discordgo.ApplicationCommand{
	Name: "name",
	Description: "description",
}

func CommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: "Response message",
		},
	})
	return
}
