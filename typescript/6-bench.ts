declare const process: { argv: string[] };

const values: number[] = [5, 6, 6, 7, 7];

function run(): number {
  return values
    .filter((v) => v % 2 === 0)
    .map((v) => v * 2)
    .reduce((a, b) => a * b);
}

const iterations = Number(process.argv[2] ?? 10_000_000);
let sink = 0;

for (let i = 0; i < iterations / 10; i++) {
  values[0] = 5 + (i % 2);
  sink ^= run();
}

const start = performance.now();
for (let i = 0; i < iterations; i++) {
  values[0] = 5 + (i % 2);
  sink ^= run();
}
const elapsed = performance.now() - start;
const nsPerOp = (elapsed * 1e6) / iterations;

console.log(
  `typescript: iterations=${iterations} result=${sink} ns/op=${nsPerOp.toFixed(2)}`,
);
