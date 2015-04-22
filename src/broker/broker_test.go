package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_PingHandler(t *testing.T) {
	time := time.Now()
	r, _ := http.NewRequest("POST", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_ViewHandler(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://foo.example.com/queue1", nil)
	w := httptest.NewRecorder()

	viewHandler(w, r)

	if w.Body.String() != "/queue1" {
		t.Errorf("expected %q but instead got %q", "/queue1", w.Body.String())
	}
}

func Test_PutHandler1(t *testing.T) {
	// storage dir not setup
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue1", strings.NewReader(""))
	w := httptest.NewRecorder()

	putHandler(w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func Test_PutHandler2(t *testing.T) {
	// normal put
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue1", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	var err error
	dir, err = ioutil.TempDir("", "broker_test")
	if err != nil {
		t.Fatalf("Failed creating TempDir(): %s", err)
	}

	putHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}

func Test_PutHandler3(t *testing.T) {
	// wrong method
	r, _ := http.NewRequest("POST", "http://foo.example.com/queue1", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	putHandler(w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}
