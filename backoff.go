package retryabletransport

import (
	"time"

	"github.com/googleapis/gax-go/v2"
)

type Backoff interface {
	Pause() time.Duration
}

// バックオフアルゴリズムは gax-go v2 の実装を利用する
type backoff = gax.Backoff

func NewBackoff(initial, max time.Duration, multiplier float64) Backoff {
	return &backoff{
		Initial:    initial,
		Max:        max,
		Multiplier: multiplier,
	}
}
