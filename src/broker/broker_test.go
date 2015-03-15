package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_PingHandler(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	w := httptest.NewRecorder()

	pingHandler(w, r)

	if w.Body.String() != "pong" {
		t.Errorf("expected %q but instead got %q", "pong", w.Body.String())
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
	r, _ := http.NewRequest("GET", "http://foo.example.com/queue1", strings.NewReader(""))
	w := httptest.NewRecorder()

	putHandler(w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}

func Test_PutHandler2(t *testing.T) {
	r, _ := http.NewRequest("POST", "http://foo.example.com/queue1", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	putHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}
