# Dart Functional Programming with fpdart

## 29 Concepts: Traditional Dart vs Functional Programming

This guide compares traditional imperative/object-oriented Dart with
functional programming (FP) using the `fpdart` package.

The examples use common application scenarios such as transforming data,
handling missing values, managing API failures, and composing
asynchronous operations.

> **Important:** FP is a programming style, not a requirement to use a
> particular package. Dart is multi-paradigm, so you can combine OOP,
> imperative programming, and FP in the same application. Some examples
> below use standard Dart features (`map`, `where`, `fold`); others use
> `fpdart`.

## Installation

``` bash
flutter pub add fpdart
```

``` dart
import 'package:fpdart/fpdart.dart';
```

------------------------------------------------------------------------

## 1. Pure Functions

A pure function: - Returns the same output for the same input. - Does
not modify external state. - Does not perform side effects.

**Without FP:**

``` dart
int calculateDiscount(int price) {
  print('Calculating discount'); // Side effect
  return price - (price * 10 ~/ 100);
}
```

**With FP:**

``` dart
int calculateDiscount(int price) {
  return price - (price * 10 ~/ 100);
}

final result = calculateDiscount(1000);
print(result); // 900
```

The second function is pure: it only calculates and returns a value.
Pure functions are a general FP principle, not something exclusive to
`fpdart`.

## 2. Immutability

Immutability means not changing existing data. Instead, create a new
value with the desired changes.

**Without FP:**

``` dart
class User {
  String name;
  int age;

  User(this.name, this.age);
}

final user = User('John', 25);
user.age = 26; // Mutates the existing object
```

**With FP:**

``` dart
class User {
  final String name;
  final int age;

  const User({
    required this.name,
    required this.age,
  });

  User copyWith({String? name, int? age}) {
    return User(
      name: name ?? this.name,
      age: age ?? this.age,
    );
  }
}

final user = User(name: 'John', age: 25);
final updatedUser = user.copyWith(age: 26);
```

`user` remains unchanged; `updatedUser` is a new object.

## 3. Higher-Order Functions

A higher-order function accepts another function as an argument or
returns a function.

**Without FP:**

``` dart
int calculate(int a, int b) => a + b;

int execute(int a, int b) {
  return calculate(a, b);
}
```

**With FP:**

``` dart
int execute(
  int a,
  int b,
  int Function(int, int) operation,
) {
  return operation(a, b);
}

final sum = execute(10, 20, (a, b) => a + b);
final product = execute(10, 20, (a, b) => a * b);
```

The same function can execute different behaviours supplied by the
caller.

## 4. First-Class Functions

Dart functions are first-class values. You can store them in variables,
pass them to other functions, and return them.

**Without FP / OOP-style:**

``` dart
class Calculator {
  int add(int a, int b) => a + b;
}

final calculator = Calculator();
final result = calculator.add(10, 20);
```

**With FP:**

``` dart
final int Function(int, int) add = (a, b) => a + b;

final result = add(10, 20);
```

Both styles are valid. The functional version treats the behaviour
itself as a value.

## 5. Function Composition

Function composition combines small functions to create a larger
operation.

**Without FP:**

``` dart
String processName(String name) {
  final trimmed = name.trim();
  final uppercase = trimmed.toUpperCase();
  return 'Hello $uppercase';
}
```

**With FP-style composition:**

``` dart
String trimName(String name) => name.trim();
String uppercase(String name) => name.toUpperCase();
String greet(String name) => 'Hello $name';

final result = greet(
  uppercase(
    trimName(' John '),
  ),
);

print(result); // Hello JOHN
```

Each function has one responsibility and can be reused independently.

## 6. Map

`map` transforms every item in a collection into another value.

**Without FP (imperative loop):**

``` dart
final numbers = [1, 2, 3, 4];
final results = <int>[];

for (final number in numbers) {
  results.add(number * 2);
}
```

**With FP-style collection processing:**

``` dart
final numbers = [1, 2, 3, 4];

final results = numbers
    .map((number) => number * 2)
    .toList();

// [2, 4, 6, 8]
```

Dart's built-in `Iterable.map` is already a functional-style operation;
it does not require `fpdart`.

## 7. Where / Filter

`where` keeps only the elements that satisfy a condition.

**Without FP (imperative loop):**

``` dart
final numbers = [1, 2, 3, 4, 5, 6];
final results = <int>[];

for (final number in numbers) {
  if (number.isEven) {
    results.add(number);
  }
}
```

**With FP-style collection processing:**

``` dart
final numbers = [1, 2, 3, 4, 5, 6];

final results = numbers
    .where((number) => number.isEven)
    .toList();

// [2, 4, 6]
```

## 8. Fold

`fold` processes a collection and combines its elements into one result.

**Without FP:**

``` dart
final prices = [100, 200, 300];
var total = 0;

for (final price in prices) {
  total += price;
}
```

**With FP-style collection processing:**

``` dart
final prices = [100, 200, 300];

final total = prices.fold(
  0,
  (sum, price) => sum + price,
);

// 600
```

The initial value is `0`; each iteration combines the accumulated value
with the next price.

## 9. Option

`Option<T>` represents a value that may or may not exist.

It has two cases: - `Some(value)`: a value exists. - `None()`: no value
exists.

**Without FP:**

``` dart
User? findUser(int id) {
  if (id == 1) {
    return User(name: 'John', age: 25);
  }

  return null;
}

final user = findUser(10);

if (user != null) {
  print(user.name);
} else {
  print('User not found');
}
```

**With fpdart:**

``` dart
Option<User> findUser(int id) {
  if (id == 1) {
    return Some(User(name: 'John', age: 25));
  }

  return None();
}

final result = findUser(10);

result.match(
  () => print('User not found'),
  (user) => print(user.name),
);
```

`Option<User>` makes the possibility of a missing user explicit in the
return type.

## 10. Either

`Either<L, R>` represents one of two possible outcomes: - `Left(error)`:
conventionally used for failure. - `Right(value)`: conventionally used
for success.

**Without FP:**

``` dart
User fetchUser() {
  try {
    return api.getUser();
  } catch (error) {
    throw Exception('Failed to fetch user');
  }
}
```

The caller must catch the exception.

**With fpdart:**

``` dart
Either<String, User> fetchUser() {
  try {
    final user = api.getUser();
    return Right(user);
  } catch (error) {
    return Left('Failed to fetch user: $error');
  }
}
```

Usage:

``` dart
final result = fetchUser();

result.match(
  (error) => print('Error: $error'),
  (user) => print('User: ${user.name}'),
);
```

The caller handles both outcomes explicitly. In real applications,
prefer meaningful typed error models over plain strings.

## 11. Map vs FlatMap

These operations are central to working with values inside containers
such as `Option` and `Either`.

-   `map` transforms the contained value using a function that returns a
    plain value.
-   `flatMap` transforms the contained value using a function that
    returns another container of the same kind, avoiding nested
    containers.

**Without FP:**

``` dart
int? parseAge(String value) {
  return int.tryParse(value);
}

int? validateAge(int age) {
  if (age < 18) return null;
  return age;
}

final parsed = parseAge('25');

final result = parsed == null
    ? null
    : validateAge(parsed);
```

**With fpdart:**

``` dart
Option<int> parseAge(String value) {
  return Option.fromNullable(int.tryParse(value));
}

Option<int> validateAge(int age) {
  if (age < 18) return None();
  return Some(age);
}

final result = parseAge('25').flatMap(validateAge);
```

Using `map` for a plain transformation:

``` dart
final doubled = parseAge('25').map((age) => age * 2);
```

Both `map` and `flatMap` preserve the `Option` context.

## 12. IO

`IO<A>` represents a synchronous computation that may perform side
effects and eventually produce an `A`.

**Without FP:**

``` dart
String readUsername() {
  print('Reading username');
  return 'John';
}

final username = readUsername(); // Runs immediately
```

**With fpdart:**

``` dart
IO<String> readUsername() {
  return IO(() {
    print('Reading username');
    return 'John';
  });
}

final usernameIO = readUsername(); // Describes the computation
final username = usernameIO.run(); // Executes it
```

`IO` helps separate describing a side-effecting operation from running
it. It does not make the side effect disappear.

## 13. Task

`Task<A>` represents an asynchronous computation that produces an `A`,
commonly backed by a `Future`.

**Without FP:**

``` dart
Future<User> fetchUser() async {
  return await api.getUser();
}
```

**With fpdart:**

``` dart
Task<User> fetchUser() {
  return Task(() async {
    return await api.getUser();
  });
}
```

Usage:

``` dart
final task = fetchUser();
final user = await task.run();
```

You can compose operations:

``` dart
final task = Task(() async => 10)
    .map((number) => number * 2);

final result = await task.run(); // 20
```

A `Task` does not model typed failures. Use `TaskEither` when failure is
part of the expected result.

## 14. TaskEither

`TaskEither<L, R>` combines asynchronous computation with explicit
success/failure handling.

**Without FP:**

``` dart
Future<User> fetchUser() async {
  try {
    return await api.getUser();
  } catch (error) {
    throw Exception('Unable to fetch user');
  }
}
```

**With fpdart:**

``` dart
TaskEither<String, User> fetchUser() {
  return TaskEither.tryCatch(
    () async => await api.getUser(),
    (error, stackTrace) => 'Unable to fetch user: $error',
  );
}
```

Usage:

``` dart
final result = await fetchUser().run();

result.match(
  (error) => print('Failed: $error'),
  (user) => print('Success: ${user.name}'),
);
```

This is especially useful for API calls, authentication, database
operations, and repository methods.

## 15. IOEither

`IOEither<L, R>` represents a synchronous computation that can fail with
an `L` or succeed with an `R`.

**Without FP:**

``` dart
int readConfiguration() {
  try {
    return int.parse('123');
  } catch (error) {
    throw Exception('Invalid configuration');
  }
}
```

**With fpdart:**

``` dart
IOEither<String, int> readConfiguration() {
  return IOEither.tryCatch(
    () => int.parse('123'),
    (error, stackTrace) => 'Invalid configuration',
  );
}
```

Usage:

``` dart
final result = readConfiguration().run();

result.match(
  (error) => print(error),
  (value) => print(value),
);
```

## 16. TaskOption

`TaskOption<A>` represents an asynchronous computation whose result may
be absent.

For example, a remote lookup may succeed but find no matching user.

**Without FP:**

``` dart
Future<User?> findUser(int id) async {
  return await database.findUser(id);
}
```

**With fpdart:**

``` dart
TaskOption<User> findUser(int id) {
  return TaskOption(() async {
    final user = await database.findUser(id);
    return Option.fromNullable(user);
  });
}
```

Usage:

``` dart
final result = await findUser(10).run();

result.match(
  () => print('User not found'),
  (user) => print(user.name),
);
```

Use `TaskOption` when absence is a normal outcome. If the operation can
also fail in a way you need to report, consider `TaskEither`.

## 17. Reader

`Reader<E, A>` represents a computation that needs an environment or
dependency `E` to produce a value `A`.

**Without FP:**

``` dart
class UserRepository {
  final ApiClient api;

  UserRepository(this.api);

  User getUser() => api.getUser();
}

final repository = UserRepository(apiClient);
final user = repository.getUser();
```

**With fpdart:**

``` dart
final getUser = Reader<ApiClient, User>(
  (api) => api.getUser(),
);

final user = getUser.run(apiClient);
```

`Reader` is useful for passing dependencies explicitly and making
computations easier to test.

## 18. ReaderTaskEither

`ReaderTaskEither<E, L, R>` combines: - `Reader`: requires a
dependency/environment. - `Task`: performs asynchronous work. -
`Either`: represents failure or success.

**Without FP:**

``` dart
Future<User> getUser(ApiClient api) async {
  try {
    return await api.getUser();
  } catch (error) {
    throw Exception('Failed: $error');
  }
}
```

**With fpdart:**

``` dart
ReaderTaskEither<ApiClient, String, User> getUser() {
  return ReaderTaskEither(
    (api) => TaskEither.tryCatch(
      () async => await api.getUser(),
      (error, stackTrace) => 'Failed: $error',
    ),
  );
}
```

Usage:

``` dart
final result = await getUser()
    .run(apiClient)
    .run();

result.match(
  (error) => print(error),
  (user) => print(user.name),
);
```

This can be useful for repository and service layers where dependencies,
asynchronous work, and expected failures need to be composed.

## 19. State

`State<S, A>` represents a computation that receives a state `S` and
produces a result `A` alongside a new state `S`.

**Without FP:**

``` dart
class Wallet {
  int balance = 100;

  void addMoney(int amount) {
    balance += amount;
  }
}

final wallet = Wallet();
wallet.addMoney(50);

print(wallet.balance); // 150
```

**With fpdart:**

``` dart
final addMoney = State<int, int>(
  (balance) {
    final newBalance = balance + 50;
    return Tuple2(newBalance, newBalance);
  },
);

final result = addMoney.run(100);

print(result.first);  // Result: 150
print(result.second); // New state: 150
```

The original input state is not mutated. The computation returns the
next state and a result.

## 20. StateAsync

`StateAsync<S, A>` represents an asynchronous stateful computation: it
works with a state `S` and asynchronously produces a result `A` and a
new state `S`.

**Without FP:**

``` dart
class AppState {
  int loadingCount = 0;

  Future<void> loadData() async {
    loadingCount++;

    await Future.delayed(const Duration(seconds: 1));

    loadingCount--;
  }
}
```

**With FP --- conceptual shape:**

``` dart
// StateAsync<S, A> models an asynchronous
// computation that transforms state S and
// produces a result A.
//
// Check the installed fpdart version's API
// documentation for the exact constructors
// and execution methods.
```

Learn ordinary `State` first. Async state abstractions are more advanced
and are not needed in every Flutter project.

## 21. IOOption

`IOOption<A>` represents a synchronous effect that may produce a value
or no value.

**Without FP:**

``` dart
String? readSetting(String key) {
  return localStorage.get(key);
}
```

**With fpdart:**

``` dart
IOOption<String> readSetting(String key) {
  return IOOption(() {
    return Option.fromNullable(localStorage.get(key));
  });
}
```

Usage:

``` dart
final result = readSetting('username').run();

result.match(
  () => print('Setting missing'),
  (value) => print(value),
);
```

This is useful when reading a synchronous source where a value may be
absent.

## 22. IORef

`IORef<A>` provides a mutable reference whose read/write operations are
represented as `IO` effects.

**Without FP:**

``` dart
var counter = 0;

counter++;
counter++;

print(counter); // 2
```

**With fpdart (illustrative API):**

``` dart
final counter = IORef.of(0);

final initial = counter.read();
final update = counter.write(10);
final result = counter.read();
```

The exact construction and execution API can vary by fpdart version.
Check the installed version's `IORef` API before using this snippet.

`IORef` is useful when controlled mutation is necessary. FP does not
mean mutation can never exist.

## 23. Predicate

A `Predicate<A>` represents a condition that returns `true` or `false`
for a value. Predicates can be composed.

**Without FP:**

``` dart
bool isAdult(int age) => age >= 18;
bool hasPermission(int age) => age >= 21;

bool canAccess(int age) {
  return isAdult(age) && hasPermission(age);
}
```

**With fpdart:**

``` dart
final isAdult = Predicate<int>((age) => age >= 18);
final hasPermission = Predicate<int>((age) => age >= 21);

final canAccess = isAdult & hasPermission;

print(canAccess(25)); // true
print(canAccess(19)); // false
```

Predicates are useful for validation, authorization rules, and filtering
conditions.

## 24. Unit

`Unit` represents a value that carries no meaningful information. Unlike
`void`, it is an actual value and can be used in type-safe functional
compositions.

**Without FP:**

``` dart
void saveUser() {
  print('User saved');
}
```

**With fpdart:**

``` dart
Unit saveUser() {
  print('User saved');
  return unit;
}
```

Use `Unit` when a computation needs to return a value even though there
is no meaningful result to provide.

## 25. Do Notation

Do notation lets you write sequential-looking code while composing
operations that may fail or produce no value.

**Without FP:**

``` dart
Market? getMarket() => /* ... */ null;

final market = getMarket();

if (market == null) return;

final banana = market.buyBanana();
if (banana == null) return;

final apple = market.buyApple();
if (apple == null) return;

final pear = market.buyPear();
if (pear == null) return;

print('$banana $apple $pear');
```

**With fpdart `Option.Do`:**

``` dart
final result = Option.Do(($) {
  final market = $(getMarket());

  final banana = $(market.buyBanana());
  final apple = $(market.buyApple());
  final pear = $(market.buyPear());

  return '$banana $apple $pear';
});

print(result.getOrElse(() => 'Purchase failed'));
```

Here, `getMarket`, `buyBanana`, `buyApple`, and `buyPear` are assumed to
return `Option` values. If any returns `None`, the computation
short-circuits.

Do notation is available for several fpdart types, including `Option`,
`Either`, and `TaskEither`.

## 26. Functor

A Functor is a structure that supports mapping a function over a value
inside a context.

For example, `Option` is a Functor.

**Without FP:**

``` dart
int? number = 10;

final result = number == null
    ? null
    : number * 2;
```

**With fpdart:**

``` dart
final Option<int> number = Some(10);

final result = number.map((value) => value * 2);
// Some(20)
```

The transformation happens inside the `Option` context. You do not need
to manually extract the value.

## 27. Monad

A Monad is an abstraction that supports chaining computations where each
next computation returns a value in the same context.

For example, `Option` supports `flatMap`.

**Without FP:**

``` dart
int? parseNumber(String value) {
  return int.tryParse(value);
}

int? doubleNumber(int number) {
  return number * 2;
}

final parsed = parseNumber('10');

final result = parsed == null
    ? null
    : doubleNumber(parsed);
```

**With fpdart:**

``` dart
Option<int> parseNumber(String value) {
  return Option.fromNullable(int.tryParse(value));
}

Option<int> doubleNumber(int number) {
  return Some(number * 2);
}

final result = parseNumber('10')
    .flatMap(doubleNumber);
```

A useful starting point: - Functor → `map` - Monad → `flatMap` (also
commonly called `bind`)

These are abstractions describing behaviour, not separate objects you
must instantiate.

## 28. Immutable Collections

FP encourages avoiding mutation of collections. Create new collections
when transforming data.

**Without FP:**

``` dart
final numbers = [1, 2, 3];

numbers.add(4);
numbers.remove(1);
```

The original list changes.

**With FP-style transformations:**

``` dart
final numbers = [1, 2, 3];

final added = [...numbers, 4];

final removed = added
    .where((number) => number != 1)
    .toList();
```

The original list remains unchanged.

For stronger immutable collection guarantees, consider the separate
`fast_immutable_collections` package. A `final` list reference alone
does not make the list's contents immutable.

## 29. IORef: Creating and Using a Reference

This final example revisits `IORef` to emphasize the distinction between
describing an effect and executing it.

**Without FP:**

``` dart
var counter = 0;

counter = 10;

print(counter); // 10
```

**With fpdart:**

``` dart
final createCounter = IORef.create(0);

final counter = createCounter.run();

counter.write(10).run();

final result = counter.read().run();

print(result); // 10
```

Check the API for your installed fpdart version before relying on this
exact snippet. The central idea is that reading and writing are
represented as `IO` operations.

------------------------------------------------------------------------

# Quick Reference

  Concept                  Main purpose
  ------------------------ --------------------------------------------
  Pure Functions           Predictable functions
  Immutability             Avoid modifying existing data
  Higher-Order Functions   Pass functions as arguments or return them
  First-Class Functions    Treat functions as values
  Composition              Combine smaller functions
  Map                      Transform collection values
  Where                    Filter collection values
  Fold                     Reduce a collection to one result
  Option                   Handle missing values
  Either                   Represent success or failure
  FlatMap                  Chain contextual computations
  IO                       Represent synchronous side effects
  Task                     Represent asynchronous computations
  TaskEither               Async computations with typed failures
  IOEither                 Sync computations with typed failures
  TaskOption               Async computations with optional results
  Reader                   Supply dependencies/environment
  ReaderTaskEither         Dependencies + async work + typed failures
  State                    Model state transitions
  StateAsync               Model asynchronous state transitions
  IOOption                 Synchronous effect with an optional result
  IORef                    Represent controlled mutable references
  Predicate                Represent and compose conditions
  Unit                     Represent an empty/meaningless result
  Do notation              Make functional chains easier to read
  Functor                  Map over a value inside a context
  Monad                    Chain computations inside a context
  Immutable Collections    Avoid mutating collection contents

# Recommended Learning Order

Do not try to memorize every abstraction at once. Learn them in this
order:

1.  **Foundation:** pure functions, immutability, higher-order
    functions, and composition.
2.  **Collection operations:** `map`, `where`, `fold`, and predicates.
3.  **Core FP types:** `Option`, `Either`, `map`, `flatMap`, Functor,
    and Monad.
4.  **Side effects:** `IO`, `Task`, `IOEither`, and `TaskEither`.
5.  **Advanced abstractions:** `Reader`, `State`, Do notation, and
    `ReaderTaskEither`.
6.  **Real Flutter application:** build an API repository,
    authentication flow, error handling, and dependency management.

## What to Prioritize in Production Flutter Projects

You do not need every fpdart abstraction in every project. Prioritize:

-   `Option` for values that may be absent.
-   `Either` for explicit success/failure results.
-   `TaskEither` for asynchronous operations that can fail.
-   `map` and `flatMap` for composing transformations.
-   Immutability for predictable application state.
-   Pure functions for business logic that is easy to test.

## Official Resources

-   [fpdart on pub.dev](https://pub.dev/packages/fpdart)
-   [fpdart API
    documentation](https://pub.dev/documentation/fpdart/latest/)
-   [fpdart GitHub repository](https://github.com/SandroMaglione/fpdart)

> **Version note:** Check the package version in your own `pubspec.lock`
> and use the matching API documentation. Some advanced abstractions and
> constructors can differ between releases. Treat the examples marked
> illustrative or conceptual as learning guides, not guaranteed
> copy-paste code.
