package main

import (
	"errors"
	"fmt"
)

// errDivideByZero lower case first char = local
var errDivideByZero = errors.New("divide by zero")

// ErrUserNotFound upper case first char = exported
var ErrUserNotFound = errors.New("user not found")

func safeDivide(a, b float64) (float64, error) {
	if b == 0 {
		return 0.0, errDivideByZero
	}

	return a / b, nil
}

func findUserAge(users map[string]int, name string) (int, error) {
	age, ok := users[name]

	if !ok {
		return 0, ErrUserNotFound
	}

	return age, nil
}

// practicing %w
func greetUser(users map[string]int, name string) (string, error) {
	age, err := findUserAge(users, name)

	if err != nil {
		return "", fmt.Errorf("greeting user %s: %w", name, err)
	}

	return fmt.Sprintf("Greeting %s! Wow you are so young, you are %d", name, age), nil
}
