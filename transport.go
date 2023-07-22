package retryabletransport

import (
	"bytes"
	"io"
	"net/http"
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
func (t *Transport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	ctx := req.Context()

	// Since the request body cannot be read more than once, read the entire request body in case a retry is necessary.
	// A less efficient but simpler implementation.
	buf := []byte{}
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		buf = b
	}

	backoff := t.backoffPolicy.New(ctx)

	for backoff.Continue() {

		if res != nil {
			// If this is the 2nd or subsequent run, and a response has been received on the previous call,
			// discard and close the response body to reuse HTTP connections.
			_, _ = io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}

		req.Body = io.NopCloser(bytes.NewReader(buf))

		transport := t.transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		res, err = transport.RoundTrip(req)

		if err != nil {
			if !t.shouldRetryError(req, err) {
				return
			}
		} else {
			if !t.shouldRetryResponse(res) {
				return
			}
		}
	}

	// If context canceled or deadline exceeded while waiting to backoff timeout, return context error.
	// Else returns the result of the last attempt.
	// For example, when the maximum number of attempts has been reached.
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return
}
