package main

import (
	"sync"
)

// checking how defer works
func doWork(log *[]string) {
	*log = append(*log, "start")
	defer func() { *log = append(*log, "cleanup A") }()
	defer func() { *log = append(*log, "cleanup B") }()
	*log = append(*log, "end")
}

// first concurrency test
func squareAll(nums []int) []int {
	var wg sync.WaitGroup

	result := make([]int, len(nums))
	for i, num := range nums {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result[i] = num * num
		}()
	}

	wg.Wait()

	return result
}

// second concurrency test, kept for the glorious history of learning concurrency in Go Lang
func racyCounter() int {
	var wg sync.WaitGroup

	racyCount := 0
	// lunching 100 goroutines
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// each goroutines increase the counter by 100 (race conditions expected)
			for range 100 {
				racyCount++
			}
		}()
	}

	wg.Wait()

	return racyCount
}

// fixed second concurrency test
func counter() int {
	var wg sync.WaitGroup
	var lock sync.Mutex

	count := 0
	// lunching 100 goroutines
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// each goroutines increase the counter by 100 (race conditions unexpected)
			for range 100 {
				lock.Lock()
				count++
				lock.Unlock()
			}
		}()
	}

	wg.Wait()

	return count
}

func fetchCount(ch chan<- int) {
	count := 0
	for range 100 {
		count++
	}
	ch <- count
}

// third concurrency test with chan
func channelCounter() int {
	var wg sync.WaitGroup

	ch := make(chan int)

	// we generate 100 goroutine that will send their values to the channel
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchCount(ch)
		}()
	}

	// another goroutine that will close the channel once they are all done (doesnt stop the process)
	go func() {
		wg.Wait()
		close(ch)
	}()

	// we reach here while the goroutines are doing their job
	total := 0
	for count := range ch {
		total += count
	}

	return total
}

// in this one, we use a buffered channel, all Done() are fired on send (if the buffer have room), not when comsumers so the wg.Wait() worked inline
// when unbuffered, all Done() are fired when the messages are consumed
func bufferedChannelCounter() int {
	var wg sync.WaitGroup
	ch := make(chan int, 100)

	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fetchCount(ch)
		}()
	}

	wg.Wait()
	close(ch)

	total := 0
	for count := range ch {
		total += count
	}

	return total
}
