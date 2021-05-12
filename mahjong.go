package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/konafx/natalya/loop"
	wr "github.com/mroth/weightedrand"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Hand struct {
	Name	string	`yaml:"name"`
	Fan		int		`yaml:"fan"`
}

const (
	MaxFan = 13
	HandsFile = "assets/hands.yml"
	SerifsFile = "assets/serifs.yml"
)

var TodayHandLoop = loop.Loop{
	Name:		"TodayHand",
	Seconds:	0,
	Minites:	0,
	Hours:		24,
	Init:		true,
}

var todayHand Hand

func TodayHandTask(s *discordgo.Session) {
	buf, err := ioutil.ReadFile(HandsFile)
	if err != nil {
		log.Errorf("Cannot read handsfile(%s), err: %v", HandsFile, err)
		return
	}

	var hands []Hand
	err = yaml.Unmarshal(buf, &hands)
	if err != nil {
		log.Errorf("Cannot read yaml file, err: %v", err)
		return
	}

	var hand Hand
	hand, err = choiceHand(hands)
	if err != nil {
		log.Errorf("hand choice error: %v", err)
		return
	}

	todayHand = hand

	log.Infof("Today hand is %s", todayHand.Name)
}

func choiceHand(hands []Hand) (Hand, error) {
	var hand Hand = Hand{}

	rand.Seed(time.Now().UTC().UnixNano())

	choices := make([]wr.Choice, len(hands))
	for i, v := range hands {
		choices[i] = wr.NewChoice(v, MaxFan - uint(v.Fan) + 1)
	}

	chooser, err := wr.NewChooser(choices...)
	if err != nil {
		log.Errorf("Cannot create chooser, err: %v", err)
		return hand, err
	}

	hand = chooser.Pick().(Hand)
	return hand, nil
}

func choiceSerif() (string, error) {
	buf, err := ioutil.ReadFile(SerifsFile)
	if err != nil {
		log.Errorf("Cannot read handsfile(%s), err: %v", SerifsFile, err)
		return "", err
	}

	var serifs []string
	err = yaml.Unmarshal(buf, &serifs)
	if err != nil {
		log.Errorf("Cannot unmarshal serifs: %v", err)
		return "", err
	}

	i := rand.Intn(len(serifs))
	return serifs[i], nil
}

var TodayHand = discordgo.ApplicationCommandOption{
	Type: discordgo.ApplicationCommandOptionSubCommand,
	Name: "今日の役",
	Description: "デイリーミッション♪",
}

var Mahjong Command = &discordgo.ApplicationCommand{
	Name: "mahjong",
	Description: "背中が煤けてるゼ…",
	Options: []*discordgo.ApplicationCommandOption{
		&TodayHand,
	},
}

func MahjongHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	content := ""

	switch i.Data.Options[0].Name {
	case "今日の役":
		var int2kanji = map[int]string{
			1:"一", 2:"二", 3:"三", 4:"四", 5:"五", 6:"六",
			7:"七", 8:"八", 9:"九", 10:"十", 11: "十一", 12:"十二", 13:"十三",
		}
		serif, err := choiceSerif()
		if err != nil {
			log.Error(err)
			return
		}
		content = fmt.Sprintf("今日の役は **%s**♪\n%s翻ダ！ %s", todayHand.Name, int2kanji[todayHand.Fan], serif)
	default:
		content = "理解できない"
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: content,
		},
	})
	return
}

func init() {
	addCommand(Mahjong, MahjongHandler)
	addLoopTask(&TodayHandLoop, TodayHandTask)
}
