package server

import (
	"testing"
	"time"
)

func TestTickerStopTerminatesGoroutine(t *testing.T) {
	timer := &Timer{FocusTime: 1, Pace: 1000}
	h := &Client{Id: "1", Username: "test", MsgCh: make(chan ServerPacket, 1)}
	room := &Room{
		Id:      "room1",
		Hikers:  map[string]*Client{"1": h},
		Session: &Session{},
		Timer:   timer,
		Host:    "1",
	}

	timer.BeginFocusTime(room)
	// allow goroutine to start
	time.Sleep(10 * time.Millisecond)
	timer.StopTicker()

	done := make(chan struct{})
	go func() {
		timer.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("update goroutine did not terminate")
	}
}
