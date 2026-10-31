package benchmark

import "testing"

func isEven(v uint32) bool          { return v%2 == 0 }
func double(v uint32) uint32        { return v * 2 }
func multiply(acc, v uint32) uint32 { return acc * v }

func runFP(items []uint32) uint32 {
	filtered := Filter(items, isEven)
	doubled := Map(filtered, double)
	return Reduce(doubled, uint32(1), multiply)
}

func runPipeline(items []uint32) uint32 {
	return Pipeline(items, isEven, double, uint32(1), multiply)
}

func runLoop(items []uint32) uint32 {
	accumulator := uint32(1)
	for _, v := range items {
		if isEven(v) {
			accumulator *= double(v)
		}
	}
	return accumulator
}

func makeValues(size int) []uint32 {
	items := make([]uint32, size)
	for i := range items {
		items[i] = uint32(i%7 + 1)
	}
	return items
}

var sink uint32

func TestImplementationsAgree(t *testing.T) {
	for _, items := range [][]uint32{
		{5, 6, 6, 7, 7},
		makeValues(1),
		makeValues(1_000),
	} {
		want := runLoop(items)
		if got := runFP(items); got != want {
			t.Fatalf("FP = %d, want %d", got, want)
		}
		if got := runPipeline(items); got != want {
			t.Fatalf("Pipeline = %d, want %d", got, want)
		}
	}
}

func benchmark(b *testing.B, run func([]uint32) uint32, items []uint32) {
	b.ReportAllocs()
	b.ResetTimer()
	var result uint32
	for i := 0; i < b.N; i++ {
		result = run(items)
	}
	sink = result
}

func BenchmarkSmallFP(b *testing.B)       { benchmark(b, runFP, []uint32{5, 6, 6, 7, 7}) }
func BenchmarkSmallLoop(b *testing.B)     { benchmark(b, runLoop, []uint32{5, 6, 6, 7, 7}) }
func BenchmarkSmallPipeline(b *testing.B) { benchmark(b, runPipeline, []uint32{5, 6, 6, 7, 7}) }

func BenchmarkLargeFP(b *testing.B)       { benchmark(b, runFP, makeValues(10_000)) }
func BenchmarkLargeLoop(b *testing.B)     { benchmark(b, runLoop, makeValues(10_000)) }
func BenchmarkLargePipeline(b *testing.B) { benchmark(b, runPipeline, makeValues(10_000)) }

var _ = sink
