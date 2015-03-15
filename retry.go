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
// 2ⁿ nanoseconds. The values returned by Exponential will increase
// up to the maximum value of time.Duration and will not overflow.
func Exponential() Strategy {
	return func(retry int) time.Duration {
		if retry < 0 {
			// Can't return a fraction of a nanosecond
			return 0
		}
		var x int64 = 1
		for i := 0; i < retry; i++ {
			if x > math.MaxInt64/2 {
				return math.MaxInt64
			}
			x *= 2
		}
		return time.Duration(x)
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
	if base == nil {
		panic("Splay called on nil Strategy")
	}
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

// Scale scales a backoff policy by a fixed duration. The returned
// Policy will return values from the policy, uniformly multiplied by
// secs.
func (base Strategy) Scale(seconds float64) Strategy {
	if base == nil {
		panic("Scale called on nil Strategy")
	}
	return func(retry int) time.Duration {
		x := base(retry).Seconds() * seconds
		return time.Duration(math.Floor(float64(time.Second) * x))
	}
}

// Units multiplies all values returned by a duration by a fixed
// duration.
func (base Strategy) Units(units time.Duration) Strategy {
	if base == nil {
		panic("Units called on nil Strategy")
	}
	return func(retry int) time.Duration {
		return base(retry) * units
	}
}

// Add adds a fixed duration to every duration returned by a
// Strategy.
func (base Strategy) Add(dur time.Duration) Strategy {
	if base == nil {
		panic("Add called on nil Strategy")
	}
	return func(retry int) time.Duration {
		return base(retry) + dur
	}
}

// Sub subtracts a fixed duration from every duration returned
// by a Strategy.
func (base Strategy) Sub(dur time.Duration) Strategy {
	if base == nil {
		panic("Sub called on nil Strategy")
	}
	return func(retry int) time.Duration {
		return base(retry) + dur
	}
}

// Unshift displaces the first len(dur) mappings of a Strategy, selecting
// durations from the given parameter list instead. Passing len(dur)
// to the returned strategy is equivalent to passing 0 to the original
// strategy.
func (base Strategy) Unshift(dur ...time.Duration) Strategy {
	if base == nil {
		panic("Unshift called on nil Strategy")
	}
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

// Shift skips the first n values of a Strategy. Passing 0+i to the
// returned Strategy is equivalent to passing n+i to the original
// Strategy.
func (base Strategy) Shift(n int) Strategy {
	if base == nil {
		panic("Shift called on nil Strategy")
	}
	return func(retry int) time.Duration {
		return base(retry + n)
	}
}

// The Floor method imposes a minimum value on the durations returned
// by a Strategy. Values returned by the resulting Strategy will always
// be greater than or equal to min.
func (base Strategy) Floor(min time.Duration) Strategy {
	if base == nil {
		panic("Overwrite called on nil Strategy")
	}
	return func(retry int) time.Duration {
		val := base(retry)
		if val < min {
			return min
		}
		return val
	}
}

// The Ceil method imposes a maximum value on the durations returned
// by a Strategy. Values returned by the resulting Strategy will always
// be less than or equal to max
func (base Strategy) Ceil(max time.Duration) Strategy {
	if base == nil {
		panic("Ceil called on nil Strategy")
	}
	return func(retry int) time.Duration {
		val := base(retry)
		if val > max {
			return max
		}
		return val
	}
}

// Overwrite replaces the first len(dur) mappings of a Strategy, selecting
// durations from the given parameter list instead. Passing len(dur) to
// the returned strategy is equivalent to passing len(dur) to the original
// strategy.
func (base Strategy) Overwrite(dur ...time.Duration) Strategy {
	if base == nil {
		panic("Overwrite called on nil Strategy")
	}
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
