package retry

import (
	"testing"
	"time"
)

func TestSplay(t *testing.T) {
	const s = 5
	backoff := Intervals(0)
	for i := 0; i < 1e5; i++ {
		b := backoff(0)
		if b < -s || b > +s {
			t.Errorf("Splay should be in the range [0, ±%d), got %d", s, b)
		}
	}
}

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

func TestFibonacciBackoff(t *testing.T) {
	ans := []time.Duration{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}
	backoff := Fibonacci(1)
	for i, v := range ans {
		if x := backoff(i); x != v {
			t.Errorf("fibonacci backoff(%d) = %s, should be %s", i, x, v)
		} else {
			t.Logf("fibonacci backoff(%d) = %s", i, x)
		}
	}
	if backoff(-1) != 0 {
		t.Errorf("fibonacci backoff(-1) = %s, should be 0ns", backoff(-1))
	}
}
		
func TestIntervalBackoff(t *testing.T) {
	ans := []time.Duration{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	backoff := Intervals(ans...)

	for i, v := range ans {
		if x := backoff(i); x != v {
			t.Errorf("fixed backoff(%d) = %s, should be %s", i, x, v)
		} else {
			t.Logf("fixed backoff(%d) = %s", i, x)
		}
	}
}
