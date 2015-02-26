package retry_test

import (
	"fmt"
	"time"

	"aqwari.net/retry"
)

func ExampleExponential() {
	backoff := retry.Exponential(time.Hour / 2).Splay(100)

	for tries := 0; tries < 10; tries++ {
		fmt.Println("Connecting to service")
		b := backoff(tries)
		fmt.Printf("Connection failed; will try again in %s", b)
		time.Sleep(b)
	}
}

func ExampleSeconds() {
	backoff := retry.Seconds(2, 4, 6, 22, 39, 18)

	for tries := 0; tries < 10; tries++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(tries))
	}
	// Output: Connection failed; will try again in 2s
	// Connection failed; will try again in 4s
	// Connection failed; will try again in 6s
	// Connection failed; will try again in 22s
	// Connection failed; will try again in 39s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
	// Connection failed; will try again in 18s
}

func ExampleMilliseconds() {
	backoff := retry.Milliseconds(2, 4, 6, 22, 39, 18)

	for tries := 0; tries < 10; tries++ {
		fmt.Printf("Connection failed; will try again in %s\n", backoff(tries))
	}
	// Output: Connection failed; will try again in 2ms
	// Connection failed; will try again in 4ms
	// Connection failed; will try again in 6ms
	// Connection failed; will try again in 22ms
	// Connection failed; will try again in 39ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
	// Connection failed; will try again in 18ms
}

func ExampleStrategy_Scale() {
	// Sleep for 2ⁿ milliseconds, not seconds
	backoff := retry.Exponential(-1).Scale(1e-3)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}

	// Output: 1ms
	// 2ms
	// 4ms
	// 8ms
	// 16ms
	// 32ms
	// 64ms
	// 128ms
	// 256ms
	// 512ms
}

func ExampleStrategy_Splay() {
	// Splay ± ½ second
	backoff := retry.Exponential(-1).Splay(.5)

	for tries := 0; tries < 10; tries++ {
		fmt.Println(backoff(tries))
	}
}
