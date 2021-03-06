package main

import (
	"github.com/bwmarrin/discordgo"
)

var Hello Command = &discordgo.ApplicationCommand{
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

func init() {
	addCommand(Hello, HelloHandler)
}
