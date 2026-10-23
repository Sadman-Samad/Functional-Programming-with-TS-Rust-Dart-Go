package main

import "fmt"

var values = [5]uint32{5, 6, 6, 7, 7}

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

func reduceValues[A, B any](items []A, initial B, combine func(B, A) B) B {
	accumulator := initial
	for _, item := range items {
		accumulator = combine(accumulator, item)
	}
	return accumulator
}

func main() {
	evens := filterValues(values[:], func(v uint32) bool { return v%2 == 0 })
	doubled := mapValues(evens, func(v uint32) uint32 { return v * 2 })
	product := reduceValues(doubled, uint32(1), func(acc, v uint32) uint32 { return acc * v })

	fmt.Println([]uint32{product})
}
