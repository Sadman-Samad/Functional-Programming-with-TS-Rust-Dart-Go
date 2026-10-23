const double = (n: number): number => n * 2;
const addOne = (n: number): number => n + 1;

const compose =
  <A, B, C>(f: (value: B) => C, g: (value: A) => B) =>
  (value: A): C =>
    f(g(value));

const doubleThenAddOne = compose(addOne, double);

console.log(doubleThenAddOne(5));
