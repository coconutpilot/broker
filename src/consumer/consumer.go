package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	client := &http.Client{}
	time := time.Now()

	req, err := http.NewRequest("POST", "http://localhost:8080/ping", strings.NewReader(time.String()))
	req.Header.Add("User-Agent", "ElTaco")
	resp, err := client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Fatalf("http.Post => %v", err.Error())
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("\n%v\n\n", string(body))
}
