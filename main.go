package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"
)

func main() {
	// it's to separate the different exercises (mostly to see the syntax and Go specific rules)
	//lesson1()
	//lesson2()
	//lesson3()
	checkingInterruption()
}

func lesson1() {
	printSyntaxes()
	addAndPrint()
	basicForLoop()
	basicForLoopSwitch()
	fizzBuzzPrint()
	whileStyleLoop()
}

func lesson2() {
	sliceOfFood()
	agesMap()
}

func lesson3() {
	result := fizzBuzz(3)
	fmt.Println(result)
}

func printSyntaxes() {
	myInt := 40
	var myString string

	fmt.Println("my int:", myInt)
	fmt.Println("my string:", myString)
}

func addAndPrint() {
	result := add(5, 10)
	fmt.Println("result add func:", result)
}

func add(num1, num2 int) int {
	return num1 + num2
}

func basicForLoop() {
	for i := 0; i <= 30; i++ {
		if i < 18 {
			fmt.Println("minor")
		} else {
			fmt.Println("adult")
		}
	}
}

func basicForLoopSwitch() {
	for i := 1; i <= 20; i++ {
		switch {
		case i%3 == 0 && i%5 == 0:
			fmt.Println("FizzBuzz")
		case i%3 == 0:
			fmt.Println("Fizz")
		case i%5 == 0:
			fmt.Println("Buzz")
		default:
			fmt.Println(i)
		}
	}
}

func fizzBuzzPrint() {
	for i := 1; i <= 20; i++ {
		if i%3 == 0 && i%5 == 0 {
			fmt.Println("FizzBuzz")
		} else if i%3 == 0 {
			fmt.Println("Fizz")
		} else if i%5 == 0 {
			fmt.Println("Buzz")
		} else {
			fmt.Println(i)
		}
	}
}

func fizzBuzz(i int) string {
	if i%3 == 0 && i%5 == 0 {
		return "FizzBuzz"
	} else if i%3 == 0 {
		return "Fizz"
	} else if i%5 == 0 {
		return "Buzz"
	}
	return strconv.Itoa(i)
}

func whileStyleLoop() {
	n := 1
	loopCount := 0
	for n <= 1000 {
		n *= 2
		loopCount++
	}
	fmt.Println("whileStyleLoop loop count:", loopCount)
}

// will ctrl c send a signal to interrupt the process
// instead of termating it, it becomes a notification that will trigger a cancel() of the context
// the total calculated so far by the goroutines will be returned by runPollers and printed
func checkingInterruption() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	defer stop()

	total := runPollers(ctx, 5, 100*time.Millisecond)

	fmt.Println(total)
}
