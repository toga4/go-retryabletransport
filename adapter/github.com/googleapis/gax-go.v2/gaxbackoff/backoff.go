package gaxbackoff

import (
	"context"
	"sync"
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

type backoff struct {
	ctx     context.Context
	first   bool
	backoff *gax.Backoff
	mu      sync.Mutex
}

func (p *BackoffPolicy) New(ctx context.Context) retryabletransport.Backoff {
	return &backoff{
		ctx:   ctx,
		first: true,
		backoff: &gax.Backoff{
			Initial:    p.Initial,
			Max:        p.Max,
			Multiplier: p.Multiplier,
		},
	}
}

func (b *backoff) Continue() bool {
	if b.isFirst() {
		return true
	}

	c := time.After(b.backoff.Pause())
	select {
	case <-b.ctx.Done():
		return false
	case <-c:
		return true
	}
}

func (b *backoff) isFirst() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	f := b.first
	b.first = false
	return f
}
