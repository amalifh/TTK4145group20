// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var i = 0

func incrementing() {
	//TODO: increment i 1000000 times
	for n := 0; n < 1000000; n++ {
		i= i + 1
		Println("inc",i)
	}
}

func decrementing() {
	//TODO: decrement i 1000000 times
	for m := 0; m < 1000000; m++ {
		i= i - 1 
		Println("dec",i)
	}
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(1)

	// TODO: Spawn both functions as goroutines
	go incrementing()
	
	go decrementing()

	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	 time.Sleep(90 * time.Second)
	Println("The magic number is:", i)
}
