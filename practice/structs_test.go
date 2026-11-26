package main

import "testing"

// just testing struct, I know the test by itself is useless
func TestPerson(t *testing.T) {
	p := person{name: "Daniel", age: 28}
	expectedName := "Daniel"
	expectedAge := 28

	if p.name != expectedName {
		t.Errorf("Failed: expected %q, got %q", expectedName, p.name)
	}

	if p.age != expectedAge {
		t.Errorf("Failed: expected %d, got %d", expectedAge, p.age)
	}
}

func TestEmptyPerson(t *testing.T) {
	var p person
	expectedName := ""
	expectedAge := 0

	// an empty str will return ""
	if p.name != expectedName {
		t.Errorf("Failed: expected %q, got %q", expectedName, p.name)
	}

	// an empty int will return 0
	if p.age != expectedAge {
		t.Errorf("Failed: expected %d, got %d", expectedAge, p.age)
	}
}

func TestGreet(t *testing.T) {
	p := person{name: "Daniel", age: 28}
	expected := "Hi, I'm Daniel"
	got := p.greet()

	if got != expected {
		t.Errorf("Failed: expected %q, got %q", expected, got)
	}
}

func TestBirthday(t *testing.T) {
	p := person{name: "Daniel", age: 40}
	p.birthday()
	expected := 41
	got := p.age

	// age change because birthday use a `person` pointer
	if got != expected {
		t.Errorf("Failed: expected %d, got %d", expected, got)
	}
}
