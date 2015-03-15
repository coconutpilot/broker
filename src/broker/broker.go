package main

import (
	"code.google.com/p/gcfg"
	"daemon"
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Config struct {
	Daemon struct {
		Port int
	}
}

var wg sync.WaitGroup

func viewHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("viewHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, html.EscapeString(r.URL.Path))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("pingHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, "pong")
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("putHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	if r.Method != "POST" {
		http.Error(w, "Wrong method", 405)
	}
	log.Println(r.URL.Path)

	queue := strings.TrimLeft(r.URL.Path, "/put/")
	log.Println(queue)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Body error: %s", err)
	}
	log.Printf("%s", body)
	fmt.Fprint(w, "OK")
	time.Sleep(100000)
	log.Println("POST done")
}

var cfgfile = flag.String("config", "broker.cfg", "config filename")

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("main()")

	// Start signal handling early (avoid case when signals are delivered before handler installed)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)
	signal.Notify(stop, syscall.SIGHUP)
	signal.Notify(stop, syscall.SIGTERM)

	flag.Parse()
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, *cfgfile)
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	srv_addr := fmt.Sprintf(":%d", cfg.Daemon.Port)

	l, err := daemon.New(srv_addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/put/", putHandler)

	server := http.Server{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		server.Serve(l)
	}()

	// This is blocking:
	select {
	case signal := <-stop:
		log.Printf("Got signal: %v\n", signal)
	}
	l.Stop()
	log.Println("Waiting for daemon to shutdown")
	wg.Wait()
	log.Println("Exiting")
}
