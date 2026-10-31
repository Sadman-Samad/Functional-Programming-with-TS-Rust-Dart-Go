use std::hint::black_box;
use std::time::Instant;

static VALUES: [u32; 5] = [5, 6, 6, 7, 7];

fn run(values: &[u32]) -> Vec<u32> {
    values
        .iter()
        .filter(|v: &&u32| *v % 2 == 0)
        .map(|v| v * 2)
        .reduce(|acc: u32, v: u32| acc * v)
        .into_iter()
        .collect()
}

fn main() {
    let iterations: u64 = std::env::args()
        .nth(1)
        .and_then(|arg| arg.parse().ok())
        .unwrap_or(10_000_000);

    let mut sink = 0u32;
    for _ in 0..iterations / 10 {
        sink ^= black_box(run(black_box(&VALUES)))[0];
    }

    let start = Instant::now();
    for _ in 0..iterations {
        sink ^= black_box(run(black_box(&VALUES)))[0];
    }
    let elapsed = start.elapsed();

    println!(
        "rust: iterations={iterations} result={sink} ns/op={:.2}",
        elapsed.as_nanos() as f64 / iterations as f64
    );
}
