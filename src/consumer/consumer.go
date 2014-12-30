package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("http://192.168.1.11:8080/ping")

	if err != nil {
		log.Fatalf("http.Get => %v", err.Error())
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("\n%v\n\n", string(body))
}
