package gaxbackoff

import (
	"time"

	"github.com/googleapis/gax-go/v2"
	"github.com/toga4/go-retryabletransport"
)

type BackoffPolicy struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
}

var _ retryabletransport.BackoffPolicy = (*BackoffPolicy)(nil)

func (p *BackoffPolicy) New() retryabletransport.Backoff {
	return &gax.Backoff{
		Initial:    p.Initial,
		Max:        p.Max,
		Multiplier: p.Multiplier,
	}
}
