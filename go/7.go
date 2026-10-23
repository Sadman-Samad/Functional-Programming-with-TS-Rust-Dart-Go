package main

import "fmt"

func compose[A, B, C any](f func(B) C, g func(A) B) func(A) C {
	return func(value A) C {
		return f(g(value))
	}
}

func main() {
	double := func(n uint32) uint32 { return n * 2 }
	addTwo := func(n uint32) uint32 { return n + 2 }

	composed := compose(double, addTwo)

	fmt.Println(composed(50))
}
