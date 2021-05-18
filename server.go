package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
)

var port string

func init() {
	port = os.Getenv("PORT")
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}

func Server() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		if err := http.ListenAndServe(":" + port, nil); err != nil {
			log.Error(err)
		}
	}()

	// シグナルを受信するまでブロック
	// https://github.com/gorilla/mux#graceful-shutdown
	// https://gist.github.com/enricofoltran/10b4a980cd07cb02836f70a4ab3e72d7
	log.Println("server is ready to handle requests at :%s", port)
	<-c

	return
}
