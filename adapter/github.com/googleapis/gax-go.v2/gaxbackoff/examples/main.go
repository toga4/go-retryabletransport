package main

import (
	"errors"
	"net/http"
	"syscall"
	"time"

	"github.com/toga4/go-retryabletransport"
	"github.com/toga4/go-retryabletransport/adapter/github.com/googleapis/gax-go.v2/gaxbackoff"
)

func Example() {
	backoffPolicy := &gaxbackoff.BackoffPolicy{
		Initial:    300 * time.Millisecond,
		Max:        2 * time.Second,
		Multiplier: 1.5,
	}

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
