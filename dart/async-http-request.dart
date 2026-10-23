import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'dart:isolate';

class Todo {
  final int userId;
  final int id;
  final String title;
  final bool completed;

  Todo({
    required this.userId,
    required this.id,
    required this.title,
    required this.completed,
  });

  factory Todo.fromJson(Map<String, dynamic> json) {
    return Todo(
      userId: json['userId'] as int,
      id: json['id'] as int,
      title: json['title'] as String,
      completed: json['completed'] as bool,
    );
  }
}

Future<Todo> fetchTodo(HttpClient client, int id) async {
  final request = await client.getUrl(
    Uri.parse('https://jsonplaceholder.typicode.com/todos/$id'),
  );
  final response = await request.close();

  if (response.statusCode != HttpStatus.ok) {
    throw HttpException('Request failed with status ${response.statusCode}');
  }

  final body = await response.transform(utf8.decoder).join();
  return Todo.fromJson(jsonDecode(body) as Map<String, dynamic>);
}

Future<void> backgroundLog(int id) async {
  final client = HttpClient();

  try {
    final todo = await fetchTodo(client, id);
    print('[spawn] background todo ${todo.id}: ${todo.title}');
  } catch (error) {
    print('[spawn] background request failed: $error');
  } finally {
    client.close(force: true);
  }
}

int heavyCpuWork(int limit) {
  var sum = 0;
  for (var i = 0; i <= limit; i++) {
    sum = (sum + i * i) % 1000000007;
  }
  return sum;
}

void printTodo(String label, Todo todo) {
  print('[$label] ${todo.id}: ${todo.title} (completed: ${todo.completed})');
}

Future<void> main() async {
  final client = HttpClient();

  try {
    unawaited(backgroundLog(9));

    print('[await] fetching sequentially');
    final first = await fetchTodo(client, 1);
    final second = await fetchTodo(client, 2);
    printTodo('await', first);
    printTodo('await', second);

    print('[join] fetching concurrently');
    final joined = await Future.wait([
      fetchTodo(client, 3),
      fetchTodo(client, 4),
    ]);
    for (final todo in joined) {
      printTodo('join', todo);
    }

    print('[select] racing request against a 2s timeout');
    final raceClient = HttpClient();

    try {
      final winner = await Future.any([
        fetchTodo(raceClient, 5),
        Future<Todo>.delayed(
          const Duration(seconds: 2),
          () => throw TimeoutException('timeout after 2s'),
        ),
      ]);
      printTodo('select', winner);
    } on TimeoutException catch (error) {
      raceClient.close(force: true);
      print('[select] ${error.message}');
    } finally {
      raceClient.close(force: true);
    }

    print('[blocking] running CPU-bound work in an isolate');
    final checksum = await Isolate.run(() => heavyCpuWork(10_000_000));
    print('[blocking] checksum = $checksum');

    await Future<void>.delayed(const Duration(milliseconds: 500));
  } finally {
    client.close(force: true);
  }
}
