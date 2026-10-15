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
