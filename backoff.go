package retryabletransport

import (
	"context"
)

// Backoff is an interface that generates next backoff interval.
type Backoff interface {
	// Continue returns when to run the next backoff. The 1st call should return
	// true immediately. The next and subsequent calls should return whether or
	// not the next should be run after a backoff timeout.
	Continue() bool
}

// BackoffPolicy is an interface that generates new Backoff instance.
type BackoffPolicy interface {
	// New returns new Backoff instance.
	New(ctx context.Context) Backoff
}
