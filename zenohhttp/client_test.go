package zenohhttp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/demo/**", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Sample{{Key: "demo/**", Value: json.RawMessage(`"v"`), Encoding: "text/plain"}})
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(ts.URL)
	s, err := c.Get(context.Background(), "demo/**")
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 1 || string(s[0].Value) != `"v"` {
		t.Fatalf("unexpected: %#v", s)
	}
}

func TestPutSetsContentType(t *testing.T) {
	var gotCT string
	mux := http.NewServeMux()
	mux.HandleFunc("/k", func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(204)
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(ts.URL)
	if err := c.Put(context.Background(), "k", []byte("x"), "text/plain"); err != nil {
		t.Fatal(err)
	}
	if gotCT != "text/plain" {
		t.Fatalf("want text/plain, got %q", gotCT)
	}
}

func TestDeleteOK(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/rm", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(ts.URL)
	if err := c.Delete(context.Background(), "rm"); err != nil {
		t.Fatal(err)
	}
}

func TestHTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/bad", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("nope"))
	})
	ts := httptest.NewServer(mux)
	defer ts.Close()

	c := New(ts.URL)
	_, err := c.Get(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("unexpected error: %v", err)
	}
}
