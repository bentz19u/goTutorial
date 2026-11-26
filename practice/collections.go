package main

import "fmt"

func sliceOfFood() {
	foods := []string{"pizza", "burger", "salad"}
	fmt.Println("foods list:", foods)
	fmt.Println("foods count:", len(foods))

	foods = append(foods, "poke", "hot dog")
	fmt.Println("foods list:", foods)
	fmt.Println("foods count:", len(foods))

	for idx, food := range foods {
		fmt.Println(idx, food)
	}

	for _, food := range foods {
		fmt.Println(food)
	}
}

func agesMap() {
	ages := map[string]int{"Daniel": 40, "Alice": 30, "Bob": 27}

	fmt.Println("ages Alice:", ages["Alice"])
	fmt.Println("ages Sandra:", ages["Sandra"])
	sandraAge, ok := ages["Sandra"]
	if !ok {
		fmt.Println("ages Sandra:", "not found")
	} else {
		fmt.Println("ages Sandra:", sandraAge)
	}

	for name, age := range ages {
		fmt.Println(name, age)
	}

	// order can be different
	for name, age := range ages {
		fmt.Println(name, age)
	}

	// order can be different
	for name, age := range ages {
		fmt.Println(name, age)
	}

	// order can be different
	for name, age := range ages {
		fmt.Println(name, age)
	}
}
