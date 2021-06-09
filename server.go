package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Natalya is running")
}

func Server(ctx context.Context, port int) {
	go func() {
		http.HandleFunc("/", handler)
		if err := http.ListenAndServe(":" + strconv.Itoa(port), nil); err != nil {
			log.Error(err)
		}
	}()
	log.Println("server is ready to handle requests at ", port)
	select {
	case <-ctx.Done():
		log.Println("server shutdown")
		return
	}
}
