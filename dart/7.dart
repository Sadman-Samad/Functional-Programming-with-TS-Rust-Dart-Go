C Function(A) compose<A, B, C>(C Function(B) f, B Function(A) g) {
  return (A value) => f(g(value));
}

void main() {
  final double = (int n) => n * 2;
  final addTwo = (int n) => n + 2;

  final composed = compose(double, addTwo);

  print(composed(50));
}
