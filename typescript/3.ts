


type Op = (a: number, b: number) => number;

const v: Op = (a, b) => a + b;

const c = (a: number, b: number, op: Op): number => op(a, b);

console.log(c(2, 3, v));            // 5
console.log(c(4, 5, (a, b) => a * b)); // 20

