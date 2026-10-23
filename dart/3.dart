typedef Op = int Function(int a, int b);

void main() {
  final Op v = (a, b) => a + b;

  final can = (int a, int b, Op c) => c(a, b);

  print(can(2, 3, v));
  print(can(4, 5, (a, b) => a * b));
}
