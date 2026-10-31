package benchmark

func Map[A, B any](items []A, transform func(A) B) []B {
	result := make([]B, 0, len(items))
	for _, item := range items {
		result = append(result, transform(item))
	}
	return result
}

func Filter[A any](items []A, keep func(A) bool) []A {
	result := make([]A, 0, len(items))
	for _, item := range items {
		if keep(item) {
			result = append(result, item)
		}
	}
	return result
}

func Reduce[A, B any](items []A, initial B, combine func(B, A) B) B {
	accumulator := initial
	for _, item := range items {
		accumulator = combine(accumulator, item)
	}
	return accumulator
}

func Pipeline[A, B, C any](
	items []A,
	keep func(A) bool,
	transform func(A) B,
	initial C,
	combine func(C, B) C,
) C {
	accumulator := initial
	for _, item := range items {
		if keep(item) {
			accumulator = combine(accumulator, transform(item))
		}
	}
	return accumulator
}
