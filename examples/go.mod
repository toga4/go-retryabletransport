module github.com/toga4/go-retryabletransport/examples

go 1.16

replace (
	github.com/toga4/go-retryabletransport => ../
	github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff => ../adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff
)

require (
	github.com/lestrrat-go/backoff/v2 v2.0.8
	github.com/toga4/go-retryabletransport v0.0.0-00010101000000-000000000000
	github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff v0.0.0-00010101000000-000000000000
)
