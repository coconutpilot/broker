package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	var err error
	ctx.datadir, err = ioutil.TempDir("", "broker_test")
	if err != nil {
		log.Panic("Failed ioutil.TempDir(): %s", err)
	}
	err = os.Mkdir(ctx.datadir+"/testing", 0777)
	if err != nil {
		log.Panic("Failed os.Mkdir(): %s", err)
	}
}

func Test_PingHandler_GET(t *testing.T) {
	time := time.Now()
	uri := fmt.Sprintf("http://foo.example.com/ping?q=%s", time.String())
	r, _ := http.NewRequest("GET", uri, nil)
	w := httptest.NewRecorder()

	pingHandler(w, r)

	if w.Body.String() != uri {
		t.Errorf("expected %q but instead got %q", uri, w.Body.String())
	}
}

func Test_PingHandler_POST(t *testing.T) {
	time := time.Now()
	r, _ := http.NewRequest("POST", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_PingHandler_PUT(t *testing.T) {
	time := time.Now()
	r, _ := http.NewRequest("PUT", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_ViewHandler(t *testing.T) {
	testdata := "http://foo.example.com/nada"
	r, _ := http.NewRequest("GET", testdata, nil)
	w := httptest.NewRecorder()

	viewHandler(w, r)

	if w.Body.String() != testdata {
		t.Errorf("expected %q but instead got %q", testdata, w.Body.String())
	}
}

func Test_QueueHandler_Invalid(t *testing.T) {
	// storage dir not setup
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/invalid", strings.NewReader(""))
	w := httptest.NewRecorder()

	queueHandler(w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func Test_QueueHandler_PUT(t *testing.T) {
	// normal put
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/testing", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	queueHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}

func Test_QueueHandler_GET(t *testing.T) {
	// normal get
	r, _ := http.NewRequest("GET", "http://foo.example.com/queue/testing", nil)
	w := httptest.NewRecorder()

	queueHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}

func Test_QueueHandler_Invalid_Method(t *testing.T) {
	// wrong method
	r, _ := http.NewRequest("POST", "http://foo.example.com/queue/", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	queueHandler(w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}
