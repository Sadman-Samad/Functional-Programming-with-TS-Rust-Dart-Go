# Rust Functional Programming

## 29 Concepts: Traditional Rust vs Functional Programming in Rust

This guide compares traditional imperative Rust with functional
programming (FP) in Rust.

Rust is unusual among the languages in this series (Dart/fpdart,
TypeScript/Effect, Go/fp-go): **its standard library already ships the
core FP types**. `Option<T>` and `Result<T, E>` are monadic
(think `Option`/`Either` from fp-ts), the `Iterator` trait is a lazy
functor with `map` / `filter` / `fold` / `flat_map`, and the `?`
operator gives you do-notation for both `Option` and `Result`
*without* a third-party library. Most production FP Rust uses the
standard library plus two crates:

- **[`itertools`](https://crates.io/crates/itertools)** — extra iterator
  combinators (`chunk_by`, `dedup`, `interleave`, `group_by`, ...).
- **[`im`](https://crates.io/crates/im)** — persistent immutable
  collections (`Vector`, `HashMap`, `HashSet`, `OrdMap`) with structural
  sharing.

For the abstract concepts that std does not provide (`Reader`, `State`
as a monad, `IO` as a type), I show the **idiomatic Rust pattern** that
achieves the same goal — usually a closure, a function pointer, or a
`Cell`/`RefCell`/`Mutex` — and mention
**[`fp-core.rs`](https://github.com/JasonShin/fp-core.rs)** /
**[`frunk`](https://github.com/lloydmeta/frunk)** where they genuinely
add something std doesn't have.

> **Important:** FP is a programming style, not a requirement to use a
> particular crate. Rust is multi-paradigm. The community deliberately
> chose to build monadic combinators into `Option` / `Result` /
> `Iterator` rather than ship a separate FP crate, so idiomatic FP Rust
> *is* std-library FP. Resist the urge to wrap everything in a
> third-party monad library — you'll fight the borrow checker and the
> ecosystem at the same time.

The examples use common application scenarios: transforming data,
handling missing values, managing API failures, and composing
synchronous and asynchronous operations.

---

## Installation

```toml
# Cargo.toml
[dependencies]
itertools = "0.13"   # extra iterator combinators
im       = "15"      # persistent immutable collections
# Optional, for advanced abstract concepts only:
# fp-core  = "0.1"   # Fantasy Land types — unmaintained since 2019
# frunk    = "0.4"   # HList, Coproduct, Validated — actively maintained
```

```rust
use itertools::Itertools;
use im::Vector as IVector;
// std FP types — no extra imports needed:
// Option, Result, Iterator, Future are all in the prelude.
```

> **Version note:** This guide targets Rust **1.75+** (uses
> `async fn` in traits, `let-else`, and stable closure capture
> features). `itertools` 0.13 and `im` 15 are current as of writing;
> check `crates.io` for newer versions. Treat examples flagged
> "illustrative" as learning guides — Rust's compiler is the final
> arbiter.

---

## 1. Pure Functions

A pure function:

- Returns the same output for the same input.
- Does not modify external state.
- Does not perform side effects.

**Without FP:**

```rust
fn calculate_discount(price: i64) -> i64 {
    println!("Calculating discount"); // Side effect
    price - (price * 10 / 100)
}
```

**With FP:**

```rust
fn calculate_discount(price: i64) -> i64 {
    price - (price * 10 / 100)
}

fn main() {
    let result = calculate_discount(1000);
    println!("{result}"); // 900
}
```

The second function is pure: it only calculates and returns a value.

**When to use:** Pure functions are the right default for business logic —
price calculations, validation, transformation. Push side effects
(`println!`, file IO, network) to `main` or to the boundary of your
application (HTTP handlers, CLI entry points).

**Common pitfall:** A function that takes `&mut Vec<T>` and pushes to it
is **not** pure, even if it returns nothing. The mutation is a side
effect visible to the caller.

---

## 2. Immutability

Immutability means not changing existing data. In Rust, **all `let`
bindings are immutable by default** — you opt into mutation with `let
mut`. This is one of the language's biggest FP wins.

**Without FP (opting into mutation):**

```rust
struct User {
    name: String,
    age: u32,
}

fn main() {
    let mut user = User { name: "John".into(), age: 25 };
    user.age = 26; // Mutates the existing value
}
```

**With FP (immutability by default):**

```rust
struct User {
    name: String,
    age: u32,
}

impl User {
    fn with_age(self, age: u32) -> User {
        User { age, ..self } // Struct update syntax consumes self, returns new value
    }
}

fn main() {
    let user = User { name: "John".into(), age: 25 };
    let updated = user.with_age(26); // user is gone (moved), updated is fresh
    // println!("{user:?}"); // error: value borrowed after move
    println!("{updated:?}");
}
```

Rust's **move semantics** make immutability near-free: instead of
copying the whole struct, the `..self` syntax reuses fields by moving
them. If you need to keep the original *and* produce a new value,
`#[derive(Clone)]` and call `.clone()`.

**Common pitfall:** `let mut v = vec![1, 2, 3];` lets you mutate the
vector. If you want to enforce immutability across API boundaries, take
`&[T]` (a shared slice) instead of `&mut Vec<T>` — the type system
prevents the caller from mutating.

---

## 3. Higher-Order Functions

A higher-order function accepts another function as an argument or
returns a function.

**Without FP:**

```rust
fn calculate(a: i64, b: i64) -> i64 { a + b }

fn execute(a: i64, b: i64) -> i64 {
    calculate(a, b)
}
```

**With FP:**

```rust
fn execute(a: i64, b: i64, op: impl Fn(i64, i64) -> i64) -> i64 {
    op(a, b)
}

fn main() {
    let sum = execute(10, 20, |a, b| a + b);
    let product = execute(10, 20, |a, b| a * b);
    println!("{sum} {product}"); // 30 200
}
```

Rust's three closure traits matter for higher-order functions:

- `FnOnce` — consumes captured variables. Can be called once.
- `FnMut` — mutates captured variables. Can be called repeatedly.
- `Fn` — only borrows captured variables. Can be called repeatedly,
  shared across threads.

Use the **least powerful trait that works**: `Fn` when possible,
`FnMut` when you need to mutate, `FnOnce` only when consuming.

**Common pitfall:** Returning a closure that captures by reference
requires a lifetime annotation: `fn make_adder<'a>(x: &'a i64) -> impl Fn(i64) -> i64 + 'a`. If you don't need the borrow, capture by move:
`move |y| x + y`.

---

## 4. First-Class Functions

Rust functions are first-class values. You can store them in variables,
pass them to other functions, and return them. Function items (defined
with `fn`) coerce to function pointers (`fn(A) -> B`) when needed.

**Without FP / method-style:**

```rust
struct Calculator;

impl Calculator {
    fn add(&self, a: i64, b: i64) -> i64 { a + b }
}

fn main() {
    let c = Calculator;
    let result = c.add(10, 20);
}
```

**With FP:**

```rust
fn add(a: i64, b: i64) -> i64 { a + b }

fn main() {
    // Function item coerces to function pointer
    let op: fn(i64, i64) -> i64 = add;
    let result = op(10, 20);
}
```

Both styles are valid. The functional version treats the behaviour
itself as a value, which makes it trivial to compose and test.
Function pointers are zero-allocation and `Copy`, so they're cheaper
than closures — use them when you don't need to capture state.

---

## 5. Function Composition

Rust does not ship a built-in `compose` / `pipe` function in std. The
idiomatic composition patterns are:

- **Method chaining** — works when types share a trait (e.g. `Iterator`
  methods).
- **Nested calls** — `f(g(h(x)))`, the math-textbook style.
- **A small helper** — define `pipe` / `compose` once in a utility
  module and reuse.

**Without FP:**

```rust
fn process_name(name: &str) -> String {
    let trimmed = name.trim();
    let upper = trimmed.to_uppercase();
    format!("Hello {upper}")
}
```

**With FP-style composition:**

```rust
fn pipe<T, U, V, F, G>(value: T, f: F, g: G) -> V
where
    F: FnOnce(T) -> U,
    G: FnOnce(U) -> V,
{
    g(f(value))
}

fn main() {
    let trim = |s: &str| s.trim();
    let upper = |s: &str| s.to_uppercase();
    let greet = |s: String| format!("Hello {s}");

    // Note: we bridge &str and String explicitly because Rust doesn't
    // do implicit coercion inside closures.
    let result = pipe(
        " John ",
        |s| s.trim().to_uppercase(),
        |s| format!("Hello {s}"),
    );
    println!("{result}"); // Hello JOHN
}
```

For longer pipelines, define `pipe2`, `pipe3`, etc., or — more
idiomatically — reach for **method chaining on `Iterator`** whenever
the data flow is over a sequence:

```rust
let result = " John "
    .trim()
    .to_uppercase()
    .split_whitespace()
    .map(|w| format!("Hello {w}"))
    .collect::<Vec<_>>();
```

**Common pitfall:** Rust closures don't unify around `&str` vs `String`
the way TypeScript arrow functions unify around `string`. You will see
type-mismatch errors at composition boundaries — fix them by being
explicit about ownership, not by sprinkling `.to_string()`.

---

## 6. Map

`Iterator::map` transforms every item in a lazy sequence.
`Option::map` and `Result::map` do the same inside a context.

**Without FP (imperative loop):**

```rust
let numbers = vec![1, 2, 3, 4];
let mut results = Vec::with_capacity(numbers.len());
for n in &numbers {
    results.push(n * 2);
}
```

**With FP:**

```rust
let numbers = vec![1, 2, 3, 4];

let results: Vec<i64> = numbers.iter().map(|&n| n * 2).collect();
// [2, 4, 6, 8]
```

`Iterator` is **lazy** — `map` returns a `Map` struct that does no work
until you `.collect()` / `.for_each()` / `.fold()` it. This is the same
discipline as Effect's `Effect` or fp-go's `IO`: describe first, run
later.

`Option::map` and `Result::map` are also in the prelude:

```rust
let some: Option<i64> = Some(10);
let doubled: Option<i64> = some.map(|n| n * 2); // Some(20)

let ok: Result<i64, &str> = Ok(10);
let doubled: Result<i64, &str> = ok.map(|n| n * 2); // Ok(20)
```

---

## 7. Where / Filter

`Iterator::filter` keeps only the elements that satisfy a predicate.

**Without FP (imperative loop):**

```rust
let numbers = vec![1, 2, 3, 4, 5, 6];
let mut results = Vec::new();
for n in &numbers {
    if n % 2 == 0 {
        results.push(*n);
    }
}
```

**With FP:**

```rust
let numbers = vec![1, 2, 3, 4, 5, 6];

let results: Vec<i64> = numbers.iter().copied().filter(|n| n % 2 == 0).collect();
// [2, 4, 6]
```

`filter` takes a predicate `FnMut(&Self::Item) -> bool`. The closure
receives a reference — use `.copied()` / `.cloned()` upstream if you'd
rather work with values.

`Option::filter` is also available — handy for "return `Some(x)` only if
`x` satisfies this condition, else `None`":

```rust
let adult = Some(25).filter(|&age| age >= 18); // Some(25)
let minor = Some(15).filter(|&age| age >= 18); // None
```

---

## 8. Fold / Reduce

`Iterator::fold` processes a collection and combines its elements into
one result. `Iterator::reduce` is the no-seed variant (panics on empty).

**Without FP:**

```rust
let prices = vec![100, 200, 300];
let mut total = 0;
for p in &prices {
    total += p;
}
```

**With FP:**

```rust
let prices = vec![100, 200, 300];

let total: i64 = prices.iter().sum();
// Or, for a non-trivial accumulator:
let total: i64 = prices.iter().fold(0, |sum, &p| sum + p);
// 600
```

`fold` takes a seed (`0`) and a closure `FnMut(Acc, Item) -> Acc`. The
first call gets the seed; each subsequent call gets the previous
accumulator value.

For sum / product specifically, `Iterator::sum` and `Iterator::product`
are special-cased and faster than `fold` — reach for them when the
operation is a plain arithmetic sum.

---

## 9. Option

`Option<T>` represents a value that may or may not exist. It has two
variants:

- `Some(value)`: a value exists.
- `None`: no value exists.

`Option` is in the std prelude — no import needed.

**Without FP:**

```rust
struct User { name: String, age: u32 }

fn find_user(id: u32) -> Option<User> {
    if id == 1 {
        Some(User { name: "John".into(), age: 25 })
    } else {
        None
    }
}

fn main() {
    let user = find_user(10);
    if let Some(u) = user {
        println!("{}", u.name);
    } else {
        println!("User not found");
    }
}
```

**With FP combinators:**

```rust
fn main() {
    let user = find_user(10);

    let message = user
        .map(|u| u.name)
        .unwrap_or_else(|| "User not found".to_string());

    println!("{message}");
}
```

`Option<User>` makes the possibility of a missing user explicit in the
return type — callers cannot accidentally dereference a null. The
combinator surface is rich:

- `map(f)` — transform the inner value.
- `and_then(f)` — flatMap (chain `Option`-returning functions).
- `unwrap_or(default)` — provide a fallback value.
- `unwrap_or_else(f)` — fallback computed lazily.
- `ok_or(err)` — convert `Option<T>` to `Result<T, E>`.
- `filter(p)` — keep `Some(x)` only if `p(x)` is true.
- `is_some()`, `is_none()` — guards.
- `as_ref()`, `as_mut()` — convert `Option<T>` to `Option<&T>` /
  `Option<&mut T>` without consuming.

**Common pitfall:** `Option::unwrap` panics on `None`. In production
code use `unwrap_or`, `unwrap_or_default`, `ok_or`, or `?` (in a
function that returns `Option`). Reserve `unwrap()` for tests and
prototypes.

---

## 10. Result

`Result<T, E>` represents either a success (`Ok(T)`) or a typed failure
(`Err(E)`). It is Rust's answer to exceptions — and unlike Go's `(T,
error)`, the error is in the type, not in a separate return slot.

**Without FP:**

```rust
fn fetch_user() -> Result<User, FetchError> {
    let user = api.get_user()?; // ? propagates Err
    Ok(user)
}
```

Wait — that's already idiomatic FP Rust. The point of comparison is
against `panic` / exceptions:

**Without FP (panicking):**

```rust
fn fetch_user() -> User {
    let user = api.get_user().expect("API failed"); // Panics on error
    user
}
```

The caller cannot recover from a panic.

**With FP:**

```rust
#[derive(Debug)]
enum FetchError { Network(String), NotFound }

fn fetch_user() -> Result<User, FetchError> {
    let user = api.get_user()
        .map_err(|e| FetchError::Network(e.to_string()))?;
    Ok(user)
}

fn main() {
    match fetch_user() {
        Ok(u) => println!("User: {}", u.name),
        Err(FetchError::Network(msg)) => println!("Network error: {msg}"),
        Err(FetchError::NotFound) => println!("Not found"),
    }
}
```

The caller handles both outcomes explicitly — exhaustively, because
`match` on an `enum` is checked at compile time.

`Result`'s combinator surface mirrors `Option`:

- `map(f)` — transform `Ok` value.
- `map_err(f)` — transform `Err` value.
- `and_then(f)` — flatMap (chain `Result`-returning functions).
- `unwrap_or`, `unwrap_or_else`, `unwrap_or_default` — fallbacks.
- `ok()`, `err()` — convert to `Option`.
- `is_ok()`, `is_err()` — guards.

**When to use `Option` vs `Result`:** `Option` is for "value or no
value" with no error semantics (a lookup, an optional config field).
`Result` is for "success or a specific failure" (an HTTP call, a parse,
a validation). Use `Option::ok_or(err)` to bridge when a lookup miss
should become a typed error.

---

## 11. Map vs and_then (FlatMap)

These two combinators are central to working with values inside
contexts like `Option` and `Result`:

- `map(f)` transforms the inner value with a function that returns a
  **plain value**.
- `and_then(f)` (Rust's name for `flatMap` / `bind`) transforms the
  inner value with a function that returns **another `Option` /
  `Result`**, avoiding nested containers.

**Without FP:**

```rust
fn parse_age(value: &str) -> Option<u32> {
    value.parse().ok()
}

fn validate_age(age: u32) -> Option<u32> {
    if age < 18 { None } else { Some(age) }
}

fn main() {
    let parsed = parse_age("25");
    let result = parsed.and_then(validate_age);
    // Or, the long way:
    // let result = match parsed {
    //     Some(age) => validate_age(age),
    //     None => None,
    // };
    println!("{result:?}"); // Some(25)
}
```

**With FP combinators:**

```rust
let result: Option<u32> = parse_age("25").and_then(validate_age);
// Some(25)

let doubled: Option<u32> = parse_age("25").map(|age| age * 2);
// Some(50)
```

If you `map` a function that returns an `Option`, you get
`Option<Option<T>>` — almost always a bug. Reach for `and_then`
instead. Same goes for `Result::and_then`.

**Naming note:** Rust deliberately chose `and_then` over `flat_map`
because it reads better in chains:
`s.parse().ok().and_then(validate).and_then(enrich)`. The semantics are
identical to Haskell's `>>=` / fp-ts's `chain` / fp-go's `Chain`.

---

## 12. IO

Rust **does not** ship an `IO` monad. The language's design philosophy
is that side effects are tracked by the type system through ownership
and lifetimes, not by wrapping effects in a container.

The idiomatic Rust approach for "describe an effect, run it later" is
**a closure**:

**Without FP:**

```rust
fn read_username() -> String {
    println!("Reading username");
    "John".to_string()
}

fn main() {
    let username = read_username(); // Runs immediately
}
```

**With FP-style lazy effect:**

```rust
fn main() {
    // Describes the computation; nothing runs yet
    let read_username: impl FnOnce() -> String = || {
        println!("Reading username");
        "John".to_string()
    };

    let username = read_username(); // Executes it
    println!("{username}");
}
```

This is the same pattern fp-go uses for `IO[A]` (which is just
`func() A`), and it's idiomatic in Rust — closures are zero-cost when
monomorphised. For a richer IO abstraction (with error channels,
retry, resource management), reach for `tokio::task` and async/await
rather than a monad library.

**fp-core.rs** ships an `io::IO` type if you genuinely need a typed IO
monad:

```rust
// fp-core (illustrative — check the crate for current API)
use fp_core::io::IO;

let read_username: IO<String> = IO::new(|| {
    println!("Reading username");
    "John".to_string()
});

let username = read_username.run(); // Executes
```

**Common pitfall:** fp-core.rs hasn't seen a release since 2019. Its
API surface is unstable. For new projects, prefer idiomatic closures +
`tokio` — the Rust ecosystem has collectively decided that's the
right layer.

---

## 13. Async (Future / Task)

`Future<Output = T>` is Rust's async-computation type — the equivalent
of fp-ts's `Task<A>` or fp-go's conceptual `Task`. The `async` /
`await` syntax desugars to `Future` combinators.

**Without FP (synchronous blocking):**

```rust
fn fetch_user() -> User {
    api.get_user_blocking()
}
```

**With async:**

```rust
async fn fetch_user() -> User {
    api.get_user_async().await
}

#[tokio::main]
async fn main() {
    let user = fetch_user().await;
}
```

`async fn` desugars to `fn fetch_user() -> impl Future<Output = User>`.
Calling `fetch_user()` only builds the future; `.await` polls it to
completion. This is the same describe-vs-run split as `Effect.promise`
or fp-go's `IOResult`.

You can compose futures:

```rust
async fn double() -> i64 { 10 }

async fn program() -> i64 {
    let n = double().await;
    n * 2
}

#[tokio::main]
async fn main() {
    let result = program().await; // 20
}
```

A `Future` by itself does not model typed failures — the `Output` type
can be `Result<T, E>` if you need that. See concept 14.

**Common pitfall:** `Future` is **lazy** — creating one without
`.await`ing or spawning it does nothing. `tokio::spawn(fut)` is the
way to run a future in the background.

---

## 14. Async with typed error (TaskEither)

The Rust equivalent of fp-ts's `TaskEither<L, R>` is simply
`async fn() -> Result<T, E>` — or equivalently
`impl Future<Output = Result<T, E>>`. The error channel and the async
channel compose naturally because `?` works inside `async fn`.

**Without FP (panicking / unwrap):**

```rust
async fn fetch_user() -> User {
    api.get_user().await.expect("API failed") // Panics on error
}
```

**With FP:**

```rust
#[derive(Debug)]
enum FetchError { Network(String) }

async fn fetch_user() -> Result<User, FetchError> {
    let user = api.get_user()
        .await
        .map_err(|e| FetchError::Network(e.to_string()))?;
    Ok(user)
}

#[tokio::main]
async fn main() {
    match fetch_user().await {
        Ok(u) => println!("User: {}", u.name),
        Err(FetchError::Network(msg)) => println!("Failed: {msg}"),
    }
}
```

The caller handles both outcomes explicitly. The `?` operator
propagates errors through `Result` (and through `Option`) inside
`async fn` exactly as it does in sync functions — this is Rust's
genuine FP ergonomics win, equivalent to do-notation in Haskell or
`Effect.gen` in TypeScript.

For concurrent fetching with parallel error handling, use `tokio::try_join!`:

```rust
async fn fetch_user_and_posts() -> Result<(User, Vec<Post>), FetchError> {
    let (user, posts) = tokio::try_join!(
        fetch_user(),
        fetch_posts(),
    )?;
    Ok((user, posts))
}
```

`try_join!` runs the futures concurrently and short-circuits on the
first `Err` — the same semantics as `Effect.all` with the default
error mode.

---

## 15. Sync with typed error (IOEither)

For a synchronous computation that can fail with a typed error, Rust
just uses `Result<T, E>` — there's no separate `IOEither` type. The
combinators live on `Result` directly.

**Without FP:**

```rust
fn read_configuration() -> i64 {
    let n: i64 = "123".parse().expect("Invalid configuration"); // Panics
    n
}
```

**With FP:**

```rust
#[derive(Debug)]
enum ConfigError { Invalid(String) }

fn read_configuration() -> Result<i64, ConfigError> {
    "123".parse::<i64>().map_err(|e| ConfigError::Invalid(e.to_string()))
}

fn main() {
    match read_configuration() {
        Ok(v) => println!("Value: {v}"),
        Err(ConfigError::Invalid(msg)) => println!("Error: {msg}"),
    }
}
```

`Result` *is* the `IOEither` type — there's no need to compose `IO`
with `Either` because Rust's type system already tracks effectfulness
through ownership. The rule of thumb:

- Sync + can fail → `Result<T, E>`.
- Async + can fail → `async fn() -> Result<T, E>` or
  `impl Future<Output = Result<T, E>>`.

---

## 16. Async Option (TaskOption)

Async computation whose result may be absent: `async fn() ->
Option<T>` or `impl Future<Output = Option<T>>`.

**Without FP:**

```rust
async fn find_user(id: u32) -> Option<User> {
    database.find_user(id).await
}
```

That's already idiomatic — there's no improvement to make. The
question is how to **compose** it.

**With FP combinators:**

```rust
async fn find_user(id: u32) -> Option<User> { /* ... */ }

async fn get_user_name(id: u32) -> Option<String> {
    // Equivalent of TaskOption.flatMap
    let user = find_user(id).await?;
    Some(user.name)
}

#[tokio::main]
async fn main() {
    if let Some(name) = get_user_name(10).await {
        println!("{name}");
    } else {
        println!("User not found");
    }
}
```

The `?` operator on `Option<T>` early-returns `None` if the
expression yields `None` — exactly what `Option.flatMap` does in
fp-ts/Effect, but written in line.

**When to model absence as `Option` vs as a `Result` with a typed
"NotFound" variant:** If absence is a normal control-flow outcome
(looking up an optional value), use `Option`. If absence is an error
that callers need to distinguish from other failures, model it as a
`NotFound` variant in your error enum and return `Result`.

---

## 17. Reader

`Reader<E, A>` represents a computation that needs an environment or
dependency `E` to produce a value `A`. Rust doesn't ship a `Reader`
type — the idiomatic equivalent is a **closure that takes the
environment by reference**:

**Without FP (constructor injection):**

```rust
trait ApiClient {
    fn get_user(&self) -> User;
}

struct UserRepository<'a> {
    api: &'a dyn ApiClient,
}

impl<'a> UserRepository<'a> {
    fn get_user(&self) -> User {
        self.api.get_user()
    }
}

fn main() {
    let api = MyApiClient;
    let repo = UserRepository { api: &api };
    let user = repo.get_user();
}
```

**With FP (closure-as-reader):**

```rust
trait ApiClient {
    fn get_user(&self) -> User;
}

// A "Reader" is a closure that takes the environment and returns a value.
type Reader<'a, E, A> = &'a dyn Fn(&E) -> A;

fn make_get_user() -> impl Fn(&dyn ApiClient) -> User {
    |api: &dyn ApiClient| api.get_user()
}

fn main() {
    let api = MyApiClient;
    let get_user = make_get_user();
    let user = get_user(&api);
    println!("{}", user.name);
}
```

The closure `make_get_user` returns a function that depends on an
`ApiClient` — that's the `Reader` monad, expressed idiomatically.
This is useful for passing dependencies explicitly and making
computations easy to test (call `get_user(&fake_api)` in tests).

For more structured dependency injection with traits and generics
(the more common Rust pattern), define a trait, implement it for your
production type and a test double, and use it as a generic bound:

```rust
fn get_user<A: ApiClient>(api: &A) -> User {
    api.get_user()
}
```

This compiles down to monomorphised, zero-cost code — strictly better
than runtime trait objects when the dependency is known at compile
time.

---

## 18. Reader + async + typed error (ReaderTaskEither)

`ReaderTaskEither<E, L, R>` combines dependency injection, async work,
and typed errors. In Rust this is just an `async fn` that takes the
dependency by reference and returns `Result`:

**Without FP:**

```rust
async fn get_user(api: &dyn ApiClient) -> Result<User, FetchError> {
    let user = api.get_user().await
        .map_err(|e| FetchError::Network(e.to_string()))?;
    Ok(user)
}
```

Wait — that *is* the idiomatic FP version. There's no monad library
to bring in. The point of comparison is against hand-managed
dependency + async + try/catch:

**Without FP (untyped panics):**

```rust
async fn get_user(api: &dyn ApiClient) -> User {
    api.get_user().await.expect("API failed")
}
```

**With FP (typed errors, dependency as parameter):**

```rust
#[derive(Debug)]
enum FetchError { Network(String) }

async fn get_user(api: &impl ApiClient) -> Result<User, FetchError> {
    let user = api.get_user()
        .await
        .map_err(|e| FetchError::Network(e.to_string()))?;
    Ok(user)
}

#[tokio::main]
async fn main() {
    let api = MyApiClient;
    match get_user(&api).await {
        Ok(u) => println!("User: {}", u.name),
        Err(FetchError::Network(msg)) => println!("Failed: {msg}"),
    }
}
```

Usage at a call site that needs to compose multiple
dependency-async-error steps:

```rust
async fn get_user_and_posts(
    api: &impl ApiClient,
    posts_api: &impl PostsClient,
) -> Result<(User, Vec<Post>), FetchError> {
    let user = get_user(api).await?;
    let posts = posts_api.get_posts(&user.id)
        .await
        .map_err(|e| FetchError::Network(e.to_string()))?;
    Ok((user, posts))
}
```

The `?` operator threads errors through, just like
`ReaderTaskEither`'s `flatMap` / `bind`. No external library needed.

---

## 19. State

`State<S, A>` represents a computation that receives state `S` and
produces a result `A` alongside a new state `S`. In Rust the idiomatic
equivalent is **a function `Fn(S) -> (S, A)`**.

**Without FP:**

```rust
struct Wallet { balance: i64 }

impl Wallet {
    fn add_money(&mut self, amount: i64) {
        self.balance += amount;
    }
}

fn main() {
    let mut wallet = Wallet { balance: 100 };
    wallet.add_money(50);
    println!("{}", wallet.balance); // 150
}
```

**With FP-style pure state:**

```rust
fn add_money(balance: i64, amount: i64) -> (i64, i64) {
    let new_balance = balance + amount;
    (new_balance, new_balance) // (new_state, result)
}

fn main() {
    let (new_state, result) = add_money(100, 50);
    println!("State: {new_state}, Result: {result}"); // State: 150, Result: 150
}
```

The original input state is not mutated. The computation returns the
next state and a result. Chaining is manual (Rust has no built-in
`State` monad `bind`):

```rust
// Chain add_money(50) then add_money(30)
let (s1, _) = add_money(100, 50);
let (s2, _) = add_money(s1, 30);
println!("{s2}"); // 180
```

For a typed `State` monad with `bind`, `map`, `get`, `put`, **fp-core.rs**
ships `State<S, A>`:

```rust
// fp-core (illustrative — check the crate for current API)
use fp_core::state::State;

let add_money: State<i64, i64> = State::new(|balance: i64| {
    let new_balance = balance + 50;
    (new_balance, new_balance)
});

let (state, result) = add_money.run(100);
```

**When to use:** `State` is the right model when you have a
*deterministic* state machine that's pure (parsers, simulations,
small games). For shared mutable state across threads, use
`Mutex<T>` / `RwLock<T>` / atomics — see concept 22.

---

## 20. StateAsync

Async state transitions: a function `Fn(S) -> impl Future<Output =
(S, A)>`. In practice, Rust async state is usually modelled with
`Mutex<T>` (or `tokio::sync::Mutex<T>` for async-aware locking) rather
than as a state monad.

**Without FP:**

```rust
struct AppState { loading_count: u32 }

impl AppState {
    async fn load_data(&mut self) -> Result<(), Error> {
        self.loading_count += 1;
        tokio::time::sleep(Duration::from_secs(1)).await;
        self.loading_count -= 1;
        Ok(())
    }
}
```

**With FP-style async state (tokio Mutex):**

```rust
use tokio::sync::Mutex;
use std::sync::Arc;

struct AppState { loading_count: u32 }

async fn load_data(state: Arc<Mutex<AppState>>) -> Result<(), Error> {
    {
        let mut s = state.lock().await;
        s.loading_count += 1;
    } // lock released here

    tokio::time::sleep(Duration::from_secs(1)).await;

    {
        let mut s = state.lock().await;
        s.loading_count -= 1;
    }
    Ok(())
}
```

The lock is acquired, mutated, and released in the smallest scope
possible — concurrent callers can read the state between operations.
This is the practical Rust answer to async state: locks (or atomics)
plus `async`/`await`, not a state monad.

Learn ordinary `State` first. Async state abstractions via locks are
the production Rust pattern; async state monads exist in fp-core but
are not widely used.

---

## 21. Sync Option combinators (IOOption)

For a synchronous effect that may produce a value or none (fp-ts's
`IOOption<A>`), the idiomatic Rust shape is a closure `FnOnce() ->
Option<T>` — or simply a function that returns `Option<T>`.

**Without FP:**

```rust
fn read_setting(key: &str) -> Option<String> {
    std::env::var(key).ok()
}
```

That's already idiomatic. The interesting case is **composing** such
operations with `?`:

**With FP combinators:**

```rust
fn read_env(key: &str) -> Option<String> {
    std::env::var(key).ok()
}

fn parse_port(s: &str) -> Option<u16> {
    s.parse().ok().filter(|&p| p > 0 && p < 65536)
}

fn get_port(key: &str) -> Option<u16> {
    // Equivalent of IOOption.chain
    let raw = read_env(key)?;
    parse_port(&raw)
}

fn main() {
    match get_port("PORT") {
        Some(p) => println!("Listening on {p}"),
        None => println!("PORT missing or invalid"),
    }
}
```

If `read_env` returns `None`, `parse_port` is never called — the chain
short-circuits via `?`. This is the same shape as `Option.flatMap`
lifted into the side-effecting context, and it's the natural Rust way
to write it.

**`envy` / `figment` / `config` crates** exist for more structured env
parsing — reach for those when you have a typed config struct rather
than a single field.

---

## 22. Cell / RefCell / Mutex (controlled mutation)

Rust's controlled-mutation primitives are the equivalents of fp-go's
`IORef` and Effect's `Ref`. They give you interior mutability: a value
that can be mutated through a shared reference, with the mutation
visible in the type system.

**Without FP (plain `mut`):**

```rust
fn main() {
    let mut counter = 0;
    counter += 1;
    counter += 1;
    println!("{counter}"); // 2
}
```

**With FP-style interior mutability:**

```rust
use std::cell::Cell;

fn main() {
    let counter = Cell::new(0);
    counter.set(counter.get() + 1);
    counter.set(counter.get() + 1);
    println!("{}", counter.get()); // 2
}
```

Rust offers three flavors:

- **`Cell<T>`** — `Copy` types only, get/set by value. No borrow
  checking, no runtime cost. Best for small `Copy` data like
  counters and flags.
- **`RefCell<T>`** — non-`Copy` types, borrow at runtime via
  `.borrow()` / `.borrow_mut()`. Panics on double-borrow. Single-
  threaded only.
- **`Mutex<T>` / `RwLock<T>`** — multi-threaded. `Mutex` from
  `std::sync` is blocking; `tokio::sync::Mutex` is async-aware.

`AtomicUsize`, `AtomicBool`, `AtomicPtr<T>`, etc. are the lock-free
equivalents for primitive types — usually the best choice for counters
shared across threads.

```rust
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::Arc;

let counter = Arc::new(AtomicU64::new(0));
let c2 = counter.clone();
std::thread::spawn(move || {
    c2.fetch_add(1, Ordering::Relaxed);
});
counter.fetch_add(1, Ordering::Relaxed);
println!("{}", counter.load(Ordering::Relaxed)); // 2 (eventually)
```

`IORef` is useful when controlled mutation is necessary inside an
otherwise-functional pipeline. FP does **not** mean mutation can
never exist; in Rust it means mutation is **visible in the type**.

---

## 23. Predicate

A predicate is a function `Fn(&A) -> bool`. Rust doesn't ship a
dedicated `Predicate` type, but the pattern is core FP — and the
`std` combinators (`&&`, `||`, `!`) work directly on `bool`.

**Without FP:**

```rust
fn is_adult(age: u32) -> bool { age >= 18 }
fn has_permission(age: u32) -> bool { age >= 21 }

fn can_access(age: u32) -> bool {
    is_adult(age) && has_permission(age)
}
```

**With FP combinators:**

```rust
fn and<'a, A: 'a>(
    p: impl Fn(&A) -> bool + 'a,
    q: impl Fn(&A) -> bool + 'a,
) -> impl Fn(&A) -> bool + 'a {
    move |a| p(a) && q(a)
}

fn main() {
    let is_adult = |age: &u32| *age >= 18;
    let has_permission = |age: &u32| *age >= 21;
    let can_access = and(is_adult, has_permission);

    println!("{}", can_access(&25)); // true
    println!("{}", can_access(&19)); // false
}
```

Predicates are useful for validation, authorization rules, and
filtering conditions. For closed-set sum-type authorization (e.g.
matching on a `Role` enum), pattern matching is more idiomatic than
predicate composition.

---

## 24. Unit

`Unit` represents a value that carries no meaningful information. In
Rust this is the **`()` (unit) type**, which has exactly one value
(also written `()`). It's a real value — you can store it, return it,
and use it as a generic argument.

**Without FP:**

```rust
fn save_user() {
    println!("User saved");
}
```

The implicit return type is `()`. There's no separate "FP" version —
`()` is the type, and `()` is the value:

**With FP:**

```rust
fn save_user() -> () {
    println!("User saved");
    ()
}

// Or, more usefully, as the success type of a Result when the
// operation has no meaningful success value:
fn save_user() -> Result<(), SaveError> {
    println!("User saved");
    Ok(())
}
```

Use `()` when a computation needs to return a value even though there
is no meaningful result to provide — for example as the success type
of `Result<(), E>` for fire-and-forget side-effecting operations.

**Common pitfall:** Don't confuse `()` with "no return". A function
that doesn't return (e.g. `panic!`s or `loop`s forever) has the
`!` (never) type, not `()`. The never type coerces to any other type,
which is what makes `return`, `break`, `panic!`, and `?` on `None` /
`Err` typecheck in any context.

---

## 25. Do notation — the `?` operator

Do notation lets you write sequential-looking code while composing
operations that may fail or produce no value. **Rust's `?` operator
is genuine do-notation** for `Option` and `Result` — no library
needed, no `function*` generators, no `Effect.gen` wrapper.

**Without FP (ladder of matches):**

```rust
fn get_market() -> Option<Market> { None }

fn buy_all() -> Option<String> {
    let market = get_market()?;
    let banana = market.buy_banana()?;
    let apple = market.buy_apple()?;
    let pear = market.buy_pear()?;
    Some(format!("{banana} {apple} {pear}"))
}
```

Wait — that's already FP Rust. The "without FP" version is the
expanded match ladder:

```rust
fn buy_all() -> Option<String> {
    match get_market() {
        None => None,
        Some(market) => match market.buy_banana() {
            None => None,
            Some(banana) => match market.buy_apple() {
                None => None,
                Some(apple) => match market.buy_pear() {
                    None => None,
                    Some(pear) => Some(format!("{banana} {apple} {pear}")),
                },
            },
        },
    }
}
```

**With FP (`?` operator):**

```rust
fn buy_all() -> Option<String> {
    let market = get_market()?;
    let banana = market.buy_banana()?;
    let apple = market.buy_apple()?;
    let pear = market.buy_pear()?;
    Some(format!("{banana} {apple} {pear}"))
}
```

`?` desugars to the match ladder above. If any step returns `None` /
`Err`, the function short-circuits and returns the failure to the
caller. This works for both `Option` and `Result`, and **also works
inside `async fn`** — which gives you async do-notation for free.

**For typed errors that need conversion**, `?` will call `From::from`
to convert the inner error to the function's return type — so you
can return a domain error enum at the top level and `?`-propagate
specific errors from below:

```rust
#[derive(Debug)]
enum AppError {
    Fetch(FetchError),
    Db(DbError),
}

impl From<FetchError> for AppError {
    fn from(e: FetchError) -> Self { AppError::Fetch(e) }
}

impl From<DbError> for AppError {
    fn from(e: DbError) -> Self { AppError::Db(e) }
}

async fn get_user_with_posts(id: u32) -> Result<(User, Vec<Post>), AppError> {
    let user = fetch_user(id).await?;    // FetchError -> AppError via From
    let posts = fetch_posts(&user).await?; // DbError -> AppError via From
    Ok((user, posts))
}
```

This is Rust's strongest FP ergonomics feature. **Common pitfall:**
`?` only works on functions whose return type is `Option<T>` /
`Result<T, E>` (or types that implement the unstable `Try` trait). You
cannot `?` your way out of a `fn main() -> ()` — wrap `main`'s body
in a real fallible function and `match` on it at the top.

---

## 26. Functor

A Functor is a structure that supports mapping a function over a
value inside a context. In Rust, `Option`, `Result`, and `Iterator`
are functors — they all have a `map` method with the right shape.

**Without FP:**

```rust
let number: Option<i64> = Some(10);

let result: Option<i64> = match number {
    Some(n) => Some(n * 2),
    None => None,
};
```

**With FP:**

```rust
let number: Option<i64> = Some(10);

let result: Option<i64> = number.map(|n| n * 2);
// Some(20)
```

The transformation happens inside the `Option` context. You do not
need to manually extract the value.

Rust deliberately does **not** define a `Functor` trait in std — the
trait would need Higher-Kinded Types, which Rust doesn't have. Instead,
each type ships its own `map` method with the same shape. This is a
pragmatic compromise: you lose generic programming over functors but
gain ergonomics, type inference, and zero-cost monomorphisation.

`Result::map`, `Iterator::map`, `Future::map` (via `futures` crate)
all follow the same pattern.

---

## 27. Monad

A Monad is an abstraction that supports chaining computations where
each next computation returns a value in the same context. In Rust,
the monadic bind operation is called **`and_then`** for `Option` and
`Result`, and **`flat_map`** for `Iterator`. (They are the same
operation with different names.)

**Without FP:**

```rust
fn parse_number(value: &str) -> Option<i64> {
    value.parse().ok()
}

fn double_number(n: i64) -> Option<i64> {
    Some(n * 2)
}

fn main() {
    let parsed = parse_number("10");
    let result = match parsed {
        None => None,
        Some(n) => double_number(n),
    };
    println!("{result:?}");
}
```

**With FP:**

```rust
fn parse_number(value: &str) -> Option<i64> {
    value.parse().ok()
}

fn double_number(n: i64) -> Option<i64> {
    Some(n * 2)
}

fn main() {
    let result = parse_number("10").and_then(double_number);
    // Some(20)
    // Or, equivalently, with ? :
    // fn chain() -> Option<i64> {
    //     let n = parse_number("10")?;
    //     double_number(n)
    // }
}
```

A useful starting point:

- Functor → `map` — transform the inner value.
- Monad → `and_then` / `flat_map` / `?` — chain operations that
  each return the same context.

These are abstractions describing behaviour, not separate types you
must instantiate. Rust implements them per-type rather than via a
shared trait, which is why you won't find `trait Monad` in std.

---

## 28. Immutable Collections

FP encourages avoiding mutation of collections. Rust's `Vec<T>` and
`HashMap<K, V>` are mutable by default — `let mut` lets you push,
pop, and insert. For real immutable collections with structural
sharing, use the **`im`** crate.

**Without FP (mutating in place):**

```rust
let mut numbers = vec![1, 2, 3];
numbers.push(4);          // [1, 2, 3, 4]
numbers.retain(|&n| n != 1); // [2, 3, 4]
```

The original `Vec` is consumed — you can't keep both "before" and
"after" snapshots without cloning.

**With FP (copy-on-write transformations on `Vec`):**

```rust
let numbers = vec![1, 2, 3];

let added = {
    let mut v = numbers.clone();
    v.push(4);
    v
}; // [1, 2, 3, 4]

let removed: Vec<i64> = added.iter().copied().filter(|&n| n != 1).collect();
// [2, 3, 4]

// `numbers` is still [1, 2, 3]
```

The clone is O(n) — fine for small collections, costly for large ones.

**With FP (persistent immutable collections via `im`):**

```rust
use im::Vector;

let numbers: Vector<i64> = Vector::unit(1).push_back(2).push_back(3);
// Vector(1, 2, 3)

let added = numbers.push_back(4);   // Vector(1, 2, 3, 4)
let removed = added.iter().filter(|&&n| n != 1).cloned().collect::<Vector<_>>();
// Vector(2, 3, 4)

// `numbers` is still Vector(1, 2, 3) — push_back returned a new vector
// sharing structure with the original. O(log n) cost, not O(n).
```

`im` provides:

- `Vector<T>` — persistent indexed sequence (RRB-tree).
- `HashMap<K, V>`, `HashSet<T>` — persistent hash collections.
- `OrdMap<K, V>`, `OrdSet<T>` — persistent B-tree collections.
- `List<T>` — cons-list.

All share structure with their predecessors on update, so "modifying"
a million-element `Vector` costs O(log n) and leaves the original
unchanged. This is the closest Rust gets to fp-ts's `ReadonlyArray`
or Effect's `Chunk` semantics, and it's the right choice when you have
large collections that you need to snapshot across time (undo stacks,
event-sourced state, copy-on-write data pipelines).

**Common pitfall:** `im` collections implement `Clone` cheaply (because
of structural sharing), but **iteration still walks the whole tree**.
Don't switch from `Vec` to `Vector` for performance reasons without
benchmarking — `Vec` is faster for almost everything except
snapshot-heavy workloads.

---

## 29. AtomicUsize — controlled mutation revisited (atomic
read-modify-write)

The final example revisits the atomic-update pattern from concept 22,
focusing on `fetch_*` operations that do read-modify-write as a
single atomic step — the lock-free equivalent of fp-go's
`IORef.modify` and Effect's `Ref.modify`.

**Without FP (race-prone under concurrency):**

```rust
let mut counter = 0;

fn increment_and_get(counter: &mut i64) -> i64 {
    *counter += 1; // Two threads doing this concurrently can lose updates
    *counter
}
```

**With FP (atomic read-modify-write):**

```rust
use std::sync::atomic::{AtomicI64, Ordering};

fn main() {
    let counter = AtomicI64::new(0);

    // fetch_update returns the previous value, atomically.
    // The closure computes the new value from the old one.
    let before = counter.fetch_update(Ordering::SeqCst, Ordering::SeqCst, |n| Some(n + 1));
    let after = counter.load(Ordering::SeqCst);

    println!("before: {before:?}, after: {after}");
    // before: Ok(0), after: 1
}
```

`fetch_update` returns `Result<OldValue, OldValue>` — `Ok` if the
update succeeded, `Err` if the closure returned `None` (used to abort).
`fetch_add`, `fetch_sub`, `fetch_or`, `fetch_and`, `fetch_xor` are
shortcuts for the common arithmetic / bitwise RMW operations, and they
return the **previous** value as a plain `i64` (not a `Result`).

**When to use which:**

- `Cell<T>` — single-threaded, `Copy` types, simple get/set.
- `RefCell<T>` — single-threaded, non-`Copy` types, runtime borrow
  check.
- `Mutex<T>` / `RwLock<T>` — multi-threaded, complex types, blocking.
- `tokio::sync::Mutex<T>` — multi-threaded async, complex types.
- `AtomicBool` / `AtomicI64` / `AtomicUsize` / etc. — multi-threaded
  lock-free, primitive types. Use `fetch_*` for atomic RMW.
- `Arc<T>` — wrap any of the above for shared ownership across threads.

FP and Rust's concurrency model are complementary. The atomic
primitives give you the same guarantees fp-go's `IORef.modify` and
Effect's `Ref.modify` give you — read and write happen as a single
step, no other thread can observe an intermediate state — without
the monad wrapper.

---

# Quick Reference

| Concept                  | Main purpose                                            |
| ------------------------ | ------------------------------------------------------- |
| Pure Functions           | Predictable functions                                   |
| Immutability             | `let` is immutable by default; `let mut` opts in        |
| Higher-Order Functions   | `impl Fn` / `FnMut` / `FnOnce` traits                   |
| First-Class Functions    | Function items coerce to `fn` pointers                  |
| Composition              | Method chaining on `Iterator`; manual `pipe` helper     |
| Map                      | `Iterator::map`, `Option::map`, `Result::map`           |
| Where / Filter           | `Iterator::filter`, `Option::filter`                    |
| Fold / Reduce            | `Iterator::fold`, `Iterator::sum`, `reduce`             |
| Option                   | `Option<T>` with `Some` / `None`                        |
| Result                   | `Result<T, E>` with `Ok` / `Err`                        |
| Map vs and_then          | `and_then` is `flatMap` / `bind`                        |
| IO                       | Closure `FnOnce() -> A`; no native IO monad             |
| Async (Task)             | `Future<Output = T>` + `async`/`await`                  |
| Async + typed error      | `async fn() -> Result<T, E>`                            |
| Sync + typed error       | `Result<T, E>` (no separate `IOEither`)                 |
| Async Option             | `async fn() -> Option<T>`                               |
| Reader                   | `impl Fn(&Env) -> A` closure                            |
| Reader + async + error   | `async fn(&impl Dep) -> Result<T, E>`                   |
| State                    | `Fn(S) -> (S, A)` function                              |
| StateAsync               | `Mutex` / `RwLock` / atomics                            |
| Sync Option chained      | `?` on `Option` for short-circuit chains                |
| Cell / RefCell / Mutex   | Interior mutability (`IORef` equivalent)                |
| Predicate                | `Fn(&A) -> bool` + manual combinators                   |
| Unit                     | The `()` type and its single value                      |
| Do notation              | The `?` operator (works for `Option` and `Result`)      |
| Functor                  | `map` method on `Option`, `Result`, `Iterator`          |
| Monad                    | `and_then` / `flat_map` / `?` (no shared trait)         |
| Immutable Collections    | `Vec` with clone; `im::Vector` for structural sharing   |
| AtomicUsize (RMW)        | Lock-free read-modify-write                             |

---

# Recommended Learning Order

Do not try to memorise every abstraction at once. Learn them in this
order:

1. **Foundation:** pure functions, immutability (`let` vs `let mut`),
   higher-order functions (`Fn` / `FnMut` / `FnOnce`), composition
   via method chaining.
2. **Collection operations:** `Iterator::map`, `filter`, `fold`,
   `flat_map`. Add `itertools` combinators (`chunk_by`,
   `dedup`, `group_by`) once the std set feels limiting.
3. **Core FP types:** `Option`, `Result`, `map`, `and_then`, `?`.
   Understand why Rust has no `Functor` trait and why each type
   ships its own `map` method.
4. **Async:** `async fn`, `.await`, `Future`. Add `tokio::try_join!`
   for concurrent composition. Understand why
   `async fn() -> Result<T, E>` is your `TaskEither`.
5. **State & mutation:** `Cell` for `Copy` single-threaded,
   `RefCell` for non-`Copy` single-threaded, `Mutex` for
   multi-threaded, `AtomicI64` / `AtomicUsize` for lock-free
   counters. Add `Arc<T>` for shared ownership.
6. **Real Rust application:** build an HTTP handler that returns
   `Result<Json<T>, AppError>` with `?`-propagation from the
   service / repository layers. Use `tokio` for the runtime,
   `serde` for serialization, `tracing` for structured logging.

---

# What to Prioritize in Production Rust Projects

You do not need every FP abstraction in every project. Most
production Rust code uses a small subset:

- **`Option`** for values that may be absent — at lookup boundaries
  (HashMap/Vec lookups), for optional struct fields, for parse
  results where the failure mode is "doesn't parse". Use `?` to
  chain.
- **`Result`** for explicit success/failure — the default return
  type for fallible operations. Use `thiserror::Error` to define
  typed domain errors with `#[from]` for `From` conversions so `?`
  just works.
- **`Iterator` combinators** (`map`, `filter`, `fold`, `flat_map`,
  `collect`) for data transformation pipelines. Add `itertools`
  when you need `chunk_by`, `group_by`, `interleave`, or
  `dedup_by`.
- **`?` operator** for short-circuit error and absence propagation.
  This is Rust's do-notation — embrace it instead of writing
  manual `match` ladders.
- **`async fn` + `tokio`** for any IO-bound work (HTTP servers,
  database clients, file IO). Use `tokio::try_join!` for
  concurrent operations and `tokio::spawn` for background tasks.
- **`Arc<T>` + `tokio::sync::Mutex<T>`** for shared state across
  async tasks. Prefer atomics (`AtomicI64`, `AtomicUsize`) for
  counters.
- **Immutability by default** — `let` bindings, `&T` references,
  `impl Fn` closures. Reach for `mut` only when mutation genuinely
  simplifies the code.
- **Pure functions for business logic** — domain modules should
  take `&self` or no `self` at all and return new values. Push IO
  to the handler / repository boundary.

**Adopt incrementally.** If you're coming from TypeScript or Go,
the biggest mental shift is that Rust already gives you FP
primitives in std — you don't need to install `fp-ts` or `fp-go`.
Start by using `Option` / `Result` + `?` everywhere; add `tokio`
when you need async; add `im` only when you have concrete
snapshot-heavy workloads; ignore `fp-core.rs` and `frunk` unless
you specifically need HList / Coproduct / Validated.

**Common pitfalls in production:**

- **`unwrap()` in non-test code.** It panics, taking down the whole
  process. Use `unwrap_or`, `unwrap_or_default`, `ok_or`, or `?`.
  Enable `#![deny(clippy::unwrap_used)]` in CI to enforce this.
- **`Mutex<T>` across `.await`.** `std::sync::Mutex` is not async-
  aware — holding its guard across `.await` will block the
  executor. Use `tokio::sync::Mutex` for async-held locks.
- **Cloning `Arc<T>` inside hot loops.** `Arc::clone` is cheap
  (atomic increment) but not free. For very hot paths, pass
  `&Arc<T>` instead of cloning.
- **Using `Vec<T>` for snapshot-heavy state.** Cloning a
  million-element `Vec` is O(n). Switch to `im::Vector` for O(log
  n) structural sharing.
- **Defining `Functor` / `Monad` traits in your own code.** Rust
  lacks Higher-Kinded Types — generic-over-functor code does not
  compile. Embrace per-type `map` / `and_then` methods instead of
  trying to abstract over them.
- **Confusing `()` with `!`.** `()` is the unit value. `!` is the
  never type, returned by `panic!`, `return`, infinite loops.
  They behave differently in type inference.
- **Forgetting `?` works on both `Option` and `Result`.** Inside
  an `Option<T>`-returning fn, `?` on `Some(x)` unwraps `x`; on
  `None` it returns `None`. Same shape for `Result`. This is the
  single biggest FP ergonomics feature in Rust — use it.

---

# Official Resources

- **[The Rust Book](https://doc.rust-lang.org/book/)** — chapters
  6 (`if let` / match), 13 (closures and iterators), and 16
  (concurrency) are the FP-relevant core.
- **[Rust by Example](https://doc.rust-lang.org/rust-by-example/)** —
  idiomatic examples of `Option`, `Result`, `?`, iterators.
- **[Rust std: `Option`](https://doc.rust-lang.org/std/option/enum.Option.html)** —
  the full combinator surface.
- **[Rust std: `Result`](https://doc.rust-lang.org/std/result/enum.Result.html)** —
  the full combinator surface.
- **[Rust std: `Iterator`](https://doc.rust-lang.org/std/iter/trait.Iterator.html)** —
  every iterator method, with examples.
- **[itertools crate](https://docs.rs/itertools/latest/itertools/)** —
  extra iterator combinators.
- **[im crate](https://docs.rs/im/latest/im/)** — persistent
  immutable collections (`Vector`, `HashMap`, `HashSet`, `OrdMap`).
- **[tokio crate](https://docs.rs/tokio/latest/tokio/)** — async
  runtime, `tokio::sync::Mutex`, `tokio::try_join!`.
- **[thiserror crate](https://docs.rs/thiserror/latest/thiserror/)** —
  ergonomic derive macro for typed error enums with `#[from]`.
- **[anyhow crate](https://docs.rs/anyhow/latest/anyhow/)** — for
  application-layer error handling when you don't need exhaustiveness.
- **[Functional Programming in Rust (serokell)](https://serokell.io/blog/rust-for-haskellers)** —
  a Haskeller's tour of Rust's FP features.
- **[fp-core.rs on GitHub](https://github.com/JasonShin/fp-core.rs)** —
  Fantasy Land types for Rust. Unmaintained since 2019; read for
  reference, do not depend on for production.
- **[frunk on GitHub](https://github.com/lloydmeta/frunk)** —
  HList, Coproduct, Validated, lenses. Actively maintained.

> **Version note:** This guide is written against Rust **1.75+**
> (uses `async fn` in traits, `let-else`, stable closure capture
> features). `itertools` 0.13, `im` 15, and `tokio` 1.x are current
> as of writing. Rust has no `Functor` / `Monad` trait in std and
> is unlikely to get one before Higher-Kinded Types stabilise —
> that feature has no timeline. Treat examples flagged illustrative
> as learning guides; the Rust compiler is the final arbiter.
