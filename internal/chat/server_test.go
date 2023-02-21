package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var (
	server  *Server
	hs      *httptest.Server
	testURL string
)

func setupTestCase(t *testing.T) func(t *testing.T) {
	server = NewServer()
	go server.Start()
	hs = httptest.NewServer(http.HandlerFunc(server.Handler))
	testURL = "ws" + strings.TrimPrefix(hs.URL, "http")
	return func(t *testing.T) {
		defer hs.Close()
		server.Stop()
		server = nil
		testURL = ""
	}
}

func setupUser(t *testing.T) func() {
	ws, _, err := websocket.DefaultDialer.Dial(testURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	return func() {
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.Close()
	}
}

func setupSupportUser(t *testing.T) func() {
	header := http.Header{}
	header.Set("type", "S")
	ws, _, err := websocket.DefaultDialer.Dial(testURL, header)
	if err != nil {
		t.Fatal(err)
	}
	return func() {
		ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		ws.Close()
	}
}

func TestServerSingleClient(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownUser := setupUser(t)
	defer teardownUser()

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	want := 1
	if len(server.queue) != 1 {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}

func TestServerMultipleClients(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for i := 0; i < 10; i++ {
		teardownUser := setupUser(t)
		defer teardownUser()
	}

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	want := 10
	if len(server.queue) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}

func TestServerSingleSupportClients(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSupportUser := setupSupportUser(t)
	defer teardownSupportUser()

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	want := 1
	if len(server.workers) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}

func TestMultipleSupportClients(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for i := 0; i < 10; i++ {
		teardownSupportUser := setupSupportUser(t)
		defer teardownSupportUser()
	}

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	want := 10
	if len(server.workers) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}

func TestSupportClientAndClient(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSupportUser := setupSupportUser(t)
	defer teardownSupportUser()
	teardownUser := setupUser(t)
	defer teardownUser()

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	supports := []*user{}
	for k, v := range server.workers {
		supports = append(supports, k)
		if v == nil {
			t.Fatal("no client")
		}
	}

	want1, want2 := 1, true
	if len(supports) != want1 && server.workers[supports[0]] != nil {
		t.Errorf("got %d and %t; want %d and %t", len(supports), server.workers[supports[0]] != nil, want1, want2)
	}
}

func TestMultipleSupportClientsAndClients(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	for i := 0; i < 5; i++ {
		teardownUser := setupUser(t)
		defer teardownUser()
	}

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()

	want := 5
	if len(server.queue) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
	server.mu.Unlock()

	for i := 0; i < 5; i++ {
		teardownSupportUser := setupSupportUser(t)
		defer teardownSupportUser()
	}

	time.Sleep(50 * time.Millisecond)
	server.mu.Lock()
	defer server.mu.Unlock()

	want = 0
	if len(server.queue) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}

func TestConnectNextUserFromQueue(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	teardownSupportUser := setupSupportUser(t)
	defer teardownSupportUser()

	for i := 0; i < 2; i++ {
		teardownUser := setupUser(t)
		defer teardownUser()
	}

	time.Sleep(50 * time.Millisecond)

	server.mu.Lock()
	want := 1
	if len(server.queue) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}

	for _, v := range server.workers {
		server.close <- v.id
	}
	server.mu.Unlock()

	time.Sleep(50 * time.Millisecond)

	server.mu.Lock()
	defer server.mu.Unlock()
	want = 0
	if len(server.queue) != want {
		t.Errorf("got %d; want %d", len(server.queue), want)
	}
}
