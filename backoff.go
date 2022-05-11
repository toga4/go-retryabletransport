package retryabletransport

import (
	"time"

	"github.com/googleapis/gax-go/v2"
)

type Backoff interface {
	Pause() time.Duration
}

type BackoffConfig interface {
	New() Backoff
}

// gax-go v2 の実装を利用したバックオフアルゴリズムの実装
type GaxBackoffConfig struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
}

// Ensure at compile time that GaxBackoffConfig implements BackoffConfig.
var _ BackoffConfig = (*GaxBackoffConfig)(nil)

func (c *GaxBackoffConfig) New() Backoff {
	return &gax.Backoff{
		Initial:    c.Initial,
		Max:        c.Max,
		Multiplier: c.Multiplier,
	}
}
