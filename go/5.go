package main

import "fmt"

var values = [5]uint32{3, 5, 6, 6, 7}

func mapValues[A, B any](items []A, transform func(A) B) []B {
	result := make([]B, 0, len(items))
	for _, item := range items {
		result = append(result, transform(item))
	}
	return result
}

func filterValues[A any](items []A, keep func(A) bool) []A {
	result := make([]A, 0, len(items))
	for _, item := range items {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}

func findValue[A any](items []A, match func(A) bool) (A, bool) {
	for _, item := range items {
		if match(item) {
			return item, true
		}
	}
	var zero A
	return zero, false
}

func reduceValues[A, B any](items []A, initial B, combine func(B, A) B) B {
	accumulator := initial
	for _, item := range items {
		accumulator = combine(accumulator, item)
	}
	return accumulator
}

func main() {
	mapped := mapValues(values[:], func(v uint32) uint32 { return v * 3 })
	fmt.Println(mapped)

	filtered := filterValues(values[:], func(v uint32) bool { return v == 6 })
	fmt.Println(filtered)

	if found, ok := findValue(values[:], func(v uint32) bool { return v == 8 }); ok {
		fmt.Println(found)
	} else {
		fmt.Println("None")
	}

	reduced := reduceValues(values[1:], values[0], func(acc, v uint32) uint32 { return acc*v + 2 })
	fmt.Println(reduced)
}
