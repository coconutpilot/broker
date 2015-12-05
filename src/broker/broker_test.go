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
}

func datadir_testing() (dir string) {
	var err error
	dir, err = ioutil.TempDir("", "broker_test_")
	if err != nil {
		log.Panic("Failed ioutil.TempDir(): %s", err)
	}
	err = os.Mkdir(dir+"/testing", 0777)
	if err != nil {
		log.Panic("Failed os.Mkdir(): %s", err)
	}
	log.Printf("Test datadir is: %s\n", dir)
	return
}

func Test_PingHandler_GET(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	time := time.Now()
	uri := fmt.Sprintf("http://foo.example.com/ping?q=%s", time.String())
	r, _ := http.NewRequest("GET", uri, nil)
	w := httptest.NewRecorder()

	pingHandler(ctx, w, r)

	if w.Body.String() != uri {
		t.Errorf("expected %q but instead got %q", uri, w.Body.String())
	}
}

func Test_PingHandler_POST(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	time := time.Now()
	r, _ := http.NewRequest("POST", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(ctx, w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_PingHandler_PUT(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	time := time.Now()
	r, _ := http.NewRequest("PUT", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(ctx, w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_PingHandler_Invalid(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	time := time.Now()
	r, _ := http.NewRequest("DELETE", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	pingHandler(ctx, w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}

func Test_ViewHandler(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	testdata := "http://foo.example.com/nada"
	r, _ := http.NewRequest("GET", testdata, nil)
	w := httptest.NewRecorder()

	viewHandler(ctx, w, r)

	if w.Body.String() != testdata {
		t.Errorf("expected %q but instead got %q", testdata, w.Body.String())
	}
}

func Test_QueueHandler_Invalid_GET(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	r, _ := http.NewRequest("GET", "http://foo.example.com/queue/invalid", strings.NewReader(""))
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func Test_QueueHandler_Invalid_PUT(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/invalid", strings.NewReader(""))
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func Test_QueueHandler_PUT(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/testing", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}

func Test_QueueHandler_GET(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	r, _ := http.NewRequest("GET", "http://foo.example.com/queue/testing", nil)
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 404 {
		t.Errorf("Expected: 404 Got: %d", w.Code)
	}
}

func Test_QueueHandler_GET2(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	time := time.Now()
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/testing", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "http://foo.example.com/queue/testing", nil)
	w = httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func Test_QueueHandler_Invalid_Method(t *testing.T) {
	var ctx context
	ctx.datadir = datadir_testing()

	r, _ := http.NewRequest("POST", "http://foo.example.com/queue/", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	queueHandler(ctx, w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}
