package main

import (
    "fmt"
    "net/http"
)

func viewHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<div>Hello world!</div>")
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "pong")
}

func main() {
    http.HandleFunc("/", viewHandler)
    http.HandleFunc("/ping", pingHandler)
    http.ListenAndServe(":8080", nil)
}
