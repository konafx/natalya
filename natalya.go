package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
) 

var (
	GuildId			= flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
	BotToken		= flag.String("token", "", "Bot access token")
	RemoveCommand	= flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
	Prefix			= "/natalya"
)

var s *discordgo.Session

func init() {
	log.SetLevel(log.DebugLevel)
}
func init() {
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
		&Hello,
		&SuperChat,
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		"hello": HelloHandler,
		"superchat": SuperChatHandler,
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.Data.Name]; ok {
			h(s, i)
		}
	})

	if err := s.Open(); err != nil {
		fmt.Println(err)
		return
	}
	defer s.Close()

	log.Printf("%#v", s.State)

	for _, v := range commands {
		log.Println(*GuildId, v)
		_, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildId, v)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", err, v.Name)
		}
	}

	fmt.Println("Natalya is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Println("\nTchau!\n")
	return
}

func ready(s *discordgo.Session, e *discordgo.Ready) {
	s.UpdateGameStatus(0, Prefix + " <command>")
}
