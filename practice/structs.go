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

func (p *person) String() string {
	return fmt.Sprintf("My name is %s and my age is %d", p.name, p.age)
}

type robot struct {
	name string
}

func (r *robot) greet() string {
	return fmt.Sprintf("Hello Human, I'm %s", r.name)
}
