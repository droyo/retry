// Package retry provides a stateless method for implementing
// exponential and other backoff policies.
//
// Users of the backoff package are expected to maintain a counter
// variable that represents the number of retries of a given operation.
// This counter is then mapped through one of the provided or user-created
// backoff strategies to produce a duration of time for the time.Sleep
// function.  The retry package is inspired by the stateless backoff
// techniques described in http://blog.gopheracademy.com/advent-2014/backoff/
package retry // import "aqwari.net/retry"

import (
	cryptrand "crypto/rand"
	"encoding/binary"
	"math"
	"math/rand"
	"time"
)

var randomsrc *rand.Rand

func init() {
	var seed int64
	if err := binary.Read(cryptrand.Reader, binary.BigEndian, &seed); err != nil {
		panic("backoff: failed to seed RNG: " + err.Error())
	}
	randomsrc = rand.New(rand.NewSource(seed))
}

// A Strategy is a mapping from a retry counter to a duration of time.
// The retry package provides a number of built-in Strategies that
// capture the most common use cases.
type Strategy func(nth int) time.Duration

// Exponential creates an exponential backoff Strategy that returns
// the lesser of 2ⁿ or max seconds. If max is negative, the values
// returned by the Strategy will continue increasing to the maximum
// value of a time.Duration (about 290 years)
func Exponential(max int) Strategy {
	ceil := time.Second * time.Duration(max)
	if max < 0 {
		ceil = time.Duration(math.MaxInt64)
	}
	return func(nth int) time.Duration {
		x := 1
		for i := 0; i < nth; i++ {
			x *= 2
		}
		val := time.Second * time.Duration(x)
		if val < 0 || val > ceil {
			val = ceil
		}
		return val
	}
}

// Milliseconds creates a backoff policy that selects the nth item in the
// array, multiplied by time.Millisecond. If the retry counter is greater
// than the number of items provided, the final item is returned.
func Milliseconds(ms ...int) Strategy {
	return func(nth int) time.Duration {
		if nth < 0 {
			nth = 0
		}
		if nth < len(ms) {
			return time.Millisecond * time.Duration(ms[nth])
		}
		return time.Millisecond * time.Duration(ms[len(ms)-1])
	}
}

// Seconds creates a backoff policy that selects the nth item in the
// array, multiplied by time.Second. If the retry counter is greater
// than the number of items provided, the final item is returned.
func Seconds(secs ...int) Strategy {
	return func(nth int) time.Duration {
		if nth < 0 {
			nth = 0
		}
		if nth < len(secs) {
			return time.Second * time.Duration(secs[nth])
		}
		return time.Second * time.Duration(secs[len(secs)-1])
	}
}

// Splay adds a random duration in the range ±seconds to values returned
// by a Strategy. Splay is useful for avoiding "thundering herd"
// scenarios, where multiple processes become inadvertently synchronized
// and use the same backoff strategy to use a shared service.
func (base Strategy) Splay(seconds float64) Strategy {
	return func(retry int) time.Duration {
		jitter := time.Duration(float64(time.Second) * seconds)
		val := base(retry)
		// avoid integer overflow
		if jitter > 0 && val > math.MaxInt64-jitter {
			jitter = -jitter
		} else if val < 0 && jitter < 0 && val < math.MinInt64-jitter {
			jitter = -jitter
		}
		return base(retry) + jitter
	}
}

// Scale scales a backoff policy by a constant number of seconds. The
// returned Policy will return values from the policy, uniformly
// multiplied by secs.
func (base Strategy) Scale(seconds float64) Strategy {
	return func(retry int) time.Duration {
		x := base(retry).Seconds() * seconds
		return time.Duration(math.Floor(float64(time.Second) * x))
	}
}
