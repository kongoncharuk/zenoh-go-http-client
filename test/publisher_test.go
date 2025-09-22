package test

import (
	"context"
	"testing"

	"github.com/kongoncharuk/zenoh-go-http-client/zenohhttp"
)

func TestPublishing(t *testing.T) {
	c := zenohhttp.New("http://localhost:8000")
	ctx, subCancel := context.WithCancel(context.Background())
	defer subCancel()
	if err := c.Put(ctx, "demo/example/hello", []byte("hi"), "text/plain"); err != nil {
		t.Errorf("Failed to publish message: %v", err)
	}
}
