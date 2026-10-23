package main

import "fmt"

func addValue(v uint32, v2 uint32) uint32 {
	value := uint32(40)
	value += v + v2
	return value
}

func main() {
	value := addValue(20, 40)
	fmt.Print(value)

	text := "hhhh"
	if 1000 == 1_000 {
		fmt.Println(text)
	} else {
		fmt.Println("None")
	}
}
