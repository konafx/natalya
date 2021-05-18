package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var hack Command = &discordgo.ApplicationCommand{
	Name: "hack",
	Description: "hacking this server",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"channel",
			Description:	"Get channel id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:			discordgo.ApplicationCommandOptionChannel,
					Name:			"channel",
					Description:	"The channel to get",
					Required:		true,
				},
			},
		},
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"role",
			Description:	"Get role id",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:			discordgo.ApplicationCommandOptionRole,
					Name:			"role",
					Description:	"The role to get",
					Required:		true,
				},
			},
		},
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"guild",
			Description:	"Get guild id",
		},
	},
}

func hackHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Data.Options[0].Name {
	case "channel":
		ch := i.Data.Options[0].Options[0].ChannelValue(s)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: fmt.Sprintf("%s id is %s", ch.Mention(), ch.ID),
				Flags: 64,
			},
		})
		return
	case "role":
		ro := i.Data.Options[0].Options[0].RoleValue(s, i.GuildID)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: fmt.Sprintf("%s id is %s", ro.Mention(), ro.ID),
				Flags: 64,
			},
		})
		return
	case "guild":
		gID := i.GuildID
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: fmt.Sprintf("this guild id is %s", gID),
				Flags: 64,
			},
		})
		return
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: "LaLaLa...",
				Flags: 64,
			},
		})
		return
	}
}

func init() {
	addCommand(hack, hackHandler)
}
