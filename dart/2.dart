void main() {
  final value = <int>[30, 40, 40];

  final value2 = [...value, 6];
  value2.add(3);

  print(value2);
}
