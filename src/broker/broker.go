package main

import (
	"code.google.com/p/gcfg"
	"daemon"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Daemon struct {
		Port int
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, html.EscapeString(r.URL.Path))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, "pong")
}

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("main()")
	// Start signal handling early (avoid case when signals are delivered before handler installed)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, syscall.SIGTERM)

	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "broker.cfg")
	if err != nil {
		log.Fatalf("Failed to parse gcfg data: %s", err)
	}

	srv_addr := fmt.Sprintf(":%d", cfg.Daemon.Port)

	l, err := daemon.New(srv_addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/ping", pingHandler)

	server := http.Server{}

	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(l)
	}()

	select {
	case signal := <-stop:
		fmt.Printf("Got signal:%v\n", signal)
	}
	log.Println("Stopping listener")
	l.Stop()
	log.Println("Waiting on server")
	wg.Wait()
	log.Println("Exiting")

}
