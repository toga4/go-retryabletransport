package gaxbackoff

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	ctx := context.Background()

	initialBackoff := 20 * time.Millisecond
	maxBackoff := 50 * time.Millisecond
	backoffMultiplier := 2.0

	backoffPolicy := &BackoffPolicy{
		Initial:    initialBackoff,
		Max:        maxBackoff,
		Multiplier: backoffMultiplier,
	}

	ctx, cancel := context.WithTimeout(ctx, 110*time.Millisecond)
	defer cancel()

	b := backoffPolicy.New(ctx)
	count := 0
	for b.Continue() {
		log.Println(count)
		count++
	}

	if count < 4 {
		t.Errorf("expected count is greater than or equal to 4 but %v", count)
	}
}
