package retry

import (
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	var next, prev time.Duration
	var bk = Exponential(time.Second)

	for i := 0; i < 100; i++ {
		next = bk(i)
		t.Logf("exponential backoff(%d) = %s", i, next)
		if next == prev {
			break
		}
		if next <= prev {
			t.Errorf("bk(%d)[%s] <= bk(%d)[%s], exponential backoff should always increase",
				i, next, i-1, prev)
		}
		prev = next
	}
}

func TestIntervalBackoff(t *testing.T) {
	x := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	const jitter = 50
	backoff := Seconds(x...)

	for i, v := range x {
		t.Logf("fixed backoff(%d) = %s", i, backoff(i))
		if backoff(i) != time.Duration(v)*time.Second {
			t.Errorf("fixed backoff(%d) = %s, should be > %s", i, backoff(i),
				time.Duration(v)*time.Second)
		}
	}
}
