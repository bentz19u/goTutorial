package main

import (
	"errors"
	"testing"
)

func TestSafeDivide(t *testing.T) {
	tests := map[string]struct {
		a, b, want float64
		err        error
	}{"Success": {10, 5, 2.0, nil}, "Failure": {10, 0, 0.0, errDivideByZero}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := safeDivide(test.a, test.b)

			if !errors.Is(err, test.err) {
				t.Errorf("Failed error: want %v, got %v", test.err, err)
			}

			if test.want != result {
				t.Errorf("Failed value: want %v, got %v", test.want, result)
			}
		})
	}
}

func TestFindUserAge(t *testing.T) {
	tests := map[string]struct {
		users  map[string]int
		search string
		want   int
		err    error
	}{"Success": {map[string]int{"Daniel": 20, "Alice": 25}, "Daniel", 20, nil},
		"Failure": {map[string]int{"Daniel": 20, "Alice": 25}, "Bob", 0, ErrUserNotFound}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := findUserAge(test.users, test.search)

			if !errors.Is(err, test.err) {
				t.Errorf("Failed error: want %v, got %v", test.err, err)
			}

			if test.want != result {
				t.Errorf("Failed value: want %d, got %d", test.want, result)
			}
		})
	}
}

func TestGreetUser(t *testing.T) {
	tests := map[string]struct {
		users  map[string]int
		search string
		want   string
		err    error
	}{"Success": {map[string]int{"Daniel": 20, "Alice": 25}, "Daniel", "Greeting Daniel! Wow you are so young, you are 20", nil},
		"Failure": {map[string]int{"Daniel": 20, "Alice": 25}, "Bob", "", ErrUserNotFound}}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := greetUser(test.users, test.search)

			if !errors.Is(err, test.err) {
				t.Errorf("Failed error: want %v, got %v", test.err, err)
			}

			if test.want != result {
				t.Errorf("Failed value: want %q, got %q", test.want, result)
			}
		})
	}
}
