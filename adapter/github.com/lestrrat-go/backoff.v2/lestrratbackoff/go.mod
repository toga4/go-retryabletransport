module github.com/toga4/go-retryabletransport/adapter/github.com/lestrrat-go/backoff.v2/lestrratbackoff

go 1.20

require (
	github.com/lestrrat-go/backoff/v2 v2.0.8
	github.com/toga4/go-retryabletransport v0.3.0
)

require github.com/lestrrat-go/option v1.0.0 // indirect

replace github.com/toga4/go-retryabletransport => ../../../../..
