package zenohhttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/r3labs/sse/v2"
)

// Sample mirrors a zenoh REST sample.
// Value is json.RawMessage to avoid forcing a schema.
type Sample struct {
	Key      string          `json:"key"`
	Value    json.RawMessage `json:"value"`
	Encoding string          `json:"encoding,omitempty"`
	Time     string          `json:"time,omitempty"`
}

type HTTPError struct {
	Op         string
	URL        string
	Status     string
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return e.Op + " " + e.URL + " failed: " + e.Status + " " + e.Body
}

// Client HTTP/SSE client for the zenoh REST plugin.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	// default headers for all requests (not for SSE Accept)
	headers map[string]string
}

func New(base string, opts ...Option) *Client {
	c := &Client{
		BaseURL: strings.TrimRight(base, "/"),
		HTTP:    &http.Client{Timeout: 0},
		headers: map[string]string{},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type Option func(*Client)

// Get performs GET /<selector> returning samples.
func (c *Client) Get(ctx context.Context, selector string) ([]Sample, error) {
	if c == nil || c.HTTP == nil {
		return nil, errors.New("nil client")
	}
	url := c.join(selector)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
		return nil, &HTTPError{Op: "GET", URL: url, Status: res.Status, StatusCode: res.StatusCode, Body: string(b)}
	}
	var out []Sample
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

// Put performs PUT /<keyexpr> with payload.
// Set contentType to specify typed values (e.g., application/json).
func (c *Client) Put(ctx context.Context, keyexpr string, payload []byte, contentType string) error {
	if c == nil || c.HTTP == nil {
		return errors.New("nil client")
	}
	url := c.join(keyexpr)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
		return &HTTPError{Op: "PUT", URL: url, Status: res.Status, StatusCode: res.StatusCode, Body: string(b)}
	}
	return nil
}

// Delete performs DELETE /<keyexpr>.
func (c *Client) Delete(ctx context.Context, keyexpr string) error {
	if c == nil || c.HTTP == nil {
		return errors.New("nil client")
	}
	url := c.join(keyexpr)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 8<<10))
		return &HTTPError{Op: "DELETE", URL: url, Status: res.Status, StatusCode: res.StatusCode, Body: string(b)}
	}
	return nil
}

// Subscribe opens an SSE stream on GET /<keyexpr> using a context.
// Cancelling the context stops the subscription.
func (c *Client) Subscribe(ctx context.Context, keyExpr string, onSample func(Sample)) (func() error, error) {
	log.Println("Subscribing to", keyExpr)
	if c == nil || c.HTTP == nil {
		return nil, errors.New("no http client provided")
	}
	url := c.join(keyExpr)

	client := sse.NewClient(url)
	client.Connection = c.HTTP
	client.Headers = map[string]string{"Accept": "text/event-stream"}
	for k, v := range c.headers {
		client.Headers[k] = v
	}

	// SubscribeWithContext ties the subscription to the provided context.
	err := client.SubscribeWithContext(ctx, "", func(ev *sse.Event) {
		log.Println("[SSE] event:", string(ev.Event), "id:", string(ev.ID), "data:", string(ev.Data))
		if len(ev.Data) > 0 && onSample != nil {
			var s Sample
			if err := json.Unmarshal(ev.Data, &s); err != nil {
				log.Println("Unable to unmarshal event:", err)
				return
			}
			onSample(s)
		}
	})
	if err != nil {
		return nil, err
	}

	// Return a no-op cancel function; the subscription is cancelled by cancelling the context.
	cancel := func() error { return nil }
	return cancel, nil
}

// join concatenates BaseURL and key/selector without cleaning '*' or '**'.
func (c *Client) join(suffix string) string {
	return c.BaseURL + "/" + strings.TrimLeft(suffix, "/")
}
