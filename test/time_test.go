package main

import (
	"fmt"
	"testing"
	"time"
)

func Test_time(t *testing.T) {
	start := time.Now()

	time.Sleep(1 * time.Second)

	duration := time.Since(start)

	fmt.Printf("%+v ms\n",duration.Milliseconds())
}
