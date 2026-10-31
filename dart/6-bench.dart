final List<int> values = [5, 6, 6, 7, 7];

int run() => values
    .where((v) => v % 2 == 0)
    .map((v) => v * 2)
    .reduce((a, b) => a * b);

void main(List<String> args) {
  final iterations = args.isNotEmpty ? int.parse(args.first) : 10000000;
  var sink = 0;

  for (var i = 0; i < iterations ~/ 10; i++) {
    values[0] = 5 + (i % 2);
    sink ^= run();
  }

  final stopwatch = Stopwatch()..start();
  for (var i = 0; i < iterations; i++) {
    values[0] = 5 + (i % 2);
    sink ^= run();
  }
  stopwatch.stop();

  final nsPerOp = stopwatch.elapsedMicroseconds * 1000 / iterations;
  print(
    'dart: iterations=$iterations result=$sink ns/op=${nsPerOp.toStringAsFixed(2)}',
  );
}
