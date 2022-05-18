package lestrratbackoff

import (
	"github.com/lestrrat-go/backoff/v2"
)

type adapter struct {
	controller backoff.Controller
}

func (a *adapter) Continue() bool {
	return backoff.Continue(a.controller)
}
