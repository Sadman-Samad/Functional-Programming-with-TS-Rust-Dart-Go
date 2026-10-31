package benchmark

import "testing"

var example6Values = []uint32{5, 6, 6, 7, 7}

func runExample6(items []uint32) uint32 {
	filtered := Filter(items, isEven)
	doubled := Map(filtered, double)
	return Reduce(doubled, uint32(1), multiply)
}

func BenchmarkExample6Go(b *testing.B) {
	b.ReportAllocs()
	var result uint32
	for i := 0; i < b.N; i++ {
		result = runExample6(example6Values)
	}
	sink = result
}
