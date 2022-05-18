package lestrratbackoff

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/backoff/v2"
)

func Test_ConstantPolicy(t *testing.T) {
	ctx := context.Background()

	backoffPolicy := NewConstantPolicy(
		backoff.WithInterval(20*time.Millisecond),
		backoff.WithMaxRetries(3),
	)

	b := backoffPolicy.New(ctx)

	start := time.Now()

	count := 0
	for b.Continue() {
		count++
	}

	durationMillis := time.Now().Sub(start)

	if count != 4 {
		t.Errorf("expected count is equal to 4 but %v", count)
	}
	min := (60 - 5) * time.Millisecond
	max := (60 + 5) * time.Millisecond
	if durationMillis < min || durationMillis > max {
		t.Errorf("expected duration is between %v and %v but %v", min, max, durationMillis)
	}
}
