package retryabletransport

import (
	"time"
)

// Backoff is an interface that generates next backoff interval.
type Backoff interface {
	// Pause returns next backoff interval.
	Pause() time.Duration
}

// BackoffPolicy is an interface that generates new Backoff instance.
type BackoffPolicy interface {
	// New returns new Backoff instance.
	New() Backoff
}
