package main

import (
	"fmt"
	"net/url"
	"regexp"

	"golang.org/x/sync/errgroup"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	u "github.com/konafx/natalya/util"
)

/*
 * [Create Reaction](https://discord.com/developers/docs/resources/channel#create-reaction)
 * This endpoint requires the 'READ_MESSAGE_HISTORY' permission to be present on the current user.
 * Additionally, if nobody else has reacted to the message using this emoji, this endpoint requires the 'ADD_REACTIONS' permission to be present on the current user. Returns a 204 empty response on success.
 */

/*
 * スラッシュコマンドから embed メッセージを送信する
 * この機能の embed メッセージに絵文字をつける
 * ユーザーが絵文字を押したら、embed メッセージに記載のある
 * ボイスチャンネルを使って、参加ユーザーの音声状態を変更する
 */

const (
	ChannelTypeLobby = iota
	ChannelTypeOnBoard
	ChannelTypeHeaven
)

// TODO: 設定できると面白いか？
const (
	EmojiMeeting	= "📢"
	EmojiMute		= "🤐"
	EmojiFinish		= "🎉"
)

var AmongUs Command  = &discordgo.ApplicationCommand{
	Name: "mover",
	Description: "３つのチャンネルを使ってAmong Usの部屋移動をやるヨ！\n",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionChannel,
			Name:			"lobby",
			Description:	"ロビー",
			Required:		true,
		},
		{
			Type:			discordgo.ApplicationCommandOptionChannel,
			Name:			"onboard",
			Description:	"船内",
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
	g, err := s.State.Guild(i.GuildID)
	if err != nil {
		log.Error(err)
		return
	}

	// チャンネル取得
	length := len(AmongUs.Options)
	chs := make([]*discordgo.Channel, length)
	for j := 0; j < length; j++ {
		ch := i.Data.Options[j].ChannelValue(s)
		var message string
		switch {
		case j != ChannelTypeOnBoard && ch.ID == g.AfkChannelID:
			if message == "" {
				message = fmt.Sprintf("%s は AFKチャンネルだから%s にできないんダ～♪ｽﾔｽﾔ～zzz", ch.Mention(), AmongUs.Options[ChannelTypeOnBoard].Description)
			}
			fallthrough
		case ch.Type != discordgo.ChannelTypeGuildVoice:
			if message == "" {
				message = fmt.Sprintf("%s だとしゃべれないヨ！", ch.Mention())
			}
			err := u.InteractionErrorResponse(s, i.Interaction, message)
			if err != nil { log.Error(err) }
			return
		}
		chs[j] = ch
	}
	log.Debug(chs[0].Name, chs[1].Name, chs[2].Name)

	// embed message 作成
	embed := new(discordgo.MessageEmbed)
	embed.Title = "Amove Us"
	embed.Description = fmt.Sprintf(`各タイミングで**絵文字を押セ！**
討論開始！→%s
SHHHHHHH!!→%s
Victory or Defeat→%s`,
	EmojiMeeting, EmojiMute, EmojiFinish)
	for j := 0; j < length; j++ {
		field := discordgo.MessageEmbedField{
			Name:	AmongUs.Options[j].Description,
			Value:	chs[j].Mention(),
		}
		embed.Fields = append(embed.Fields, &field)
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
	if m.Author.ID != s.State.User.ID || len(m.Embeds) == 0 {
		return
	}

	if m.Embeds[0].Title != "Amove Us" {
		return
	}

	log.Debug("AmoveUs starts to react")
	emojis := []string{EmojiMeeting, EmojiMute, EmojiFinish}
	for _, v := range emojis {
		err := s.MessageReactionAdd(m.ChannelID, m.Message.ID, url.QueryEscape(v))
		if err != nil {
			log.Error(err)
		}
	}
	return
}

func AmongUsReactionAddHandler(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if r.UserID == s.State.User.ID {
		return
	}

	m, err := s.ChannelMessage(r.ChannelID, r.MessageID)
	if err != nil {
		log.Errorf("Cannot message: %v", err)
		return
	}

	if len(m.Embeds) == 0 || m.Embeds[0].Title != "Amove Us" {
		return
	}

	// チャンネル取得
	var chs [3]*discordgo.Channel
	// TODO: スライスの長さと Fields の長さが不一致なことはあるだろうか。
	for i, v := range m.Embeds[0].Fields {
		// matches: ["<#123456789>", "123456789"]
		matches := regexp.MustCompile(`<#(\d+)>`).FindStringSubmatch(v.Value)
		if len(matches) == 0 {
			return
		}
		ch, err := s.Channel(matches[1])
		if err != nil {
			log.Errorf("%s was not found: %v", v.Value, err)
			return
		}
		if ch.Type != discordgo.ChannelTypeGuildVoice {
			log.Errorf("%s is not VoiceChannel", ch.Name)
			return
		}

		chs[i] = ch
	}

	g, _ := s.State.Guild(r.GuildID)
	isAfk := chs[ChannelTypeOnBoard].ID == g.AfkChannelID

	// Mover 本体
	var eg errgroup.Group
	for _, vs := range g.VoiceStates {
		// TODO: このコメントを消す https://qiita.com/koduki/items/55c277efe8c4ee77910b
		log.Debugf("%#v", vs)
		session := *s
		guildID := g.ID
		userID  := vs.UserID
		switch r.Emoji.Name {
		case EmojiMeeting:
			log.Debug("Meeting")
			switch vs.ChannelID {
			case chs[ChannelTypeOnBoard].ID:
				eg.Go(func () error {
					return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamMute(false), u.ModifyVSParamChannelID(chs[ChannelTypeLobby].ID))
				})
			case chs[ChannelTypeHeaven].ID:
				eg.Go(func () error {
					return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamMute(true), u.ModifyVSParamChannelID(chs[ChannelTypeLobby].ID))
				})
			}
		case EmojiMute:
			log.Debug("SHIIIIIIIIIIII")
			if vs.ChannelID != chs[ChannelTypeLobby].ID {
				continue
			}
			switch vs.Mute || vs.SelfMute {
			case false:
				eg.Go(func () error {
					if isAfk {
						return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamChannelID(chs[ChannelTypeOnBoard].ID))
					}
					return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamMute(true), u.ModifyVSParamChannelID(chs[ChannelTypeOnBoard].ID))
				})
			case true:
				eg.Go(func () error {
					return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamMute(false), u.ModifyVSParamChannelID(chs[ChannelTypeHeaven].ID))
				})
			}
		case EmojiFinish:
			log.Debug("End")
			switch vs.ChannelID {
			case chs[ChannelTypeLobby].ID, chs[ChannelTypeOnBoard].ID, chs[ChannelTypeHeaven].ID:
				eg.Go(func () error {
					return u.RequestModifyVS(&session, guildID, userID, u.ModifyVSParamMute(false), u.ModifyVSParamChannelID(chs[ChannelTypeLobby].ID))
				})
			}
		default:
			log.Debug("Nothing to do")
		}
	}

	eg.Go(func () error { return removeAddedReaction(s, r) })

	if err := eg.Wait(); err != nil {
		log.Error(err)
	}
	return
}

func removeAddedReaction(s *discordgo.Session, r *discordgo.MessageReactionAdd) error {
	emojiID := r.Emoji.ID
	if emojiID == "" {
		emojiID = url.QueryEscape(r.Emoji.Name)
	}

	err := s.MessageReactionRemove(r.ChannelID, r.MessageID, emojiID, r.UserID)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func init() {
	log.Debugf("%v, %v", AmongUs, AmongUsHandler)
	addCommand(AmongUs, AmongUsHandler)
	addHandler(AmongUsReactionAddHandler, AmongUsMessageCreateHandler)
}
