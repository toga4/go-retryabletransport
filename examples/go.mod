module github.com/toga4/go-retryabletransport/examples

go 1.16

require (
	github.com/lestrrat-go/backoff/v2 v2.0.8
	github.com/toga4/go-retryabletransport v0.2.0
	github.com/toga4/go-retryabletransport/adapter/github.com/googleapis/gax-go.v2/gaxbackoff v0.2.1
	github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff v0.2.1
)

replace (
	github.com/toga4/go-retryabletransport => ../
	github.com/toga4/go-retryabletransport/adapter/github.com/googleapis/gax-go.v2/gaxbackoff => ../adapter/github.com/googleapis/gax-go.v2/gaxbackoff
	github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff => ../adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff
)
