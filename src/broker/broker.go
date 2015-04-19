package main

import (
	"code.google.com/p/gcfg"
	"daemon"
	//"path/filepath"
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
	Storage struct {
		Dir string
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

	if r.Method != "POST" {
		http.Error(w, "Wrong method", 405)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Body error: %s", err)
	}
	log.Printf("%s", body)

	fmt.Fprintf(w, "%s", body)
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("putHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	if r.Method != "PUT" {
		http.Error(w, "Wrong method", 405)
		return
	}
	log.Println(r.URL.Path)

	queue := strings.TrimLeft(r.URL.Path, "/put/")
	log.Println(queue)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Body error: %s", err)
		http.Error(w, "Retry operation", 400)
		return
	}
	// log.Printf("%s", body)

	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%s/%d", dir, timestamp)
	log.Printf("Creating file: %s", filename)

	f, err := os.Create(filename)
	if err != nil {
		log.Printf("Create file error: %s", err)
		http.Error(w, "Retry operation", 503)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s", body)

	fmt.Fprintln(w, "OK")
	log.Println("putHandler() exit")
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("getHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	//	fmt.Fprintf(w, html.EscapeString(r.URL.Path))

	d, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Open dir error: %s", err)
	}
	defer d.Close()

	de, err := d.Readdirnames(10)
	if err != nil {
		// check for io.EOF
		log.Fatalf("Readdir error: %s", err)
	}

	var data []byte
	for _, de := range de {
		filename := dir + "/" + de
		log.Printf("Trying to lock file: %s\n", filename)
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("Open file error: %s", err)
		}
		defer f.Close()

		fd := f.Fd()
		err = syscall.Flock(int(fd), syscall.LOCK_EX+syscall.LOCK_NB)
		if err != nil {
			log.Fatalf("Lock blowed up: %s\n", err)
		}
		time.Sleep(10000000000)
		data, err = ioutil.ReadAll(f)
		if err != nil {
			log.Fatalf("Read file error: %s", err)
		}
		// order is wrong
		os.Remove(filename)
		if err != nil {
			log.Fatalf("Remove file error: %s\n", err)
		}

		break
		log.Fatalln("WTF")
	}
	fmt.Fprintf(w, string(data))
}

func getnextHandler(w http.ResponseWriter, r *http.Request) {
	wg.Add(1)
	defer wg.Done()

	log.Println("getnextHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, html.EscapeString(r.URL.Path))
}

var cfgfile = flag.String("config", "broker.cfg", "config filename")
var dir string

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

	dir = cfg.Storage.Dir
	//	log.Fatalf("%s", dir)

	srv_addr := fmt.Sprintf(":%d", cfg.Daemon.Port)

	l, err := daemon.New(srv_addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %s", err)
	}

	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/put/", putHandler)
	http.HandleFunc("/get/", getHandler)
	http.HandleFunc("/getnext/", getnextHandler)

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
