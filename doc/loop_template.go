package doc

import (
	"github.com/bwmarrin/discordgo"

	"github.com/konafx/natalya/loop"
)

/*
	main.go
	```
	loops := []*loop.Loop{
		// add here loop
		&cogs.SampleLoop,
	}

	tasks := map[string]func(s *discordgo.Session) {
		// add here task
		cogs.SampleLoop.Name: cogs.SampleTask,
	}
	```
*/

var SampleLoop = loop.Loop{
	Name:		"Name",
	Seconds:	0,
	Minites:	0,
	Hours:		24,
	Init:		false,
}

func SampleTask(s *discordgo.Session) {
	return
}
