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

func datadirTesting() (dir string) {
	var err error
	dir, err = ioutil.TempDir("", "broker_test_")
	if err != nil {
		log.Fatalf("Failed ioutil.TempDir(): %s", err)
	}
	err = os.Mkdir(dir+"/testing", 0777)
	if err != nil {
		log.Fatalf("Failed os.Mkdir(): %s", err)
	}
	log.Printf("Test datadir is: %s\n", dir)
	return
}

func TestPingHandlerGET(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	time := time.Now()
	uri := fmt.Sprintf("http://foo.example.com/ping?q=%s", time.String())
	r, _ := http.NewRequest("GET", uri, nil)
	w := httptest.NewRecorder()

	d.pingHandler(w, r)

	if w.Body.String() != uri {
		t.Errorf("expected %q but instead got %q", uri, w.Body.String())
	}
}

func TestPingHandlerPOST(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	time := time.Now()
	r, _ := http.NewRequest("POST", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	d.pingHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func TestPingHandlerPUT(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	time := time.Now()
	r, _ := http.NewRequest("PUT", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	d.pingHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func TestPingHandlerInvalid(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	time := time.Now()
	r, _ := http.NewRequest("DELETE", "", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	d.pingHandler(w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}

func TestViewHandler(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	testdata := "http://foo.example.com/nada"
	r, _ := http.NewRequest("GET", testdata, nil)
	w := httptest.NewRecorder()

	d.viewHandler(w, r)

	if w.Body.String() != testdata {
		t.Errorf("expected %q but instead got %q", testdata, w.Body.String())
	}
}

func TestQueueHandlerInvalidGET(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	r, _ := http.NewRequest("GET", "http://foo.example.com/queue/invalid", strings.NewReader(""))
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func TestQueueHandlerInvalidPUT(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/invalid", strings.NewReader(""))
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 503 {
		t.Errorf("Expected: 503 Got: %d", w.Code)
	}
}

func TestQueueHandlerPUT(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/testing", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}
}

func TestQueueHandlerGET(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	r, _ := http.NewRequest("GET", "http://foo.example.com/queue/testing", nil)
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 404 {
		t.Errorf("Expected: 404 Got: %d", w.Code)
	}
}

func TestQueueHandlerGET2(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	time := time.Now()
	r, _ := http.NewRequest("PUT", "http://foo.example.com/queue/testing", strings.NewReader(time.String()))
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected: 200 Got: %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "http://foo.example.com/queue/testing", nil)
	w = httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Body.String() != time.String() {
		t.Errorf("expected %q but instead got %q", time.String(), w.Body.String())
	}
}

func TestQueueHandlerInvalidMethod(t *testing.T) {
	var d daemon
	d.datadir = datadirTesting()

	r, _ := http.NewRequest("POST", "http://foo.example.com/queue/", strings.NewReader("PAYLOAD"))
	w := httptest.NewRecorder()

	d.queueHandler(w, r)

	if w.Code != 405 {
		t.Errorf("Expected: 405 Got: %d", w.Code)
	}
}
