package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"sync"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
	"github.com/glassonion1/xgo"
	"github.com/konafx/natalya/repository"
	"github.com/konafx/natalya/service"
	u "github.com/konafx/natalya/util"
	"github.com/mattn/go-pubsub"
	log "github.com/sirupsen/logrus"
)

var ps *pubsub.PubSub
type StageFinishEvent struct {
	gameRefID	string
	stage	uint
}
type PoemUpdateEvent struct {
	gameRefID	string
	poetID	string
}
type PoetLeaveEvent struct {
	gameRefID	string
	poetID	string
}

var haikuRepo *repository.HaikuRepository

var haiku Command = &discordgo.ApplicationCommand{
	Name: "haiku",
	Description:	"「詠み人知らず」１音ずつ詠んで、みんなで俳句を作るゲームなんダ♪",
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"はじまり",
			Description:	"ゲーム開始だゾ♪詠み人を指名してネ（できればプロデューサー自身も指名してほしいナ…）",
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
		},
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"筆を置く",
			Description:	"中断するゾ…",
		},
		{
			Type:			discordgo.ApplicationCommandOptionSubCommand,
			Name:			"help",
			Description:	"説明書",
		},
	},
}

func haikuHelper(s *discordgo.Session) *discordgo.MessageEmbed {
	help := discordgo.MessageEmbed{
		Title: "詠み人知らず",
		Fields: []*discordgo.MessageEmbedField{
			u.MakeEmbedField(
				"遊び方",
				"各参加者は空っぽの１句を最初にもらう",
				fmt.Sprintf("最初の一字（「じゃ」など可）を決めて、%sにDMで送信", s.State.User.Mention()),
				"みんな一音決めたら、次の人にその１句が渡されます",
				"DMでもらった句に続く一字を決めて送信する",
				"この作業を１７回繰り返して、みんなの一字でキメラの１句を人数分作っていく"),
			u.MakeEmbedField(
				fmt.Sprintf("コマンド：%s", haiku.Options[0].Name),
				"詠み人（参加者）を２～１０人指名してください"),
			u.MakeEmbedField(
				fmt.Sprintf("コマンド: %s", haiku.Options[1].Name),
				"参加している句会（このゲーム）を中断できます",
				"現段階で作られた途中までの句が確認できます"),
			u.MakeEmbedField(
				"その他ルール",
				"・送信する一字はひらがな１文字、または「じゃ」などの拗音一個でおねがいします",
				"・みんなの一字が確定するまで何度も一字を送信することでやり直しできます",
				"・バグったらすまん")},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "ナターリアbot - 製作者：inari#5104",
			IconURL: s.State.User.AvatarURL("png"),
		},
	}
	return &help
}

func haikuHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()

	switch i.Data.Options[0].Name {
	case "help":
		embed := haikuHelper(s)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Embeds: []*discordgo.MessageEmbed{embed},
			},
		})
		return
	case "筆を置く":
		log.Debug("get gamesnap")
		var userID string
		if i.User != nil {
			userID = i.User.ID
		} else {
			userID = i.Member.User.ID
		}
		gamesnap, err := getGamesnapByUserID(ctx, client, userID)
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("開かれてる句会に参加してないみたイ…"))
			return
		}

		log.Debug("update game status to suspend")
		if _, err := gamesnap.Ref.Update(ctx, []firestore.Update{
			{
				Path: "status",
				Value: GameStatusSuspend,
			},
		}); err != nil {
			u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("中断できなかったヨ…"))
			return
		}

		log.Debug("get poets")
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("詠み人が取得できなかったヨ…"))
			return
		}

		log.Debug("get poems")
		poems, err := getPoems(ctx, gamesnap.Ref)
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("俳句が取得できなかったヨ…"))
			return
		}

		log.Debug("make result")
		result := "ここまでつくってたヨ\n"
		for i, l := 0, len(poems); i<l; i++ {
			result = fmt.Sprintf("%s\n%s %s", result, u.ToUser(poets[i].UserID), poems[i].formatHaiku())
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionApplicationCommandResponseData{
				Content: result,
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: []discordgo.AllowedMentionType{
						discordgo.AllowedMentionTypeUsers,
					},
				},
			},
		})

		return
	case "はじまり":
		// CREATE poets
		var poets []*repository.Poet
		for k, v := range i.Data.Options[0].Options {
			user := v.UserValue(s)
			if user.Bot {
				u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("Bot %s に俳句は詠めないヨ", user.Mention()))
				return
			}
			poet := repository.Poet{
				ID:	user.ID,
				NextPoemNumber: uint(k),
			}
			poets = append(poets, &poet)
		}
		log.Debugln(poets)

		// 重複チェック
		{
			m := make(map[string]struct{})
			for _, v := range poets {
				if _, ok := m[v.ID]; ok {
					u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("%s が二回指名されてるゾ", u.ToUser(v.ID)))
					return
				}
				m[v.ID] = struct{}{}
			}
		}

		// 順番ランダム
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(poets), func(i, j int) { poets[i], poets[j] = poets[j], poets[i] })

		for _, v := range poets {
			exist, _ := haikuRepo.GetGameAndPoetByPoetID(v.ID)
			if exist != nil {
				u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("%sが別の句会に参加しているみたイ", u.ToUser(v.ID)))
				log.Debugf("user:%s is composing another poem", v.ID)
				return
			}
		}

		game := repository.NewHaikuGame(poets)
		err := haikuRepo.StoreGame(ctx, game)
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, "ウーン、句会が開けないみたイ…")
			log.Error(err)
			return
		}

		message := "今回の俳人はお前らダ！よろしくナ♪\n"
		for k, v := range poets {
			if k == 0 {
				message = fmt.Sprintf("%s%s", message, u.ToUser(v.ID))
				continue
			}
			message = fmt.Sprintf("%s, %s", message, u.ToUser(v.ID))
		}
		log.Debug(message)

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

		log.Debugln("game start")

		game.Status = repository.GameStatusPlaying
		if err := haikuRepo.UpdateGame(ctx, game); err != nil {
			log.Error(err)
		}

		// ゲーム開始
		{
			c := make(chan struct{})
			watch := make(chan bool)

			log.Debugln("Start to wait writing poems")
			go waitFinishWritingPoems(s, ctx, game, c)
			go watchSuspendedGame(ctx, gamedoc, watch)
			select {
			case _, ok := <-c:
				if !ok {
					s.ChannelMessageSend(i.ChannelID, "なんか起きたので中断されました")
					if _, err := gamedoc.Update(ctx, []firestore.Update{
						{
							Path: "status",
							Value: GameStatusSuspend,
						},
					}); err != nil {
						log.Error(err)
					}
					return
				}
			case v := <-watch:
				if v {
					log.Infof("Suspended game while waiting to finish writing poems")
					return
				}
			}
			log.Debugln("End to wait writing poems")
		}

		poems, _ := getPoems(ctx, gamedoc)

		result := "俳句ができたヨ\n"
		for i, l := 0, len(poems); i<l; i++ {
			result = fmt.Sprintf("%s\n%s %s", result, u.ToUser(poets[i].UserID), poems[i].formatHaiku())
		}
		log.Debug(result)

		if _, err := gamedoc.Update(ctx, []firestore.Update{
			{
				Path: "status",
				Value: GamestatusGraceful,
			},
		}); err != nil {
			log.Error(err)
		}

		s.ChannelMessageSend(i.ChannelID, result)

		return
	default:
		log.Debug(i.Data)
	}

}

func waitFinishWritingPoems(s *discordgo.Session, ctx context.Context, game *repository.HaikuGame, c chan struct{}) {
	defer close(c)

	// 各ステージの終了を検知する
	for {
		poets := game.Poets

		// DM を送る
		for _, v := range poets {
			if err := sendDraftPoem(s, ctx, game, v); err != nil {
				log.Error(err)
				return
			}
		}

		c2 := make(chan int)
		log.Debugln("Start to wait finish stage")
		go waitFinishStage(ctx, game, c2)
		stage, ok := <-c2
		log.Debugln("End to wait finish stage")
		if !ok {
			log.Error("something happened on waitFinishStage")
			return
		}

		{
			it := gamedoc.Collection("poets").Documents(ctx)
			for {
				snap, err := it.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					log.Error(err)
					return
				}
				// TODO: goroutinize
				if _, err := snap.Ref.Update(ctx, []firestore.Update{
					{
						Path: "next",
						Value: firestore.Increment(1),
					},
				}); err != nil {
					log.Error(err)
					return
				}
			}
		}

		stage = stage + 1
		_, err = gamedoc.Update(ctx, []firestore.Update{
			{
				Path: "stage",
				Value: stage,
			},
		})
		if err != nil {
			log.Error(err)
			return
		}

		// stage = [1, 17]
		if stage > 17 {
			c <- struct{}{}
			return
		}
	}
}

// 各ステージの終了を待つ
func waitFinishStage(ctx context.Context, game *repository.HaikuGame, c chan uint) {
	defer close(c)

	// タスク管理
	var wg sync.WaitGroup
	wg.Add(len(game.Poets))

	// 時限管理
	ctxTimeout, cancelTimeout := context.WithTimeout(ctx, 120*time.Second)
	defer cancelTimeout()

	// composeds はpoet.IDをキーにしたチャンネルのマップ
	composeds := make(map[string](chan struct{}))
	for _, v := range game.Poets {
		composeds[v.ID] = make(chan struct{})

		// 時限までのタスク消化
		go func (ctx context.Context, poetID string) {
			select {
			case <- composeds[poetID]:
				wg.Done()
			case <- ctx.Done():
				log.Infof("Timeup waitFinishStage on Game(%v)", game.RefID)
				// timeout!
			}
		}(ctxTimeout, v.ID)
	}

	// イベントをチャンネル完了へ変換
	ps.Sub(func (e *PoemUpdateEvent) {
		composeds[e.poetID] <- struct{}{}
	})

	wg.Wait()
	c <- game.Stage
	// 処理完了
}

// 各ユーザーにDMおくって放置
func sendDraftPoem(s *discordgo.Session, ctx context.Context, game *repository.HaikuGame, poet *repository.Poet) error {
	// send dm
	ch, err := s.UserChannelCreate(poet.ID)
	if err != nil {
		return err
	}

	poem := haikuRepo.GetNextPoem(poet, game)

	var message string
	if len(poem.PoemRunes) == 0 {
		message = "最初の一文字を決めてほしいナ♪"
	} else {
		message = fmt.Sprintf("次の一文字を入力してネ！\n> %s", poem.FormatHaiku())
	}

	if _, err = s.ChannelMessageSend(ch.ID, message); err != nil {
		return err
	}

	return nil
}

func updatePoem(content string, poet *repository.Poet, game *repository.HaikuGame) string {
	poem := haikuRepo.GetNextPoem(poet, game)

	poemRune := repository.NewPoemRune(poet.ID, content)
	poem.PoemRunes = append(poem.PoemRunes[:game.Stage-1], poemRune)

	ctx := context.Background()
	haikuRepo.UpdateGame(ctx, game)

	ps.Pub(&PoemUpdateEvent{game.RefID, poet.ID})

	return poem.FormatHaiku()
}

func dmHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		return
	}
	if s.State.User.ID == m.Author.ID {
		return
	}

	game, poet := haikuRepo.GetGameAndPoetByPoetID(m.Author.ID)
	if game == nil {
		return
	}

	if game.Status != repository.GameStatusPlaying {
		return
	}
	// ここまでで参加者であることがおそらく保証されている

	{
		handler := func() {
			s.ChannelMessageSendReply(m.ChannelID, "ひらがな１文字か「しゃ」「きゃ」の拗音にしてネ", m.Reference())
		}
		switch utf8.RuneCountInString(m.Content) {
		case 1:
			if !regexp.MustCompile(`^\p{Hiragana}$`).MatchString(m.Content) {
				handler()
				return
			}
		case 2:
			if !xgo.Contains(contracteds, m.Content) {
				handler()
				return
			}
		default:
			handler()
			return
		}
	}

	haikuDraft := updatePoem(m.Content, poet, game)

	s.ChannelMessageSendReply(m.ChannelID, haikuDraft + "\n\nしばしお待ちを", m.Reference())

	return
}

func watchSuspendedGame(ctx context.Context, gamedoc *firestore.DocumentRef, c chan bool) {
	defer close(c)

	ctx2, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	var game UnknownPoetGame

	it := gamedoc.Snapshots(ctx2)
	for {
		snap, err := it.Next()
		if status.Code(err) == codes.DeadlineExceeded {
			log.Error(err)
			return
		}
		if err != nil {
			log.Error(err)
			return
		}
		if !snap.Exists() {
			// 消す予定なければ、消えるのはヤバイのでエラー
			log.Errorln("Document no longer exist")
			return
		}
		snap.DataTo(&game)
		if game.Status == GameStatusSuspend {
			c <- true
		}
	}
	return
}

var contracteds = []string{
	// 開拗音
	"きゃ", "きゅ", "きょ",
	"ぎゃ", "ぎゅ", "ぎょ",
	"しゃ", "しゅ", "しょ",
	"じゃ", "じゅ", "じょ",
	"ちゃ", "ちゅ", "ちょ",
	"ぢゃ", "ぢゅ", "ぢょ",
	"にゃ", "にゅ", "にょ",
	"ひゃ", "ひゅ", "ひょ",
	"びゃ", "びゅ", "びょ",
	"みゃ", "みゅ", "みょ",
	"りゃ", "りゅ", "りょ",
	// 合拗音
	"くゎ", "ぐゎ",
	// ↑ここまで五十音に記載？
	"くぁ", "ぐぁ",
	"つぁ", "つぃ", "つぇ", "つぉ",
	"てぃ", "でぃ", "とぅ", "どぅ", "でゅ",
	"ふぁ", "ふぃ", "ふぇ", "ふぉ", "ふゅ",
	"うぃ", "うぇ", "うぉ", "ゔぁ", "ゔぃ", "ゔぇ", "ゔぉ",
	"ちぇ", "しぇ", "じぇ",
}

func init() {
	ps = pubsub.New()

	ctx := context.Background()
	firestore := service.InitializeFirestore(ctx)
	var err error
	haikuRepo, err = repository.NewHaikuRepository(ctx, firestore)
	if err != nil {
		log.Errorf("Cannot init haikuGame: %v", err)
		return
	}
	addCommand(haiku, haikuHandler)
	addHandler(dmHandler)
}

