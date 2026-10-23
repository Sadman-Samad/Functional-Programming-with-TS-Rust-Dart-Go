const List<int> values = [5, 6, 6, 7, 7];

void main() {
  final filtered = values.where((v) => v % 2 == 0).map((v) => v * 2).toList();
  final reduced = filtered.reduce((a, b) => a * b);

  print([reduced]);
}
