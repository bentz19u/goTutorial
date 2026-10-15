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