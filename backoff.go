package retryabletransport

import (
	"time"
)

type Backoff interface {
	Pause() time.Duration
}

type BackoffPolicy interface {
	New() Backoff
}
