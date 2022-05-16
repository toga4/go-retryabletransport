package lestrratbackoff

import (
	"github.com/lestrrat-go/backoff/v2"
	"github.com/toga4/go-retryabletransport"
)

type constantPolicy struct {
	options []backoff.ConstantOption
}

func NewConstantPolicy(options ...backoff.ConstantOption) retryabletransport.BackoffPolicy {
	return &constantPolicy{options}
}

func (p *constantPolicy) New() retryabletransport.Backoff {
	return &adapter{backoff.NewConstantInterval(p.options...)}
}
