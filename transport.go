package retryabletransport

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type ShouldRetryErrorFunc func(*http.Request, error) bool
type ShouldRetryResponseFunc func(*http.Response) bool

var (
	defaultShouldRetryError   ShouldRetryErrorFunc    = func(r *http.Request, err error) bool { return false }
	defaultShouldRetrResponse ShouldRetryResponseFunc = func(r *http.Response) bool { return false }
)

type Transport struct {
	backoffPolicy       BackoffPolicy
	shouldRetryError    ShouldRetryErrorFunc
	shouldRetryResponse ShouldRetryResponseFunc
	transport           http.RoundTripper
}

// Ensure at compile time that Transport implements http.RoundTripper.
var _ http.RoundTripper = (*Transport)(nil)

// New creates a http.RoundTripper with retry.
func New(backoffPolicy BackoffPolicy, options ...TransportOption) *Transport {
	shouldRetryError := defaultShouldRetryError
	shouldRetryResponse := defaultShouldRetrResponse
	var transport http.RoundTripper

	for _, option := range options {
		switch option.Ident() {
		case identShouldRetryError{}:
			shouldRetryError = option.Value().(ShouldRetryErrorFunc)
		case identShouldRetryResponse{}:
			shouldRetryResponse = option.Value().(ShouldRetryResponseFunc)
		case identTransport{}:
			transport = option.Value().(http.RoundTripper)
		}
	}

	return &Transport{
		backoffPolicy:       backoffPolicy,
		shouldRetryError:    shouldRetryError,
		shouldRetryResponse: shouldRetryResponse,
		transport:           transport,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Since the request body cannot be read more than once, read the entire request body in case a retry is necessary.
	// A less efficient but simpler implementation.
	buf := []byte{}
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		buf = b
	}

	backoff := t.backoffPolicy.New()

	for {
		req.Body = ioutil.NopCloser(bytes.NewReader(buf))

		transport := t.transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		res, err := transport.RoundTrip(req)

		if err != nil {
			if !t.shouldRetryError(req, err) {
				return nil, err
			}
		} else {
			if !t.shouldRetryResponse(res) {
				return res, nil
			}
		}

		if res != nil {
			// Discards response body to reuse HTTP connections.
			_, _ = io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		}

		wait := backoff.Pause()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
}
