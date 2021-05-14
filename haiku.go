package main

import (
	"context"
	"fmt"
	"math/rand"
	"math"
	"regexp"
	"sort"
	"strconv"
	"time"
	"unicode/utf8"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	u "github.com/konafx/natalya/util"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// docID: increment
type Poem struct {
	Poem	string		`json:"poem" firestore:"poem"`
}

// docID: increment
type Poet struct {
	Next	int			`json:"next" firestore:"next"`
	UserID	string		`json:"userId" firestore:"userId"`
}

// docID: UserID
type UnknownPoetGamePlayer struct {
	PlayingGameID	string	`json:"playingGameId" firestore:"playingGameId"`
}

// docID: auto
type UnknownPoetGame struct {
	GuildID			string		`json:"guildId" firestore:"guildId"`
	NumberOfPlayers	int			`json:"numberOfPlayers" firestore:"numberOfPlayers"`
	Stage			int			`json:"stage" firestore:"stage"`
}

var Haiku Command = &discordgo.ApplicationCommand{
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
			Next:	k,
			UserID:	u.ID,
		}
		poets = append(poets, poet)
	}
	log.Debugln(poets)

	// 重複チェック
	{
		var userIds []string
		for _, v := range poets {
			userIds = append(userIds, v.UserID)
		}
		sort.Slice(userIds, func(i, j int) bool { return userIds[i] < userIds[j] })
		for k, v := range userIds {
			log.Debugln(k, v)
			if k + 1 == len(userIds) {
				break
			}
			if v == userIds[k+1] {
				u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("%s が二回指名されてるゾ", u.ToUser(v)))
				return
			}
		}
	}

	// 順番ランダム
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(poets), func(i, j int) { poets[i], poets[j] = poets[j], poets[i] })

	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	var gamedoc *firestore.DocumentRef
	{
		game := UnknownPoetGame{
			GuildID:	i.GuildID,
			NumberOfPlayers: len(poets),
			Stage:		1,
		}
		doc, _, err := client.Collection("unknownPoetGames").Add(ctx, game)
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, "ウーン、ここは俳句を詠むにはうるさすぎるみたイ…")
			log.Error(err)
			return
		}

		// TODO: 同じユーザーが2つ以上のゲームをできないようにするフィルター、データ構造
		for _, v := range poets {
			player := UnknownPoetGamePlayer{
				PlayingGameID: doc.ID,
			}
			if _, err := client.Collection("unknownPoetGamePlayers").Doc(v.UserID).Set(ctx, player); err != nil {
				u.InteractionErrorResponse(s, i.Interaction, "ウーン、俳句を詠む心が備わってないないみたイ…")
				log.Error(err)
				return
			}
		}

		for k, v := range poets {
			if _, err := doc.Collection("poets").Doc(strconv.Itoa(k)).Set(ctx, v); err != nil {
				u.InteractionErrorResponse(s, i.Interaction, "ウーン、俳句を詠む心が備わってないないみたイ…")
				log.Error(err)
				return
			}
		}

		poems := make([]Poem, len(poets))
		for k, v := range poems {
			if _, err := doc.Collection("poems").Doc(strconv.Itoa(k)).Set(ctx, v); err != nil {
				u.InteractionErrorResponse(s, i.Interaction, "ウーン、俳句を詠む心が備わってないないみたイ…")
				log.Error(err)
				return
			}
		}
		gamedoc = doc
	}

	message := "今回の俳人はお前らダ！よろしくナ♪\n"
	for k, v := range poets {
		if k == 0 {
			message = fmt.Sprintf("%s%s", message, u.ToUser(v.UserID))
			continue
		}
		message = fmt.Sprintf("%s, %s", message, u.ToUser(v.UserID))
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
	// ゲーム開始
	{
		c := make(chan struct{})
		log.Debugln("Start to wait writing poems")
		go waitFinishWritingPoems(s, ctx, gamedoc, c)
		_, ok := <-c
		log.Debugln("End to wait writing poems")
		if !ok {
			u.InteractionErrorResponse(s, i.Interaction, "なんか起きたので中断されました")
			return
		}
		// ゲーム終了後はresult としてどっか別のDBに保存 and game からは消す
	}

	poems, _ := getPoems(ctx, gamedoc)

	result := "俳句ができたヨ\n"
	for i, l := 0, len(poems); i<l; i++ {
		result = fmt.Sprintf("%s\n%s %s", result, u.ToUser(poets[i].UserID), poems[i].Poem)
	}
	log.Debug(result)

	s.ChannelMessageSend(i.ChannelID, result)

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
}

func waitFinishWritingPoems(s *discordgo.Session, ctx context.Context, gamedoc *firestore.DocumentRef, c chan struct{}) {
	defer close(c)

	// 各ステージの終了を検知する
	for {
		poets, err := getPoets(ctx, gamedoc)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugln(poets)

		// DM を送る
		for _, v := range poets {
			if err = sendDraftPoem(s, ctx, gamedoc, v); err != nil {
				log.Error(err)
				return
			}
		}

		c2 := make(chan int)
		log.Debugln("Start to wait finish stage")
		go waitFinishStage(ctx, gamedoc, c2)
		stage, ok := <-c2
		log.Debugln("End to wait finish stage")
		if !ok {
			log.Error("something happened on waitFinishStage")
			return
		}

		stage = stage + 1
		// TODO stage 更新処理
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
		if stage > 3 {
			c <- struct{}{}
			return
		}
	}
}

// 各ステージの終了を待つ
func waitFinishStage(ctx context.Context, gamedoc *firestore.DocumentRef, c chan int) {
	defer close(c)

	ctx2, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	var game UnknownPoetGame
	{
		snap, _ := gamedoc.Get(ctx)
		snap.DataTo(&game)
	}

	it := gamedoc.Collection("poems").Snapshots(ctx2)
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
		log.Debug("catch change poem")
		if snap != nil {
			iter := gamedoc.Collection("poems").Documents(ctx)
			for {
				doc, err := iter.Next()
				if err == iterator.Done {
					log.Info("waitFinish!!")
					c <- game.Stage
					return // ok
				}
				if err != nil {
					log.Errorf("Documents.Next: %v", err)
					return
				}
				var poem Poem
				doc.DataTo(&poem)
				log.Debug(poem.Poem, utf8.RuneCountInString(poem.Poem), game.Stage)
				// TODO: 拗音対応
				if utf8.RuneCountInString(poem.Poem) < game.Stage {
					break
				}
			}
		}
	}
}

// 各ユーザーにDMおくって放置
func sendDraftPoem(s *discordgo.Session, ctx context.Context, gamedoc *firestore.DocumentRef, poet *Poet) error {
	// send dm
	ch, err := s.UserChannelCreate(poet.UserID)
	if err != nil {
		return err
	}

	snap, err := gamedoc.Collection("poems").Doc(strconv.Itoa(poet.Next)).Get(ctx)
	if err != nil {
		return err
	}

	var poem Poem
	snap.DataTo(&poem)

	var message string
	if len(poem.Poem) == 0 {
		message = "最初の一文字を決めてほしいナ♪"
	} else {
		// TODO: 5 7 5 (合計17もじ）、空白を埋める（○）で
		message = fmt.Sprintf("> %s\n\n次の一文字を入力してネ！", poem.Poem)
	}

	if _, err = s.ChannelMessageSend(ch.ID, message); err != nil {
		return err
	}

	return nil
}

// リプきた！
func UnknownPoetDMHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != "" {
		return
	}
	if s.State.User.ID == m.Author.ID {
		return
	}

	ctx := context.Background()
	client := createClient(ctx)

	log.Info("get player")
	var player UnknownPoetGamePlayer
	{
		snap, err := client.Collection("unknownPoetGamePlayers").Doc(m.Author.ID).Get(ctx)
		if status.Code(err) == codes.NotFound {
			return
		}
		if err != nil {
			log.Error(err)
			return
		}
		snap.DataTo(&player)
	}

	// 参加者であることがおそらく保証されている
	log.Info("get game")
	var gamedoc *firestore.DocumentRef
	{
		snap, err := client.Collection("unknownPoetGames").Doc(player.PlayingGameID).Get(ctx)
		if err != nil {
			log.Error(err)
			s.ChannelMessageSendReply(m.ChannelID, "エラーだヨ…", m.Reference())
			return
		}
		gamedoc = snap.Ref
	}

	if !regexp.MustCompile(`^\p{Hiragana}$`).MatchString(m.Content) {
		s.ChannelMessageSendReply(m.ChannelID, "1文字にしてね", m.Reference())
		return
	}

	log.Info("get poet")
	var poetsnap *firestore.DocumentSnapshot
	{
		it := gamedoc.Collection("poets").Where("userId", "==", m.Author.ID).Documents(ctx)
		defer it.Stop()

		for {
			snap, err := it.Next()
			if err == iterator.Done {
				log.Error("Cannot find poem in game played by dm user")
				return
			}
			if err != nil {
				log.Error(err)
				s.ChannelMessageSendReply(m.ChannelID, "エラーだヨ…", m.Reference())
				return
			}
			poetsnap = snap
			break
		}
	}
	var poet Poet
	poetsnap.DataTo(&poet)

	log.Infof("get poem %d", poet.Next)
	var poem Poem
	{
		snap, err := gamedoc.Collection("poems").Doc(strconv.Itoa(poet.Next)).Get(ctx)
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, "エラーだヨ…", m.Reference())
			return
		}
		
		snap.DataTo(&poem)
	}

	log.Debug(poet, poem)

	var game UnknownPoetGame
	{
		snap, err := gamedoc.Get(ctx)
		if err != nil {
			log.Error(err)
			return
		}
		snap.DataTo(&game)
	}

	// write
	poem.Poem += m.Content

	// update
	log.Info("update poem and poet")
	_, err := gamedoc.Collection("poems").Doc(strconv.Itoa(poet.Next)).Set(ctx, poem)
	if err != nil {
		log.Error(err)
		return
	}
	_, err = poetsnap.Ref.Update(ctx, []firestore.Update{
		{
			Path: "next",
			Value: int(math.Mod(float64(poet.Next + 1), float64(game.NumberOfPlayers))),
		},
	})
	if err != nil {
		log.Error(err)
		return
	}

	s.ChannelMessageSendReply(m.ChannelID, "しばしお待ちを", m.Reference())

	return
}

func getPoets(ctx context.Context, gamedoc *firestore.DocumentRef) (poets []*Poet, err error) {
	it := gamedoc.Collection("poets").Documents(ctx)
	defer it.Stop()

	for {
		var snap *firestore.DocumentSnapshot
		snap, err = it.Next()
		if err == iterator.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		var poet Poet
		snap.DataTo(&poet)
		poets = append(poets, &poet)
	}
	return
}

func getPoems(ctx context.Context, gamedoc *firestore.DocumentRef) (poems []*Poem, err error) {
	it := gamedoc.Collection("poems").Documents(ctx)
	defer it.Stop()

	for {
		var snap *firestore.DocumentSnapshot
		snap, err = it.Next()
		if err == iterator.Done {
			err = nil
			break
		}
		if err != nil {
			return
		}
		var poem Poem
		snap.DataTo(&poem)
		poems = append(poems, &poem)
	}
	return
}

func getGame(ctx context.Context, gamedoc *firestore.DocumentRef) (game *UnknownPoetGame, err error) {
	snap, err := gamedoc.Get(ctx)
	if err != nil {
		return
	}
	snap.DataTo(&game)
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

func init() {
	addCommand(Haiku, HaikuHandler)
	addHandler(UnknownPoetDMHandler)
}
