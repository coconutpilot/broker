package main

import (
	"code.google.com/p/gcfg"
	"fmt"
	"html"
	"log"
	"net/http"
)

type Config struct {
	Daemon struct {
		Port int
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, html.EscapeString(r.URL.Path))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func main() {
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "broker.cfg")
	if err != nil {
		log.Fatalf("Failed to parse gcfg data: %s", err)
	}

	srv_addr := fmt.Sprintf(":%d", cfg.Daemon.Port)

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(srv_addr, nil)
}
