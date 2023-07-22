package lestrratbackoff

import (
	"context"
	"testing"
	"time"

	"github.com/lestrrat-go/backoff/v2"
)

func Test_ExponentialPolicy(t *testing.T) {
	ctx := context.Background()

	backoffPolicy := NewExponentialPolicy(
		backoff.WithMinInterval(20*time.Millisecond),
		backoff.WithMaxInterval(50*time.Millisecond),
		backoff.WithMultiplier(2.0),
		backoff.WithMaxRetries(3),
	)

	b := backoffPolicy.New(ctx)

	start := time.Now()

	count := 0
	for b.Continue() {
		count++
	}

	duration := time.Since(start)

	if count != 4 {
		t.Errorf("expected count is equal to 4 but %v", count)
	}
	min := (110 - 5) * time.Millisecond
	max := (110 + 5) * time.Millisecond
	if duration < min || duration > max {
		t.Errorf("expected duration is between %v and %v but %v", min, max, duration)
	}
}
