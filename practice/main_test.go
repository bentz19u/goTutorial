package main

import "testing"

func TestFizzBuzz(t *testing.T) {
	tests := map[int]string{3: "Fizz", 5: "Buzz", 7: "7", 15: "FizzBuzz"}
	for i, expected := range tests {
		result := fizzBuzz(i)

		if result != expected {
			t.Errorf("Failed: input %d, expected %q, got %q", i, expected, result)
		}
	}
}
