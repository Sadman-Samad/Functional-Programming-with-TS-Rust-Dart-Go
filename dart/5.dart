const List<int> values = [3, 5, 6, 6, 7];

void main() {
  final mapped = values.map((v) => v * 3).toList();
  print(mapped);

  final filtered = values.where((v) => v == 6).toList();
  print(filtered);

  final found = values.where((v) => v == 8).firstOrNull;
  print(found ?? "None");

  final reduced = values
      .skip(1)
      .fold<int>(values.first, (acc, v) => acc * v + 2);
  print(reduced);
}
