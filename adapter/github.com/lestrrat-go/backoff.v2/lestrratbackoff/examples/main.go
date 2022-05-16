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

func Example_Constant() {
	backoffPolicy := lestrratbackoff.NewConstantPolicy(
		backoff.WithInterval(300*time.Millisecond),
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
