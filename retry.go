// Package retry provides a stateless method for implementing
// exponential and other backoff strategies.
//
// The retry package defines a Strategy as a function that maps an
// integer to a time.Duration. Users of the retry package are expected
// to maintain a counter variable that represents the number of retries
// of a given operation.  This counter is then mapped through one of
// the provided or user-created backoff strategies to produce a duration
// of time for the time.Sleep function.  The retry package is inspired
// by the stateless backoff techniques described in
// http://blog.gopheracademy.com/advent-2014/backoff/
//
// Complex backoff strategies can be built by using the methods defined
// in the retry package to add random splay, overwrite, or otherwise
// manipulate the values returned by a Strategy.
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
func Exponential(max time.Duration) Strategy {
	if max < 0 {
		max = math.MaxInt64
	}
	return func(nth int) time.Duration {
		x := 1
		for i := 0; i < nth; i++ {
			x *= 2
		}
		val := time.Second * time.Duration(x)
		if val < 0 || val > max {
			val = max
		}
		return val
	}
}

// Fixed creates a backoff policy that selects the nth duration in the
// argument list. If the retry counter is greater than the number of
// items provided, the final item is returned.  If the retry counter
// is less than 0 the first item is returned.  If the parameter list
// is empty, the returned strategy will always return 0.
func Fixed(dur ...time.Duration) Strategy {
	if len(dur) == 0 {
		return func(int) time.Duration { return 0 }
	}
	return func(nth int) time.Duration {
		if nth < 0 {
			nth = 0
		}
		if nth < len(dur) {
			return dur[nth]
		}
		return dur[len(dur)-1]
	}
}

// Milliseconds creates a backoff policy that selects the nth item in
// the array, multiplied by time.Millisecond. If the retry counter is
// greater than the number of items provided, the final item is returned.
func Milliseconds(ms ...int) Strategy {
	if len(ms) == 0 {
		return func(int) time.Duration { return 0 }
	}
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
// array, multiplied by time.Second.
func Seconds(secs ...int) Strategy {
	if len(secs) == 0 {
		return func(int) time.Duration { return 0 }
	}
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

// Splay adds a random duration in the range ±duration to values
// returned by a Strategy. Splay is useful for avoiding "thundering
// herd" scenarios, where multiple processes become inadvertently
// synchronized and use the same backoff strategy to use a shared
// service.
func (base Strategy) Splay(d time.Duration) Strategy {
	return func(retry int) time.Duration {
		jitter := time.Duration(randomsrc.Int63n(int64(d)))
		if randomsrc.Int()%2 == 0 {
			jitter = -jitter
		}
		val := base(retry)
		// avoid integer overflow
		if jitter > 0 && val > math.MaxInt64-jitter {
			jitter = -jitter
		} else if val < 0 && jitter < 0 && val < math.MinInt64-jitter {
			jitter = -jitter
		}
		return val + jitter
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

// Prepend displaces the first len(dur) mappings of a Strategy, selecting
// durations from the given parameter list instead. Passing len(dur)
// to the returned strategy is equivalent to passing 0 to the original
// strategy.
func (base Strategy) Prepend(dur ...time.Duration) Strategy {
	return func(nth int) time.Duration {
		if nth < 0 {
			nth = 0
		}
		if nth < len(dur) {
			return dur[nth]
		}
		return base(nth - len(dur))
	}
}

// Overwrite replaces the first len(dur) mappings of a Strategy, selecting
// durations from the given parameter list instead. Passing len(dur) to
// the returned strategy is equivalent to passing len(dur) to the original
// strategy.
func (base Strategy) Overwrite(dur ...time.Duration) Strategy {
	return func(nth int) time.Duration {
		if nth < 0 {
			nth = 0
		}
		if nth < len(dur) {
			return dur[nth]
		}
		return base(nth)
	}
}
