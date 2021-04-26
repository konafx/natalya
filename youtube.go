package main

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/width"
)

var (
	MinTip int64 = 100
	MaxTip int64 = 50000
)

var SuperChat = discordgo.ApplicationCommand{
	Name: "superchat",
	Description: "センキュー・スパチャ♪ ┗(┓卍^o^)卍ﾄﾞｩﾙﾙﾙﾙﾙﾙ↑↑",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionInteger,
			Name:			"pay",
			Description:	fmt.Sprintf("お気持ち [%d,%d]", MinTip, MaxTip),
			Required:		true,
		},
		{
			Type:			discordgo.ApplicationCommandOptionString,
			Name:			"comment",
			Description:	"コメント",
			Required:		false,
		},
	},
}

func SuperChatHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := new(discordgo.MessageEmbed)

	embed.Author = new(discordgo.MessageEmbedAuthor)
	embed.Author.Name = i.Member.Nick
	if len(embed.Author.Name) == 0 {
		embed.Author.Name = i.Member.User.Username
	}
	embed.Author.IconURL = i.Member.User.AvatarURL("png")

	pay := i.Data.Options[0].IntValue()
	if pay < MinTip || pay > MaxTip { 
		log.Debug("範囲外のチップ")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: fmt.Sprintf("ミクが「みくは彼氏がいるから%s円から%s円までにゃ！」って言ってたヨ？", uIntToZenkakuOkuman(uint(MinTip)),uIntToZenkakuOkuman(uint(MaxTip))),
				Flags: 64,	// set to 64 to make your response ephemeral
			},
		})
		return
	}

	// カンマを入れるために NewPrinter
	p := message.NewPrinter(language.Japanese)
	embed.Title = p.Sprintf("¥%d", pay)

	var err error
	embed.Color, err = getChatcolor(int(pay))
	if err != nil {
		log.Error(err)
	}

	if len(i.Data.Options) >= 2 {
		comment := i.Data.Options[1].StringValue()
		embed.Description = comment
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Embeds: []*discordgo.MessageEmbed{
				embed,
			},
		},
	})
	return
}

func uIntToZenkakuOkuman(num uint) (okuman string) {
	var (
		// Kansuji = []string{"零", "一", "二", "三", "四", "五", "六", "七", "八", "九"}
		// Suffix = []string{"千", "百", "十", ""}
		Delimiter = []string{"", "万", "億", "兆", "京", "垓"}
	)

	if num == 0 {
		okuman = "零"
		return okuman
	}
	for i := 0; num > 0; i, num = i+1, num/10000 {
		fours := num % 10000
		log.Debugf("i: %d, fours: %d, num: %d", i, fours, num)
		if fours == 0 { continue }

		han := fmt.Sprintf("%d", fours)
		zen := width.Widen.String(han)

		delimiter := Delimiter[i]
		okuman = fmt.Sprintf("%s%s%s", zen, delimiter, okuman)
		log.Debugf("okuman: %s, 残り: %d", okuman, num)
	}

	return okuman
}

func getChatcolor(pay int) (color int, err error) {
	var chatColors = map[int]string{
		100: "Blue",
		200: "Aqua",
		500: "Green",
		1000: "Yellow",
		2000: "Orange",
		5000: "Magenta",
		10000: "Red",
	}
	var colors = map[string]int{
		"Blue": 0x134A9D,
		"Aqua": 0x28E4FD,
		"Green": 0x32E8B7,
		"Yellow": 0xFCD748,
		"Orange": 0xF37C22,
		"Magenta": 0xE72564,
		"Red": 0xE32624,
	}

	var price int = 0
	payColor := ""
	for th, colorname := range chatColors {
		if pay >= th && price < th {
			price = th
			payColor = colorname
		}
	}
	if payColor == "" {
		return 0, errors.New("your pay is less than 100 Yen")
	}

	color, ok := colors[payColor]
	if !ok {
		return 0, errors.New("colorcode not found :" + payColor)
	}

	return color, nil
}
