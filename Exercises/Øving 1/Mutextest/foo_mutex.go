// Use `go run foo.go` to run your program

package main

import (
	. "fmt"
	"runtime"
	"sync"
)

//var i = 0

var (
	i    = 0
	lock sync.Mutex
)

func incrementing(wg *sync.WaitGroup) {
	//TODO: increment i 1000000 times
	//ch := make(chan int)
	for n := 0; n < 1000000; n++ {
		lock.Lock()
		i = i + 1
		//cha <- i
		//<-cha
		//Println("inc", i)
		lock.Unlock()

	}
	//close(cha)
	defer wg.Done()

	//cha <- ch
}

func decrementing(wg *sync.WaitGroup) {
	//TODO: decrement i 1000000 times
	//ch := make(chan int)
	for m := 0; m < 1000000; m++ {
		lock.Lock()
		i = i - 1
		//cha <- i
		//<-cha
		//Println("dec", i)
		lock.Unlock()

	}
	//close(cha)
	defer wg.Done()
	//cha <- ch

}

func main() {
	// What does GOMAXPROCS do? What happens if you set it to 1?
	runtime.GOMAXPROCS(2)
	var wg sync.WaitGroup

	//cha1 := make(chan int)
	//cha2 := make(chan int, 2000000)
	// TODO: Spawn both functions as goroutines
	wg.Add(2)
	go incrementing(&wg)
	go decrementing(&wg)

	wg.Wait()
	// We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
	// We will do it properly with channels soon. For now: Sleep.
	//time.Sleep(90 * time.Second)
	Println("The magic number is:", i)
}
