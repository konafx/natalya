package cogs

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	// "google.golang.org/api/iterator"
	"cloud.google.com/go/firestore"
	u "github.com/konafx/natalya/util"
)

type Poet struct {
	Number	int			`json:"number" firestore:"number"`
	UserID	string		`json:"userId" firestore:"userId"`
	Poem	string		`json:"poem" firestore:"poem"`
	User	*discordgo.User	`json:"-" firestore:"-"`
}

type UnknownPoetGame struct {
	GuildID	string		`json:"guildId" firestore:"guildId"`
}

var Haiku = discordgo.ApplicationCommand{
	Name: "haiku",
	Description: "俳句ゲームだヨ♪\n",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet1",
			Description:	"詠み人",
			Required:		true,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet2",
			Description:	"詠み人",
			Required:		true,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet3",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet4",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet5",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet6",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet7",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet8",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet9",
			Description:	"詠み人",
			Required:		false,
		},
		{
			Type:			discordgo.ApplicationCommandOptionUser,
			Name:			"poet10",
			Description:	"詠み人",
			Required:		false,
		},
	},
}

func HaikuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var poets []Poet
	for k, v := range i.Data.Options {
		u := v.UserValue(s)
		poet := Poet{
			Number:	k,
			UserID:	u.ID,
			User:	u,
		}
		poets = append(poets, poet)
	}
	log.Debugln(poets)

	message := "今回の俳人はお前らダ！よろしくナ♪\n"
	for k, v := range poets {
		if k == 0 {
			message = fmt.Sprintf("%s%s", message, v.User.Mention())
			continue
		}
		message = fmt.Sprintf("%s, %s", message, v.User.Mention())
	}
	log.Debug(message)

	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	{
		game := UnknownPoetGame{
			GuildID:	i.GuildID,
		}
		doc, _, err := client.Collection("unknownPoetGames").Add(ctx, game)
		if err != nil {
			u.InteractionErrorResponse(s, *i.Interaction, "ウーン、ここは俳句を詠むにはうるさすぎるみたイ…")
			log.Error(err)
			return
		}
		for _, v := range poets {
			if _, _, err := doc.Collection("poets").Add(ctx, v); err != nil {
				u.InteractionErrorResponse(s, *i.Interaction, "ウーン、俳句を詠む心が備わってないないみたイ…")
				log.Error(err)
				return
			}
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: message,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{
					discordgo.AllowedMentionTypeUsers,
				},
			},
		},
	})
	return
}

func UnknownPoetDMHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		return
	}
	// m.Author.ID
	return
}

// db
func createClient(ctx context.Context) *firestore.Client {
	// Sets your Google Cloud Platform project ID.
	projectID := "natalya"

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	// Close client when done with
	// defer client.Close()
	return client
}
