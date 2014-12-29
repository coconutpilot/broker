package main

import (
	"fmt"
	"html"
	"net/http"
)

func viewHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, html.EscapeString(r.URL.Path))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "pong")
}

func main() {
	http.HandleFunc("/", viewHandler)
	http.HandleFunc("/ping", pingHandler)
	http.ListenAndServe(":8080", nil)
}
