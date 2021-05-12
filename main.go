package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/konafx/natalya/loop"
	log "github.com/sirupsen/logrus"
)

var (
	GuildId			= flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken		= flag.String("token", "", "Bot access token")
	RemoveCommand	= flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

type Command *discordgo.ApplicationCommand

type CommandHandler func(*discordgo.Session, *discordgo.InteractionCreate)

var (
	s *discordgo.Session
	commands	[]Command
	commandHandlers		map[string]CommandHandler = map[string]CommandHandler{}
	loops	[]*loop.Loop
	tasks	map[string]func(*discordgo.Session) = map[string]func(*discordgo.Session){}
	handlers	[]interface{}
	initializers	[]func()
)

func addCommand(c Command, ch CommandHandler) {
	commands = append(commands, c)
	commandHandlers[c.Name] = ch
}

func addLoopTask(l *loop.Loop, t func(*discordgo.Session)) {
	loops = append(loops, l)
	tasks[l.Name] = t
}

func addHandler(h ...interface{}){
	handlers = append(handlers, h...)
}

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
	s.LogLevel = 4
	if err != nil {
		fmt.Printf("error creating Discord session: %v", err)
		os.Exit(1)
	}
}

func main() {
	s.AddHandler(ready)

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

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

	for _, v := range handlers {
		s.AddHandler(v)
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
