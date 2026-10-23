package main

import "fmt"

func main() {
	value := []uint32{30, 40, 40}

	value2 := make([]uint32, len(value), len(value)+2)
	copy(value2, value)
	value2 = append(value2, 6, 3)

	fmt.Println(value2)
}
