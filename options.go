package retryabletransport

import (
	"net/http"

	"github.com/toga4/go-retryabletransport/internal/option"
)

type TransportOption interface {
	option.Option
}

type transportOption struct {
	option.Option
}

type identShouldRetryError struct{}
type identShouldRetryResponse struct{}
type identTransport struct{}

func WithShouldRetryError(f ShouldRetryErrorFunc) TransportOption {
	return &transportOption{option.New(identShouldRetryError{}, f)}
}

func WithShouldRetryResponse(f ShouldRetryResponseFunc) TransportOption {
	return &transportOption{option.New(identShouldRetryResponse{}, f)}
}

func WithTransport(rt http.RoundTripper) TransportOption {
	return &transportOption{option.New(identTransport{}, rt)}
}
