package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type daemon struct {
	datadir string
	port    int
	mux     *http.ServeMux
}

func (d *daemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mux.ServeHTTP(w, r)
}

func (d *daemon) viewHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("viewHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")
	fmt.Fprintf(w, r.URL.String())
}

func (d *daemon) pingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("pingHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")

	switch r.Method {
	case "GET":
		log.Printf("pong: %s", r.URL.String())
		fmt.Fprintf(w, r.URL.String())

	case "POST":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Body error: %s", err)
			http.Error(w, "Error", 500)
			return
		}
		log.Printf("pong: %s", body)
		fmt.Fprintf(w, "%s", body)

	case "PUT":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Body error: %s", err)
			http.Error(w, "Error", 500)
			return
		}
		log.Printf("pong: %s", body)
		fmt.Fprintf(w, "%s", body)

	default:
		http.Error(w, "Wrong method", 405)
		return
	}
	log.Println("pingHandler() exit")
}

func (d *daemon) queueHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("queueHandler()")

	w.Header().Set("cache-control", "private, max-age=0, no-store")

	queue := strings.TrimPrefix(r.URL.Path, "/queue/")
	if r.URL.Path == queue {
		log.Println("Internal Error, path expected a prefix of /queue/, was", r.URL.Path)
		http.Error(w, "Internal Error", 500)
		return
	}

	queue = strings.TrimSuffix(queue, "/")

	switch r.Method {
	case "GET":
		dir, err := os.Open(d.datadir + "/" + queue)
		if err != nil {
			// XXX: return a permanent error if queue doesn't exist
			log.Printf("Open dir error: %s", err)
			http.Error(w, "Retry operation", 503)
			return
		}
		defer dir.Close()

		for {
			de, err := dir.Readdirnames(10)
			if err != nil {
				if err == io.EOF {
					log.Println("Queue empty")
					http.Error(w, "Queue empty", 404)
					return
				}
				log.Printf("Readdir error: %s", err)
				http.Error(w, "Retry operation", 503)
				return
			}

			var data []byte
			for _, de := range de {
				filename := d.datadir + "/" + queue + "/" + de
				log.Printf("Trying to lock file: %s\n", filename)
				f, err := os.Open(filename)
				if err != nil {
					log.Printf("Open file error: %s", err)

					// give the OS a chance to catch up
					time.Sleep(time.Millisecond * 10)
					// redo the Readdirnames
					break
				}
				defer f.Close()

				fd := f.Fd()
				err = syscall.Flock(int(fd), syscall.LOCK_EX+syscall.LOCK_NB)
				if err != nil {
					if err == syscall.EAGAIN {
						f.Close()
						continue
					}
					log.Printf("Flock error: %s\n", err)
					http.Error(w, "Retry operation", 503)
					return
				}

				// free up the handle immediately
				dir.Close()

				// time.Sleep(3 * time.Second) // testing aid

				// do a chunked read?
				data, err = ioutil.ReadAll(f)
				if err != nil {
					log.Printf("Read file error: %s", err)
					http.Error(w, "Retry operation", 503)
					return
				}

				fmt.Fprintf(w, string(data))

				err = os.Remove(filename)
				if err != nil {
					log.Printf("Remove file error: %s", err)
				}
				return
			}
		}
	case "PUT":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("Body error: %s", err)
			http.Error(w, "Retry operation", 503)
			return
		}
		// log.Printf("%s", body)

		timestamp := time.Now().UnixNano()
		filename := fmt.Sprintf("%s/%s/%d", d.datadir, queue, timestamp)
		log.Printf("Creating file: %s", filename)

		f, err := os.Create(filename)
		if err != nil {
			log.Printf("Create file error: %s", err)
			http.Error(w, "Retry operation", 503)
			return
		}
		defer f.Close()
		// time.Sleep(10 * time.Second) // testing aid

		fmt.Fprintf(f, "%s", body)

		fmt.Fprintln(w, "OK")

	default:
		log.Println("Wrong method", r.Method)
		http.Error(w, "Wrong method", 405)
		return
	}

	log.Println("putHandler() exit")
}

func main() {
	// Start signal handling early (avoid case when signals are delivered before handler installed)
	bus := make(chan os.Signal, 1)
	signal.Notify(bus, syscall.SIGINT)
	signal.Notify(bus, syscall.SIGHUP)
	signal.Notify(bus, syscall.SIGTERM)

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("main()")

	dirtemp := flag.String("datadir", "", "datadir")
	porttemp := flag.Int("port", 8080, "listen port")
	flag.Parse()

	if *dirtemp == "" {
		log.Fatal("--datadir <datadir> missing")
	}

	var d daemon
	d.datadir = fmt.Sprintf("%s", *dirtemp)
	d.port = *porttemp
	srvAddr := fmt.Sprintf(":%d", *porttemp)

	mux := http.NewServeMux()

	mux.HandleFunc("/", d.viewHandler)
	mux.HandleFunc("/ping/", d.pingHandler)
	mux.HandleFunc("/queue/", d.queueHandler)

	server := http.Server{Addr: srvAddr, Handler: mux}

	go func() {
		server.ListenAndServe()
		log.Print("Shutting down")
	}()

	// This is blocking:
	select {
	case signal := <-bus:
		log.Printf("Got signal: %v\n", signal)
	}
	server.Shutdown(context.Background())
	log.Println("Exiting")
}
