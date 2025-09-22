# zenoh-go-http-client

Go wrapper for [Apache Zenoh](https://zenoh.io/) REST plugin (HTTP). Supports:
- Get topic messages (equivalent to `GET /<selector>`)
- Subscribe to live updates with Server‑Sent Events (`GET /<keyexpr>` `Accept: text/event-stream`)
- Publish a message (`PUT /<keyexpr>`)
- Remove messages (`DELETE /<keyexpr>`)

> Requires Zenoh REST plugin to be enabled (see [here](https://zenoh.io/docs/manual/plugin-http/)).

## Install

```bash
go get github.com/kongoncharuk/zenoh-go-http-client@latest
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kongoncharuk/zenoh-go-http-client/zenohhttp"
)

func main() {
    c := zenohhttp.New("http://localhost:8000")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Read messages
    samples, err := c.Get(ctx, "demo/example/**")
    if err != nil {
        panic(err)
    }
    for _, s := range samples {
        fmt.Printf("%s => %s\n", s.Key, string(s.Value))
    }

    // Subscribe to the topic
    subCtx, subCancel := context.WithCancel(context.Background())
    defer subCancel()
    _, err := c.Subscribe(subCtx, "demo/example/**", func(s zenohhttp.Sample) {
        fmt.Printf("[SUB] %s (%s) %s\n", s.Key, s.Encoding, string(s.Value))
    })
    if err != nil {
        panic(err)
    }

	// Send a plain text message
	if err := c.Put(ctx, "demo/example/hello", []byte("hi"), "text/plain"); err != nil {
		panic(err)
	}
}
```

## Notes

- Set `Content-Type` for typed values (e.g., `text/plain`, `application/json`, `application/integer`, `application/float`).
- For subscriptions, this client uses `github.com/r3labs/sse/v2` package.

## License

MIT License. See [LICENSE](./LICENSE) for details.
