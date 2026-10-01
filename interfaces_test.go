package main

import (
	"fmt"
	"testing"
)

func TestIntroducePerson(t *testing.T) {
	p := person{name: "Daniel", age: 42}
	// we cannot use p as a value (compared to a pointer) because greet is using a pointer receiver
	// got := introduce(p)
	got := introduce(&p)
	want := "Introduction: Hi, I'm Daniel"

	if got != want {
		t.Errorf("Failed: want %q, got %q", want, got)
	}
}

func TestIntroduceRobot(t *testing.T) {
	r := robot{name: "Cylon"}
	got := introduce(&r)
	want := "Introduction: Hello Human, I'm Cylon"

	if got != want {
		t.Errorf("Failed: want %q, got %q", want, got)
	}
}

func TestPersonStringer(t *testing.T) {
	p := person{name: "Daniel", age: 42}

	// &p is a *person, whose method set includes String(), fmt uses the Stringer
	got := fmt.Sprintf("%v", &p)
	want := "My name is Daniel and my age is 42"

	if got != want {
		t.Errorf("Failed: want %q, got %q", want, got)
	}

	// p is a person VALUE — String() has a pointer receiver
	// so it's not in the value's method set
	// fmt silently falls back to the default struct rendering
	got = fmt.Sprintf("%v", p)
	want = "{Daniel 42}"

	if got != want {
		t.Errorf("Failed: want %q, got %q", want, got)
	}
}
