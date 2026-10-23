int addValue(int v, int v2) {
  int value = 40;
  value += v + v2;
  return value;
}

void main() {
  int value = addValue(20, 40);
  print(value);

  String text = "hhhh";
  if (1000 == 1_000) {
    print(text);
  } else {
    print("None");
  }
}
