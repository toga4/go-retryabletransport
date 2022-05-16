package lestrratbackoff

import (
	"testing"
	"time"

	"github.com/lestrrat-go/backoff/v2"
)

func Test_ConstantPolicy(t *testing.T) {
	backoffPolicy := NewConstantPolicy(
		backoff.WithInterval(1000*time.Millisecond),
		backoff.WithJitterFactor(0.1),
	)

	min := 900 * time.Millisecond
	max := 1100 * time.Millisecond

	b := backoffPolicy.New()

	for i := 0; i < 100_000; i++ {
		val := b.Pause()
		if val < min {
			t.Errorf("expected %v to be greater than %v", val, min)
		}
		if val > max {
			t.Errorf("expected %v to be less than %v", val, max)
		}
	}
}
