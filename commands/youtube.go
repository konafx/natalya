package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	MinTip = 100
	MaxTip = 50000
)

var SuperChat = discordgo.ApplicationCommand{
	Name: "superchat",
	Description: "センキュー・スパチャ♪ ┗(┓卍^o^)卍ﾄﾞｩﾙﾙﾙﾙﾙﾙ↑↑",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionInteger,
			Name:			"purchase",
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
	type Purchase struct {
		price	int64
		comment	string
	}
	p := new(Purchase)
	p.price = i.Data.Options[0].IntValue()
	p.comment = i.Data.Options[1].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: fmt.Sprintf("%d : %s", p.price, p.comment),
		},
	})
	return
}

// ChatColors = {
//     100: 'BLUE',
//     200: 'AQUA',
//     500: 'GREEN',
//     1000: 'YELLOW',
//     2000: 'ORANGE',
//     5000: 'MAGENTA',
//     10000: 'RED'
// }
// 
// COLORS = {
//     'BLUE': 0x134A9D,
//     'AQUA': 0x28E4FD,
//     'GREEN': 0x32E8B7,
//     'YELLOW': 0xFCD748,
//     'ORANGE': 0xF37C22,
//     'MAGENTA': 0xE72564,
//     'RED': 0xE32624
// }
// 
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
