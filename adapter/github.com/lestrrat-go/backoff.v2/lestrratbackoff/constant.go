package lestrratbackoff

import (
	"context"

	"github.com/lestrrat-go/backoff/v2"
	"github.com/toga4/go-retryabletransport"
)

type constantPolicy struct {
	policy backoff.Policy
}

func NewConstantPolicy(options ...backoff.Option) retryabletransport.BackoffPolicy {
	return &constantPolicy{backoff.Constant(options...)}
}

func (p *constantPolicy) New(ctx context.Context) retryabletransport.Backoff {
	return &adapter{p.policy.Start(ctx)}
}
