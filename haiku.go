package main

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"

	"cloud.google.com/go/firestore"
	u "github.com/konafx/natalya/util"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

type Poem struct {
	Next	int			`json:"next" firestore:"next"`
	UserID	string		`json:"userId" firestore:"userId"`
	Poem	string		`json:"poem" firestore:"poem"`
}

type UnknownPoetGame struct {
	GuildID	string		`json:"guildId" firestore:"guildId"`
	UserIDs	[]string	`json:"userIds" firestore:"userIds"`
	Stage	int			`json:"stage" firestore:"stage"`
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
	var poems []Poem
	for k, v := range i.Data.Options {
		u := v.UserValue(s)
		poem := Poem{
			Next:	k,
			UserID:	u.ID,
		}
		poems = append(poems, poem)
	}
	log.Debugln(poems)

	// TODO: 同じユーザーが2つ以上のゲームをできないようにするフィルター、データ構造
	var userIds []string
	for _, v := range poems {
		userIds = append(userIds, v.UserID)
	}
	sort.Slice(userIds, func(i, j int) bool { return userIds[i] < userIds[j] })
	for k, v := range userIds {
		if k == len(userIds) {
			break
		}
		if v == userIds[k+1] {
			u.InteractionErrorResponse(s, i.Interaction, fmt.Sprintf("%s が二回指名されてるゾ", u.ToUser(v)))
			return
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(poems), func(i, j int) { poems[i], poems[j] = poems[j], poems[i] })

	ctx := context.Background()
	client := createClient(ctx)
	defer client.Close()

	var gamedoc *firestore.DocumentRef
	{
		game := UnknownPoetGame{
			GuildID:	i.GuildID,
			UserIDs:	userIds,
		}
		doc, _, err := client.Collection("unknownPoetGames").Add(ctx, game)
		if err != nil {
			u.InteractionErrorResponse(s, i.Interaction, "ウーン、ここは俳句を詠むにはうるさすぎるみたイ…")
			log.Error(err)
			return
		}
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
	for k, v := range poems {
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

	// ゲーム開始
	{
		c := make(chan struct{})
		waitFinishWritingPoems(s, ctx, gamedoc, c)
		_, ok := <-c
		if !ok {
			u.InteractionErrorResponse(s, i.Interaction, "なんか起きたので中断されました")
			return
		}
	}

	poems2, _ := getPoems(ctx, gamedoc)

	result := "俳句ができたヨ\n"
	for _, v := range poems2 {
		result = fmt.Sprintf("%s\n%s %s", result, u.ToUser(v.UserID), v.Poem)
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
}

func waitFinishWritingPoems(s *discordgo.Session, ctx context.Context, gamedoc *firestore.DocumentRef, c chan struct{}) {
	defer close(c)

	poems, err := getPoems(ctx, gamedoc)
	if err != nil {
		log.Error(err)
		return
	}

	// DM を送る
	for _, v := range poems {
		if err = sendDraftPoem(s, ctx, gamedoc, v); err != nil {
			log.Error(err)
			return
		}
	}

	// 各ステージの終了を検知する
	for {
		c2 := make(chan int)
		waitFinishStage(ctx, gamedoc, c2)
		stage, ok := <-c2
		if !ok {
			log.Errorf("something happened on waitFinishStage")
			return
		}

		stage = stage + 1
		// TODO stage 更新処理
		gamedoc.Update(ctx, []firestore.Update{
			{
				Path: "Stage",
				Value: stage,
			},
		})

		// stage = [0, 16]
		if stage > 16 {
			c <- struct{}{}
			return
		}
	}
}

// 各ステージの終了を待つ
func waitFinishStage(ctx context.Context, gamedoc *firestore.DocumentRef, c chan int) {
	defer close(c)

	ctx2, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

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
		gamesnap, err := gamedoc.Get(ctx)
		var game UnknownPoetGame
		gamesnap.DataTo(&game)
		if snap != nil {
			iter := gamedoc.Collection("poems").Documents(ctx)
			for {
				doc, err := iter.Next()
				if err == iterator.Done {
					c <- game.Stage
					return
					// ok
				}
				if err != nil {
					log.Errorf("Documents.Next: %v", err)
					return
				}
				var poem Poem
				doc.DataTo(&poem)
				// ひとつでも足りなければフェーズ続行
				// TODO: 拗音対応
				if len(poem.Poem) < game.Stage {
					break
				}
			}
		}
	}
}

// 各ユーザーにDMおくって放置
func sendDraftPoem(s *discordgo.Session, ctx context.Context, gamedoc *firestore.DocumentRef, poem *Poem) error {
	snap, err := gamedoc.Collection("poems").Doc(strconv.Itoa(poem.Next)).Get(ctx)
	if err != nil {
		return err
	}
	var target Poem
	snap.DataTo(&target)

	// send dm
	ch, err := s.UserChannelCreate(target.UserID)
	if err != nil {
		return err
	}

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

	log.Info("get game")
	var gameRef *firestore.DocumentRef
	{
		it := client.Collection("unknownPoetGames").Where("userIds", "array-contains", m.Author.ID).Documents(ctx)
		defer it.Stop()

		for {
			doc, err := it.Next()
			if err == iterator.Done {
				return
			}
			if err != nil {
				log.Error(err)
				s.ChannelMessageSendReply(m.ChannelID, "エラーだヨ…", m.Reference())
				return
			}
			gameRef = doc.Ref
			break
		}
	}

	// 参加者であることがおそらく保証されている
	// TODO: ここらへんで m.Content が一文字だけか、ひらがなかチェックしていこう
	if len(m.Content) != 1 {
		s.ChannelMessageSendReply(m.ChannelID, "1文字にしてね", m.Reference())
		return
	}

	var poem Poem
	log.Info("get your poem")
	{
		it := gameRef.Collection("poems").Where("userId", "==", m.Author.ID).Documents(ctx)
		defer it.Stop()

		for {
			doc, err := it.Next()
			if err == iterator.Done {
				log.Error("Cannot find poem in game played by dm user")
				return
			}
			if err != nil {
				log.Error(err)
				s.ChannelMessageSendReply(m.ChannelID, "エラーだヨ…", m.Reference())
				return
			}
			doc.DataTo(&poem)
			break
		}
	}

	// poem が格納されている document の ID
	var ID string
	{
		docsnap, err := gameRef.Collection("poems").Doc(strconv.Itoa(poem.Next)).Get(ctx)
		if err != nil {
			log.Error(err)
			return
		}
		pmap := docsnap.Data()
		for k, v := range pmap {
			ID = k
			poem = v.(Poem)
		}
	}
	log.Debugf("docID: %s, poem: %v", ID, poem)

	var game UnknownPoetGame
	{
		docsnap, err := gameRef.Get(ctx)
		if err != nil {
			log.Error(err)
			return
		}
		docsnap.DataTo(&game)
	}

	// write
	if len(game.UserIDs) >= poem.Next {
		poem.Next = 0
	} else {
		poem.Next = poem.Next + 1
	}

	poem.Poem = poem.Poem + m.Content

	// update
	_, err := gameRef.Collection("poems").Doc(ID).Set(ctx, poem)
	if err != nil {
		log.Error(err)
		return
	}
	s.ChannelMessageSendReply(m.ChannelID, "しばしお待ちを", m.Reference())

	return
}

func getPoems(ctx context.Context, gamedoc *firestore.DocumentRef) (poems []*Poem, err error) {
	it := gamedoc.Collection("poems").Documents(ctx)
	defer it.Stop()

	for {
		var snap *firestore.DocumentSnapshot
		snap, err = it.Next()
		if err == iterator.Done {
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
