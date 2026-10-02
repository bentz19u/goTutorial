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

func TestValidateAge(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		age := 20
		e := validateAge(age)

		if e != nil {
			t.Errorf("Failed: want nil, got %s", e)
		}

		var vErr *ValidationError
		if errors.As(e, &vErr) {
			t.Errorf("Failed: age %d is invalid", age)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		age := -8
		e := validateAge(age)

		if e == nil {
			t.Fatalf("Failed: want not nil, got %v", e)
		}

		// old version
		var vErr *ValidationError
		if !errors.As(e, &vErr) {
			t.Fatalf("Failed: age %d should be invalid, got %v", age, e)
		}

		if vErr.Reason != "must not be negative" {
			t.Errorf("Failed reason: want %q, got %q", "must not be negative", vErr.Reason)
		}

		// new version of handling that kind of error
		if _, ok := errors.AsType[*ValidationError](e); !ok {
			t.Errorf("Failed: age %d should be invalid", age)
		}
	})
}

func TestRegisterUser(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		name := "Daniel"
		age := 20
		e := registerUser(name, age)

		if e != nil {
			t.Fatalf("Failed: want nil, got %v", e)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		name := "Daniel"
		age := -8
		e := registerUser(name, age)

		if e == nil {
			t.Fatalf("Failed: want not nil, got %v", e)
		}

		vErr, ok := errors.AsType[*ValidationError](e)
		if !ok {
			t.Fatalf("Failed: want ValidationError, got %v", e)
		}

		if vErr.Field != "age" {
			t.Errorf("Failed: want age, got %s", vErr.Field)
		}

		if vErr.Value != "-8" {
			t.Errorf("Failed: want -8, got %s", vErr.Value)
		}

		reason := "must not be negative"
		if vErr.Reason != reason {
			t.Errorf("Failed: want %q, got %q", reason, vErr.Reason)
		}
	})
}
