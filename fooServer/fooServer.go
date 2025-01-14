// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"sync"
)

var i = 0

func Server(cha1 chan bool, cha2 chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-cha1:
			if !ok {
				cha1 = nil
			} else {
				i++
			}
		case _, ok := <-cha2:
			if !ok {
				cha2 = nil
			} else {
				i--
			}
		}
		if cha1 == nil && cha2 == nil {
			return
		}
	}
}

func incrementing(cha chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	//TODO: increment i 1000000 times
	for n := 0; n < 1000000; n++ {
		cha <- true
	}
	close(cha)
}

func decrementing(cha chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	//TODO: decrement i 1000000 times
	for m := 0; m < 1000000; m++ {
		cha <- false
	}
	close(cha)
}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(2)
	var wg sync.WaitGroup

	// TODO: Spawn both functions as goroutines
	for {
		chInc := make(chan bool)
		chDec := make(chan bool)
		wg.Add(3)
		go incrementing(chInc, &wg)
		go decrementing(chDec, &wg)
		go Server(chInc, chDec, &wg)
		wg.Wait()
		//time.Sleep(5 * time.Second)
		Println("The magic number is:", i)
		i = 0
	}
}
