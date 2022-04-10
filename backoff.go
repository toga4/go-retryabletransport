package retryabletransport

import (
	"time"

	"github.com/googleapis/gax-go/v2"
)

type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64

	// バックオフアルゴリズムは gax-go v2 の実装を利用する
	backoff *gax.Backoff
}

func (b *Backoff) Pause() time.Duration {
	if b.backoff == nil {
		b.backoff = &gax.Backoff{
			Initial:    b.Initial,
			Max:        b.Max,
			Multiplier: b.Multiplier,
		}
	}
	return b.backoff.Pause()
}
