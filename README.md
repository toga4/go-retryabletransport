# go-retryabletransport
[![Go Reference](https://pkg.go.dev/badge/github.com/toga4/go-retryabletransport.svg)](https://pkg.go.dev/github.com/toga4/go-retryabletransport) [![Test](https://github.com/toga4/go-retryabletransport/actions/workflows/ci.yaml/badge.svg)](https://github.com/toga4/go-retryabletransport/actions/workflows/ci.yaml)

This library provides an implementation of http.RoundTripper for retrying HTTP requests with backoff. To configure backoff, you can use with 3rd-party libraries, or implement the Backoff and BackoffPolicy interfaces.

## Usage

```go
package main

import (
	"errors"
	"net/http"
	"syscall"
	"time"

	"github.com/lestrrat-go/backoff/v2"
	"github.com/toga4/go-retryabletransport"
	"github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff"
)

func Example_Exponential() {
	backoffPolicy := lestrratbackoff.NewExponentialPolicy(
		backoff.WithMinInterval(300*time.Millisecond),
		backoff.WithMaxInterval(2*time.Second),
		backoff.WithJitterFactor(0.05),
	)

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 20

	retryableTransport := retryabletransport.New(
		backoffPolicy,
		retryabletransport.WithTransport(transport),
		retryabletransport.WithShouldRetryError(func(r *http.Request, err error) bool {
			return errors.Is(err, syscall.ECONNRESET)
		}),
		retryabletransport.WithShouldRetryResponse(func(r *http.Response) bool {
			return r.StatusCode == http.StatusBadGateway
		}),
	)

	client := http.Client{
		Transport: retryableTransport,
		Timeout:   5 * time.Second,
	}

	_, _ = client.Get("http://example.com")
}
```

## Backoff

Some implementation for using 3rd-party backoff libraries are provided. See [adapter](./adapter).

## Caveats

- No limit on the length of retries. Instead, use `http.Client.Timeout` or `context.WithTimeout` to limit the length of retries.
