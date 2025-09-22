package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kongoncharuk/zenoh-go-http-client/zenohhttp"
)

func TestSubscribing(t *testing.T) {
	c := zenohhttp.New("http://localhost:8000")
	subCtx, subCancel := context.WithCancel(context.Background())
	defer subCancel()

	done := make(chan struct{})
	_, err := c.Subscribe(subCtx, "demo/example/**", func(s zenohhttp.Sample) {
		fmt.Printf("[SUB] %s (%s) %s\n", s.Key, s.Encoding, string(s.Value))
		t.Log("Received a message:", string(s.Value))
		// Signal that the message was received
		subCancel()
		close(done)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Wait for a message or timeout
	select {
	case <-done:
		// Message received; subscription will be cancelled by deferred cancel
	case <-time.After(30 * time.Second):
		t.Fatal("Timeout waiting for SSE message")
	}
}
