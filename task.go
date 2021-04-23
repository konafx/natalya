package main

import (
	"time"
)

type Loop struct {
	Name		string
	Seconds		int
	Minites		int
	Hours		int
	Init		bool
}

func (loop *Loop) ExecFn(fn func(), init bool) {
	d := time.Hour * time.Duration(loop.Hours) +
	time.Minute * time.Duration(loop.Minites) +
	time.Second * time.Duration(loop.Seconds)
	ticker := time.NewTicker(d)
	if init {
		fn()
	}
	for {
		<-ticker.C
		fn()
	}
}
