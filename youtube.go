package main

import (
	"fmt"
	"errors"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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
			Description:	"お気持ち",
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
				Content: fmt.Sprintf("スパチャは %d円から %d円 で行ってください", MinTip, MaxTip),
				Flags: 64,	// set to 64 to make your response ephemeral
			},
		})
		return
	}

	embed.Title = fmt.Sprintf("¥%d", pay)

	var err error
	embed.Color, err = getChatcolor(pay)
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

func getChatcolor(pay int64) (color int, err error) {
	var chatColors = map[int64]string{
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

	var price int64 = 0
	payColor := ""
	for th, colorname := range chatColors {
		if pay > th && price < th {
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

// 
// class YouTube(commands.Cog):
//     def __init__(self, bot: commands.Bot):
//         self.bot = bot
// 
//     @commands.command(
//             usage='<値段> <?コメント>',
//             brief='センキュー・スパチャ♪ ┗(┓卍^o^)卍ﾄﾞｩﾙﾙﾙﾙﾙﾙ↑↑',
//             help='気持ちを伝えるゾ！１００円玉から諭吉サン５枚までで盛り上げるよ♪\n'
//                  'でも、改行は伝えられないミタイ…？\n'
//                  '例: !superchat 2434 かわいい\n'
//                  '例: !superchat 50000\n',
//             aliases=['スパチャ', '投げ銭']
//             )
//     async def superchat(self, ctx: commands.Context, tip: int, *comments):
//         # 円マーク
//         JPY = b'\\xa5'.decode('unicode-escape')
//         embed = discord.Embed(
//                 title=f'{JPY}{tip:,}',
//                 description=' '.join(comments),
//                 color=COLORS[chatcolor(tip)]
//                 )
//         embed.set_author(name=ctx.author.display_name, icon_url=ctx.author.avatar_url_as(
//             format='png',
//             static_format='png'
//             ))
//         await ctx.send(embed=embed)
// 
//     @superchat.error
//     async def superchat_error(self, ctx: commands.Context, error: Exception):
//         print(f'{error=}')
//         if isinstance(error, commands.BadArgument):
//             await ctx.reply('整数しかワカラナイヨ…')
//         elif isinstance(error, commands.MissingRequiredArgument):
//             await ctx.reply(msg.command_usage(ctx))
//         elif isinstance(error, ValueError):
//             await ctx.reply(f'{MIN_TIP}円から{MAX_TIP}円まででお願いシマス♪')
//                 await ctx.send('ミクが「みくは彼氏がいるから１円から５万円までにゃ！」って言ってたヨ？')
// 
// 
// def chatcolor(tip: int, chatcolors: dict[int, str] = CHATCOLORS) -> str:
//     if (tip < MIN_TIP or tip > MAX_TIP):
//         raise ValueError(f'Range Over [{MIN_TIP}, {MAX_TIP}]')
// 
//     allow_chatcolors = list(filter(lambda cc: cc[0] <= tip, chatcolors.items()))
//     color = allow_chatcolors.pop()[1]
//     return color
// 
// 
// def setup(bot: commands.Bot):
//     bot.add_cog(YouTube(bot))
