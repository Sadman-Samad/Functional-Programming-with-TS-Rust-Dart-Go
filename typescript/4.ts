const repeat = (
  value: number,
  operation: (n: number) => number
): number => operation(value);

const double = (n: number): number => n * 2;

console.log(repeat(10, double))
