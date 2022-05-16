package lestrratbackoff

import (
	"github.com/lestrrat-go/backoff/v2"
	"github.com/toga4/go-retryabletransport"
)

type exponentialPolicy struct {
	options []backoff.ExponentialOption
}

func NewExponentialPolicy(options ...backoff.ExponentialOption) retryabletransport.BackoffPolicy {
	return &exponentialPolicy{options}
}

func (p *exponentialPolicy) New() retryabletransport.Backoff {
	return &adapter{backoff.NewExponentialInterval(p.options...)}
}
