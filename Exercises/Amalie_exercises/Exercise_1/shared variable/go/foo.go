// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"time"
)

var i = 0

func incrementing(incrementChannel chan int) {
	//TODO: increment i 1000000 times
	for j := 0; j < 1000000; j++ {
		i = i + 1
		incrementChannel <- i
	}
	close(incrementChannel)

}

func decrementing(decrementChannel chan int) {
	//TODO: decrement i 1000000 times
	for k := 0; k < 1000000; k++ {
		i = i - 1
		decrementChannel <- i
	}
	close(decrementChannel)
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(1)
	// setting it to 1, means only one thread will run Go code at any time.

	// TODO: Spawn both functions as goroutines

	//Making channels
	incrementChannel := make(chan int)
	decrementChannel := make(chan int)

	go incrementing(incrementChannel)
	go decrementing(decrementChannel)

	time.Sleep(500 * time.Millisecond)
	Println("The magic number is:", i)
}
