package lestrratbackoff

import (
	"time"

	"github.com/lestrrat-go/backoff/v2"
)

type adapter struct {
	backoff backoff.IntervalGenerator
}

func (a *adapter) Pause() time.Duration {
	return a.backoff.Next()
}
