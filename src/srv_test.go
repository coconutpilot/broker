package main

import (
    "testing"
    "net/http"
    "net/http/httptest"
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
    r, _ := http.NewRequest("GET", "http://foo.example.com/asdf", nil)
    w := httptest.NewRecorder()

    viewHandler(w, r)

    if w.Body.String() != "/asdf" {
        t.Errorf("expected %q but instead got %q", "/asdf", w.Body.String())
    }
}
