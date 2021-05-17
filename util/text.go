package util

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func ToUser(id string) string {
	return fmt.Sprintf("<@%s>", id)
}

func ToChannel(id string) string {
	return fmt.Sprintf("<#%s>", id)
}

func ToRole(id string) string {
	return fmt.Sprintf("<@&%s>", id)
}

func MakeEmbedField(name string, values ...string) *discordgo.MessageEmbedField {
	field := discordgo.MessageEmbedField{}
	field.Name = name
	field.Value = strings.Join(values, "\n")
	return &field
}
