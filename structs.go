package main

import "fmt"

type person struct {
	name string
	age  int
}

func (p *person) greet() string {
	return fmt.Sprintf("Hi, I'm %s", p.name)
}

func (p *person) birthday() {
	p.age++
}
