package cogs

import (
	"github.com/bwmarrin/discordgo"
)

var Hello = discordgo.ApplicationCommand{
	Name: "hello",
	Description: "Hello command",
}

func HelloHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: "Bon tarde!",
		},
	})
	return
}
