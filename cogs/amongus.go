package cogs

import (
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
	"github.com/konafx/natalya/util"
	log "github.com/sirupsen/logrus"
)

/*
 * [Create Reaction](https://discord.com/developers/docs/resources/channel#create-reaction)
 * This endpoint requires the 'READ_MESSAGE_HISTORY' permission to be present on the current user.
 * Additionally, if nobody else has reacted to the message using this emoji, this endpoint requires the 'ADD_REACTIONS' permission to be present on the current user. Returns a 204 empty response on success.
 */

const (
	ChannelTypeLobby = iota
	ChannelTypeHeaven
)

const (
	EmojiMeeting	= "📢"
	EmojiMute		= "🤐"
	EmojiFinish		= "🎉"
)

// type Player struct {
// 	voiceState	*discordgo.VoiceState
// 	member		*discordgo.Member
// }


var AmongUs = discordgo.ApplicationCommand{
	Name: "mover",
	Description: "２つのチャンネルを使ってAmong Usの部屋移動をやるヨ！\n",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionChannel,
			Name:			"lobby",
			Description:	"生者のお部屋",
			Required:		true,
		},
		{
			Type:			discordgo.ApplicationCommandOptionChannel,
			Name:			"heaven",
			Description:	"天界",
			Required:		true,
		},
	},
}

func AmongUsHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	length := len(AmongUs.Options)
	chs := make([]*discordgo.Channel, length)
	for j := 0; j < length; j++ {
		ch := i.Data.Options[j].ChannelValue(s)
		if ch.Type != discordgo.ChannelTypeGuildVoice {
			message := fmt.Sprintf("%s だとしゃべれないヨ！", ch.Mention())
			err := util.InteractionErrorResponse(s, *i.Interaction, message)
			if err != nil {
				log.Error(err)
				return
			}
		}
		chs[j] = ch
	}

	log.Debug(chs[0].Name, chs[1].Name)
	embed := new(discordgo.MessageEmbed)
	embed.Title = "Amove Us"
	embed.Description = fmt.Sprintf(`各タイミングで**絵文字を押セ！**
討論開始！→%s
SHHHHHHH!!→%s
Victory or Defeat→%s`,
	EmojiMeeting, EmojiMute, EmojiFinish)
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:	"生者のお部屋",
			Value:	chs[0].Mention(),
		},
		{
			Name:	"天界",
			Value:	chs[1].Mention(),
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Embeds: []*discordgo.MessageEmbed{ embed },
		},
	})
	return
}

func AmongUsMessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Debug("AmongUs")
	if m.Author.ID != s.State.User.ID && len(m.Embeds) == 0 {
		log.Debug("it's not me'")
		return
	}

	if m.Embeds[0].Title != "Amove Us" {
		log.Debug("it's not Amove Us'")
		return
	}

	emojis := []string{EmojiMeeting, EmojiMute, EmojiFinish}
	for _, e := range emojis {
		err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, url.QueryEscape(e))
		if err != nil {
			log.Error(err)
		}
	}
	return
}
// 
// func AmongUsReactionAddHandler(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
// 	m.
// 	var heaven *discordgo.Channel
// 	isDead := func (p *Player) bool {
// 		return p.voiceState.ChannelID == heaven.ID
// 	}
// 
// }

