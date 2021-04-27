package main

import (
	"testing"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/konafx/natalya/cogs"
	"github.com/konafx/natalya/loop"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
) 

var (
	GuildId			= flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken		= flag.String("token", "", "Bot access token")
	RemoveCommand	= flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var s *discordgo.Session

func init() {
	log.SetLevel(log.DebugLevel)
}

func init() {
	testing.Init()
	flag.Parse()
}

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		fmt.Printf("error creating Discord session: %v", err)
		os.Exit(1)
	}
}

func main() {
	s.AddHandler(ready)

	commands := []*discordgo.ApplicationCommand{
		&cogs.Hello,
		&cogs.SuperChat,
		&cogs.Mahjong,
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cogs.Hello.Name: cogs.HelloHandler,
		cogs.SuperChat.Name: cogs.SuperChatHandler,
		cogs.Mahjong.Name: cogs.MahjongHandler,
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	loops := []*loop.Loop{
		&cogs.TodayHandLoop,
	}

	tasks := map[string]func(s *discordgo.Session) {
		cogs.TodayHandLoop.Name: cogs.TodayHandTask,
	}

	for _, loop := range loops {
		task := func (s *discordgo.Session) func() {
			return func () { tasks[loop.Name](s) }
		}
		go loop.ExecFn(task(s), loop.Init)
	}


	if err := s.Open(); err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()

	log.Printf("%#v", s.State)

	for _, v := range commands {
		log.Println(*GuildId, v)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildId, v)
		v.ID = cmd.ID
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", err, v.Name)
		}
	}


	fmt.Println("Natalya is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("Tchau!")
	if !*RemoveCommand { return }

	for _, v := range commands {
		if err := s.ApplicationCommandDelete(s.State.User.ID, *GuildId, v.ID); err != nil {
			log.Errorf("Skip delete cmd: %s (ID: %s)", v.Name, v.ID)
		}
	}

	return
}

func ready(s *discordgo.Session, e *discordgo.Ready) {
	s.UpdateGameStatus(0, "Dancing!")
}
