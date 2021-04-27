package cogs

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/konafx/natalya/util"
	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
)

const (
	ChannelTypeLobby = iota
	ChannelTypeHeaven
)

var (
	EmojiMeeting	= emoji.Sprintf(":loudspeaker:")
	EmojiMute		= emoji.Sprintf(":zipper-mouth_face:")
	EmojiFinish		= emoji.Sprintf(":party_popper:")
)

// type Player struct {
// 	voiceState	*discordgo.VoiceState
// 	member		*discordgo.Member
// }

// func AmongUsReactionAddHandler(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
// 	var heaven *discordgo.Channel
// 	isDead := func (p *Player) bool {
// 		return p.voiceState.ChannelID == heaven.ID
// 	}
// 
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
		log.Debug(e)
		err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, e)
		if err != nil {
			log.Error(err)
		}
	}
	return
}
