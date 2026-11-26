package main

import (
	"errors"
	"fmt"
	"strconv"
)

type ValidationError struct {
	Field  string
	Value  string
	Reason string
}

// errDivideByZero lower case first char = local
var errDivideByZero = errors.New("divide by zero")

// ErrUserNotFound upper case first char = exported
var ErrUserNotFound = errors.New("user not found")

func (err *ValidationError) Error() string {
	return fmt.Sprintf("field %q, value %q, reason %q", err.Field, err.Value, err.Reason)
}

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

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{
			Field:  "age",
			Value:  strconv.Itoa(age),
			Reason: "must not be negative",
		}
	}
	return nil
}

func registerUser(name string, age int) error {
	e := validateAge(age)
	if e != nil {
		return fmt.Errorf("registering user %s (age %d): %w", name, age, e)
	}

	// simulating creating user, we don't return anything, that's not the purpose of the exercise
	return nil
}
