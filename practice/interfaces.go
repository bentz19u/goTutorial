package main

import "fmt"

type greeter interface {
	greet() string
}

func introduce(g greeter) string {
	return fmt.Sprintf("Introduction: %s", g.greet())
}
