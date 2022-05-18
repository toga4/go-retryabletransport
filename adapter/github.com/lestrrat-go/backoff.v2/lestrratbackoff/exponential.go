package lestrratbackoff

import (
	"context"

	"github.com/lestrrat-go/backoff/v2"
	"github.com/toga4/go-retryabletransport"
)

type exponentialPolicy struct {
	policy backoff.Policy
}

func NewExponentialPolicy(options ...backoff.ExponentialOption) retryabletransport.BackoffPolicy {
	return &exponentialPolicy{backoff.Exponential(options...)}
}

func (p *exponentialPolicy) New(ctx context.Context) retryabletransport.Backoff {
	return &adapter{p.policy.Start(ctx)}
}
