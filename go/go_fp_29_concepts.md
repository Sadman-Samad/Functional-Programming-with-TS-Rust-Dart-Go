# Go Functional Programming with fp-go

## 29 Concepts: Traditional Go vs Functional Programming with fp-go

This guide compares traditional imperative Go with functional programming
(FP) using the **[fp-go](https://github.com/IBM/fp-go)** library
(`github.com/IBM/fp-go/v2`).

fp-go is the most actively maintained functional programming library for Go.
It is strongly influenced by `fp-ts` for TypeScript and ships `Option`,
`Either`, `Result`, `IO`, `IOEither`, `IOOption`, `IOResult`, `Reader`,
`ReaderEither`, `ReaderIOEither`, `State`, `IORef`, and a full optics module
(lenses, prisms, traversals) for immutable updates.

> **A note on the library path.** During the planning conversation a path
> `github.com/TeoDev1611/fp-go` was mentioned. That path returns 404 on
> GitHub at the time of writing — no such public repository exists. The
> canonical fp-go lives at **`github.com/IBM/fp-go`** (formerly
> `github.com/repeale/fp-go`) and this guide targets **v2**, which requires
> **Go 1.24+** because it relies on generic type aliases. If you are on an
> older Go version, use the v1 import path (`github.com/IBM/fp-go/option`,
> etc.) without the `/v2` segment.

> **Important:** FP is a programming style, not a requirement to use a
> particular package. Go is multi-paradigm. Some examples below use plain Go
> features (`for` loops, generics, `errors.Is`); others use fp-go. The two
> styles compose well — most production Go codebases mix them.

The examples use common application scenarios: transforming data, handling
missing values, managing API failures, and composing synchronous and
asynchronous operations.

---

## Installation

```bash
go get github.com/IBM/fp-go/v2
```

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    O "github.com/IBM/fp-go/v2/option"
    E "github.com/IBM/fp-go/v2/either"
    R "github.com/IBM/fp-go/v2/result"
    IO "github.com/IBM/fp-go/v2/io"
    IOE "github.com/IBM/fp-go/v2/ioeither"
    IOR "github.com/IBM/fp-go/v2/ioresult"
    IOO "github.com/IBM/fp-go/v2/iooption"
    RD "github.com/IBM/fp-go/v2/reader"
    RIOE "github.com/IBM/fp-go/v2/readerioeither"
    S "github.com/IBM/fp-go/v2/state"
    IORF "github.com/IBM/fp-go/v2/ioref"
    P "github.com/IBM/fp-go/v2/predicate"
    A "github.com/IBM/fp-go/v2/array"
    N "github.com/IBM/fp-go/v2/number"
    Str "github.com/IBM/fp-go/v2/string"
    L "github.com/IBM/fp-go/v2/lazy"
)
```

> **Version note:** This guide targets fp-go **v2**. Some function signatures
> in v1 differ (notably around `Map` / `Chain` currying and `Ap` type-parameter
> ordering). Treat examples flagged "illustrative" as learning guides and
> verify against `pkg.go.dev/github.com/IBM/fp-go/v2/<pkg>` for your installed
> version.

---

## 1. Pure Functions

A pure function:

- Returns the same output for the same input.
- Does not modify external state.
- Does not perform side effects.

**Without FP:**

```go
func calculateDiscount(price int) int {
    log.Println("Calculating discount") // Side effect
    return price - (price * 10 / 100)
}
```

**With FP:**

```go
func calculateDiscount(price int) int {
    return price - (price * 10 / 100)
}

func main() {
    result := calculateDiscount(1000)
    fmt.Println(result) // 900
}
```

The second function is pure: it only calculates and returns a value.

**When to use:** Pure functions are the right default for business logic —
price calculations, validation, transformation. Push side effects (`log`,
network, `os.Getenv`) to the edges: a `main` function, an HTTP handler, an
`init` block.

**Common pitfall:** Go passes pointers by reference. A function that takes
`*User` and mutates a field is **not** pure, even if it returns nothing.

---

## 2. Immutability

Immutability means not changing existing data. Go structs with value
semantics are immutable by default *if you don't take their address*.

**Without FP:**

```go
type User struct {
    Name string
    Age  int
}

func main() {
    u := User{Name: "John", Age: 25}
    u.Age = 26 // Mutates the existing value
}
```

**With FP:**

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func (u User) WithAge(age int) User {
    return User{Name: u.Name, Age: age}
}

func main() {
    u := User{Name: "John", Age: 25}
    updated := u.WithAge(26) // u is unchanged
}
```

`u` remains unchanged; `updated` is a new value. Copying a struct in Go is a
cheap value copy — there is no need for special `copyWith` machinery.

For deep immutable updates of nested fields, fp-go ships **optics** (lenses):

```go
import "github.com/IBM/fp-go/v2/optics"

var ageLens = func() optics.Lens[User, int] {
    l, _ := optics.MakeLens(
        func(u User) int { return u.Age },
        func(u User, age int) User { return User{Name: u.Name, Age: age} },
    )
    return l
}()

func main() {
    u := User{Name: "John", Age: 25}
    updated := ageLens.Set(u, 26) // User{Name: John, Age: 26}
}
```

**Common pitfall:** `var s = []int{1, 2, 3}; s = append(s, 4)` may share the
backing array. For real immutability, copy first: `cp := append([]int(nil),
s...)`.

---

## 3. Higher-Order Functions

A higher-order function accepts another function as an argument or returns a
function.

**Without FP:**

```go
func calculate(a, b int) int { return a + b }

func execute(a, b int) int {
    return calculate(a, b)
}
```

**With FP:**

```go
func execute(a, b int, op func(int, int) int) int {
    return op(a, b)
}

func main() {
    sum := execute(10, 20, func(a, b int) int { return a + b })
    product := execute(10, 20, func(a, b int) int { return a * b })
    fmt.Println(sum, product) // 30 200
}
```

Since Go 1.18, generic higher-order functions are first-class:

```go
func Map[A, B any](xs []A, f func(A) B) []B {
    out := make([]B, len(xs))
    for i, x := range xs {
        out[i] = f(x)
    }
    return out
}
```

**Common pitfall:** Closures capturing loop variables in pre-1.22 Go shared
the variable across iterations. On Go 1.22+ each iteration has its own
variable — the footgun is gone, but be aware of it when reading older code.

---

## 4. First-Class Functions

Go functions are first-class values. You can store them in variables, pass
them to other functions, and return them.

**Without FP / OOP-style:**

```go
type Calculator struct{}

func (Calculator) Add(a, b int) int { return a + b }

func main() {
    c := Calculator{}
    result := c.Add(10, 20)
}
```

**With FP:**

```go
var add func(int, int) int = func(a, b int) int { return a + b }

func main() {
    result := add(10, 20)
}
```

Both styles are valid. The functional version treats the behaviour itself as a
value, which makes it trivial to compose with `F.Pipe1` / `F.Flow2`.

---

## 5. Function Composition

fp-go ships two composition helpers worth memorising:

- `F.Pipe1(x, f1, f2, ..., fN)` — feeds `x` left-to-right through a list of
  unary functions. Variants exist for up to `F.Pipe10`.
- `F.Flow2(f1, f2)`, `F.Flow3(f1, f2, f3)` ... — returns a new function
  equivalent to `(x) => Pipe1(x, f1, f2, ...)`. Useful when you want to name
  a composed pipeline without yet supplying the input.

**Without FP:**

```go
func processName(name string) string {
    trimmed := strings.TrimSpace(name)
    upper := strings.ToUpper(trimmed)
    return "Hello " + upper
}
```

**With FP-style composition:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    Str "github.com/IBM/fp-go/v2/string"
)

var (
    trimName  = Str.TrimSpace
    uppercase = Str.ToUpper
    greet     = func(s string) string { return "Hello " + s }
)

var processName = F.Flow3(trimName, uppercase, greet)

func main() {
    fmt.Println(processName(" John ")) // Hello JOHN
}
```

Each function has one responsibility and can be reused independently.

**Common pitfall:** All functions passed to `Pipe1` / `Flow2` must be unary
(take a single argument). To compose a binary function, partially apply it
first: `F.Curry2(add)(10)` returns a function `(int) int`.

---

## 6. Map

`Map` transforms every item in a collection into another value.

**Without FP (imperative loop):**

```go
numbers := []int{1, 2, 3, 4}
results := make([]int, 0, len(numbers))
for _, n := range numbers {
    results = append(results, n*2)
}
```

**With fp-go:**

```go
import A "github.com/IBM/fp-go/v2/array"

numbers := []int{1, 2, 3, 4}

results := A.Map(func(n int) int { return n * 2 })(numbers)
// [2, 4, 6, 8]
```

`A.Map` is curried: `A.Map(f)` returns a function `func([]A) []B`. This shape
is what makes `F.Pipe1` composition work.

**When to use:** Almost always prefer `A.Map` over a hand-rolled loop when the
loop body is a pure transformation. It removes the off-by-one and indexing
hazards. Reserve imperative `for` loops for stateful accumulation that doesn't
fit `A.Fold`.

---

## 7. Where / Filter

`Filter` keeps only the elements that satisfy a predicate.

**Without FP (imperative loop):**

```go
numbers := []int{1, 2, 3, 4, 5, 6}
results := []int{}
for _, n := range numbers {
    if n%2 == 0 {
        results = append(results, n)
    }
}
```

**With fp-go:**

```go
import A "github.com/IBM/fp-go/v2/array"

numbers := []int{1, 2, 3, 4, 5, 6}

results := A.Filter(func(n int) bool { return n%2 == 0 })(numbers)
// [2, 4, 6]
```

`A.FilterMap` combines both in a single pass — useful when you're already
transforming each element and some transforms return `None`.

---

## 8. Fold / Reduce

`Fold` processes a collection and combines its elements into one result.
fp-go's `A.Fold` is the canonical reducer.

**Without FP:**

```go
prices := []int{100, 200, 300}
total := 0
for _, p := range prices {
    total += p
}
```

**With fp-go:**

```go
import A "github.com/IBM/fp-go/v2/array"

prices := []int{100, 200, 300}

total := A.Fold(
    func(sum int, p int) int { return sum + p },
    0,
)(prices)
// 600
```

`A.Fold[Item, Acc]` takes `(func(Acc, Item) Acc, Acc) func([]Item) Acc`. The
first argument is the combining function; the second is the seed.

**Common pitfall:** `A.Fold` differs from `A.Reduce` in some fp-ts-derived
libraries — `Reduce` typically does not take a seed and uses the first
element. fp-go's `A.Fold` always takes an explicit seed, which makes the
empty-collection case well-defined.

---

## 9. Option

`Option[T]` represents a value that may or may not exist. It's a type-safe
alternative to nil pointers and zero-value sentinels.

It has two cases:

- `Some(value)`: a value exists.
- `None[T]()`: no value exists.

**Without FP:**

```go
type User struct {
    Name string
    Age  int
}

func findUser(id int) *User {
    if id == 1 {
        return &User{Name: "John", Age: 25}
    }
    return nil
}

func main() {
    u := findUser(10)
    if u != nil {
        fmt.Println(u.Name)
    } else {
        fmt.Println("User not found")
    }
}
```

**With fp-go:**

```go
import O "github.com/IBM/fp-go/v2/option"

func findUser(id int) O.Option[User] {
    if id == 1 {
        return O.Some(User{Name: "John", Age: 25})
    }
    return O.None[User]()
}

func main() {
    result := findUser(10)

    fmt.Println(O.Fold(
        func() string { return "User not found" },
        func(u User) string { return u.Name },
    )(result))
}
```

`Option[User]` makes the possibility of a missing user explicit in the return
type. Callers cannot accidentally dereference a nil pointer.

Other useful `Option` functions:

- `O.FromNillable[T](ptr)` — converts `*T` to `Option[T]`.
- `O.ToNillable[T](opt)` — converts back to `*T` (interoperability with
  legacy APIs).
- `O.IsSome(opt)`, `O.IsNone(opt)` — guards.
- `O.Unwrap(opt)` — returns `(T, bool)`, the idiomatic Go tuple form.
- `O.GetOrElse(func() T)` — provides a default for the `None` case.

**Common pitfall:** `O.Some(ptr)` where `ptr` is nil still constructs a
`Some(nil)` — a real bug. Use `O.FromNillable` at pointer boundaries instead.

---

## 10. Either / Result

`Either[L, R]` represents one of two possible outcomes:

- `Left[L]`: conventionally used for failure.
- `Right[R]`: conventionally used for success.

For the common case where the left side is a Go `error`, fp-go ships
`Result[A] = Either[error, A]` as a more ergonomic alias. The README
recommends `Result` over `Either` for error handling.

**Without FP:**

```go
func fetchUser() (User, error) {
    u, err := api.GetUser()
    if err != nil {
        return User{}, fmt.Errorf("failed to fetch user: %w", err)
    }
    return u, nil
}
```

The caller must remember to check `err`.

**With fp-go:**

```go
import R "github.com/IBM/fp-go/v2/result"

func fetchUser() R.Result[User] {
    u, err := api.GetUser()
    if err != nil {
        return R.Left[User](fmt.Errorf("failed to fetch user: %w", err))
    }
    return R.Right[User](u)
}

// Or, more concisely, wrap the existing function:
var fetchUser = R.Eitherize1(api.GetUser)
```

Usage:

```go
result := fetchUser()

fmt.Println(R.Fold(
    func(err error) string { return fmt.Sprintf("Error: %v", err) },
    func(u User) string { return fmt.Sprintf("User: %s", u.Name) },
)(result))
```

The caller handles both outcomes explicitly. `Eitherize1` is a useful
migration tool: it lifts an existing `(A) (B, error)` function into a function
returning `Result[B]` without rewriting the body.

**When to use Result vs Either:** Use `Result` when the error is a plain
`error`. Use `Either` (with a custom sum-type left) when the failure modes are
a closed set you want to pattern-match exhaustively (`Either[OrderError, *Order]`
where `OrderError` is an interface with `InvalidItem`, `OutOfStock`, etc.).

---

## 11. Map vs Chain (FlatMap)

These operations are central to working with values inside containers such as
`Option` and `Result`.

- `Map` transforms the contained value using a function that returns a plain
  value.
- `Chain` (the fp-ts name for `flatMap` / `bind`) transforms the contained
  value using a function that returns another container of the same kind,
  avoiding nested containers.

**Without FP:**

```go
func parseAge(value string) *int {
    n, err := strconv.Atoi(value)
    if err != nil {
        return nil
    }
    return &n
}

func validateAge(age int) *int {
    if age < 18 {
        return nil
    }
    return &age
}

parsed := parseAge("25")
var result *int
if parsed != nil {
    result = validateAge(*parsed)
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    O "github.com/IBM/fp-go/v2/option"
)

var parseAge = func(value string) O.Option[int] {
    n, err := strconv.Atoi(value)
    if err != nil {
        return O.None[int]()
    }
    return O.Some(n)
}

var validateAge = func(age int) O.Option[int] {
    if age < 18 {
        return O.None[int]()
    }
    return O.Some(age)
}

var result = F.Pipe1(
    parseAge("25"),
    O.Chain(validateAge),
)
// Some(25)
```

Using `Map` for a plain transformation:

```go
var doubled = F.Pipe1(
    parseAge("25"),
    O.Map(func(n int) int { return n * 2 }),
)
// Some(50)
```

Both `Map` and `Chain` preserve the `Option` context. If you `Map` a function
that *returns* an `Option`, you get `Option[Option[T]]` — almost always a bug.
Reach for `Chain` instead.

**Naming note:** fp-go uses `Chain` for `flatMap` and `Map` for `map`. The
`MonadChain` helper is the right-to-left composition variant
(`MonadChain(f, g)` applies `f` first, then `g`).

---

## 12. IO

`IO[A]` represents a synchronous computation that may perform side effects
and eventually produce an `A`. In fp-go, `IO[A]` is a type alias for
`func() A`.

**Without FP:**

```go
func readUsername() string {
    fmt.Println("Reading username")
    return "John"
}

func main() {
    username := readUsername() // Runs immediately
}
```

**With fp-go:**

```go
import IO "github.com/IBM/fp-go/v2/io"

var readUsername IO.IO[string] = func() string {
    fmt.Println("Reading username")
    return "John"
}

// Or, equivalently:
var readUsername = IO.Of("John") // Pure IO

func main() {
    // Describes the computation; nothing runs yet
    username := readUsername() // Executes it
    fmt.Println(username)
}
```

`IO.Of(x)` lifts a pure value into an `IO`. `IO.Map(f)` transforms the
result. `IO.Chain(f)` sequences two `IO`s.

```go
decorated := F.Pipe2(
    IO.Of("Hello, fp-go!"),
    IO.Map(strings.ToUpper),
    IO.Map(func(s string) string { return ">>> " + s + " <<<" }),
)

fmt.Println(decorated()) // >>> HELLO, FP-GO! <<<
```

`IO` helps separate *describing* a side-effecting operation from *running*
it. It does not make the side effect disappear.

**Common pitfall:** Unlike Effect's `Effect.sync`, fp-go's `IO` cannot carry
a typed error. Use `IOEither` or `IOResult` when the operation can fail.

---

## 13. Asynchronous computation (Task)

> **Library gap:** fp-go **does not** ship a dedicated `Task` /
> `TaskEither` monad the way fp-ts does. Go handles concurrency natively via
> goroutines and channels, and the maintainers chose not to wrap that.
> Instead, fp-go offers `IO` (synchronous) and `IOResult` (synchronous with
> typed error). For async work you compose `IOResult` with goroutines
> yourself, or use the `idiomatic/readerioresult` package which is designed
> for `func(ctx) (A, error)` workflows including HTTP handlers.

This is the biggest mental shift coming from fp-ts or Effect. The patterns
below show how to fill the gap idiomatically.

**Without FP:**

```go
func fetchUser(ctx context.Context) (*User, error) {
    return api.GetUser(ctx)
}

func main() {
    u, err := fetchUser(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(u.Name)
}
```

**With fp-go (async via goroutine + IOResult):**

```go
import (
    "context"

    IOR "github.com/IBM/fp-go/v2/ioresult"
)

// A "Task" is an IOResult that captures context at call time.
type Task[A any] func(context.Context) IOR.IOResult[A]

func fromPromise[A any](run func(context.Context) (A, error)) Task[A] {
    return func(ctx context.Context) IOR.IOResult[A] {
        // IOResult is sync, so we need to wait for the async work here.
        // For non-blocking composition, return a channel-based wrapper instead.
        type result struct {
            v   A
            err error
        }
        ch := make(chan result, 1)
        go func() {
            v, err := run(ctx)
            ch <- result{v, err}
        }()
        r := <-ch
        if err != nil {
            return IOR.Left[A](err)
        }
        return IOR.Right[A](r.v)
    }
}

var fetchUser = fromPromise(func(ctx context.Context) (User, error) {
    return api.GetUser(ctx)
})
```

The takeaway: in Go, **a function returning `IOResult[A]` is your `TaskEither`**,
and a function returning `IO[A]` is your `Task`. The async scheduling is done
with `go` statements and channels — that's the language's idiom.

**When to reach for plain Go instead:** If your function does one HTTP call
and returns, `(*User, error)` is fine. Save `IOResult` for code that benefits
from `Map` / `Chain` composition (pipelines of transformations, validation
chains).

---

## 14. IOResult (async + typed error, "TaskEither")

`IOResult[A]` = `IO[Result[A]]` = `func() Result[A]`. It is the type to use
for synchronous or async-completed computations that can fail with a typed
error.

**Without FP:**

```go
func fetchUser() (*User, error) {
    u, err := api.GetUser()
    if err != nil {
        return nil, fmt.Errorf("unable to fetch user: %w", err)
    }
    return u, nil
}
```

**With fp-go:**

```go
import IOR "github.com/IBM/fp-go/v2/ioresult"

func fetchUser() IOR.IOResult[User] {
    u, err := api.GetUser()
    if err != nil {
        return IOR.Left[User](fmt.Errorf("unable to fetch user: %w", err))
    }
    return IOR.Right[User](u)
}
```

Usage:

```go
result := fetchUser()()

fmt.Println(IOR.Fold(
    func(err error) string { return fmt.Sprintf("Failed: %v", err) },
    func(u User) string { return fmt.Sprintf("Success: %s", u.Name) },
)(result))
```

Or use `TryCatch` to lift an existing `(A, error)` function:

```go
var fetchUser = IOR.TryCatch(
    func() (User, error) { return api.GetUser() },
    func(err error) error { return fmt.Errorf("unable to fetch user: %w", err) },
)
```

This is the pattern to reach for in API clients, repositories, and
authentication flows.

**`idiomatic/ioresult` for performance:** If this is on a hot path, the
`idiomatic/ioresult` package replaces the struct wrapper with Go's native
`(A, error)` tuple shape and gives 2-10x speedup with zero allocations.

---

## 15. IOEither

`IOEither[L, R]` is `IO[Either[L, R]]` — a synchronous computation that can
fail with a typed error of any type (not just `error`). Use it when the left
side is a domain-specific sum type.

**Without FP:**

```go
func readConfiguration() (int, error) {
    n, err := strconv.Atoi("123")
    if err != nil {
        return 0, errors.New("invalid configuration")
    }
    return n, nil
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    IOE "github.com/IBM/fp-go/v2/ioeither"
)

type ConfigError struct{ Reason string }

var readConfiguration IOE.IOEither[ConfigError, int] = IOE.TryCatch(
    func() (int, error) { return strconv.Atoi("123") },
    func(err error) ConfigError { return ConfigError{Reason: err.Error()} },
)

func main() {
    result := readConfiguration()

    fmt.Println(IOE.Fold(
        func(e ConfigError) string { return fmt.Sprintf("Error: %s", e.Reason) },
        func(v int) string { return fmt.Sprintf("Value: %d", v) },
    )(result))
}
```

Use `IOResult` when the error is a plain `error`; use `IOEither` when the
left side is a richer domain type you want to pattern-match.

---

## 16. IOOption

`IOOption[A]` is `IO[Option[A]]` — a synchronous effect whose result may be
absent.

**Without FP:**

```go
func readSetting(key string) *string {
    v, ok := os.LookupEnv(key)
    if !ok {
        return nil
    }
    return &v
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    IOO "github.com/IBM/fp-go/v2/iooption"
    O "github.com/IBM/fp-go/v2/option"
)

var readSetting = func(key string) IOO.IOOption[string] {
    return func() O.Option[string] {
        v, ok := os.LookupEnv(key)
        if !ok {
            return O.None[string]()
        }
        return O.Some(v)
    }
}
```

Usage:

```go
result := readSetting("HOME")()

fmt.Println(O.Fold(
    func() string { return "Setting missing" },
    func(v string) string { return v },
)(result))
```

Useful when reading a synchronous source where a value may legitimately be
absent (env vars, in-memory cache, query params).

---

## 17. Reader

`Reader[E, A]` represents a computation that needs an environment or
dependency `E` to produce a value `A`. In fp-go it's a type alias for
`func(E) A`.

**Without FP:**

```go
type UserRepository struct {
    api ApiClient
}

func (r UserRepository) GetUser() User {
    return r.api.GetUser()
}

func main() {
    repo := UserRepository{api: apiClient}
    u := repo.GetUser()
}
```

**With fp-go:**

```go
import RD "github.com/IBM/fp-go/v2/reader"

type ApiClient interface {
    GetUser() User
}

var getUser RD.Reader[ApiClient, User] = func(api ApiClient) User {
    return api.GetUser()
}

func main() {
    u := getUser(apiClient)
    fmt.Println(u.Name)
}
```

`Reader` is useful for passing dependencies explicitly and making
computations easy to test — call `getUser(fakeApi)` in tests.

For dependency *plus* IO *plus* typed errors, jump straight to
`ReaderIOResult` (concept 18).

---

## 18. ReaderIOResult

`ReaderIOResult[R, A]` combines:

- `Reader`: requires a dependency/environment `R`.
- `IO`: performs work (potentially side-effecting).
- `Result`: represents failure (`error`) or success.

This is fp-go's closest equivalent to fp-ts's `ReaderTaskEither`.

**Without FP:**

```go
func getUser(ctx context.Context, api ApiClient) (*User, error) {
    u, err := api.GetUser(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed: %w", err)
    }
    return u, nil
}
```

**With fp-go (idiomatic variant):**

```go
import (
    "context"

    RIR "github.com/IBM/fp-go/v2/idiomatic/readerioresult"
)

type ApiClient interface {
    GetUser(ctx context.Context) (User, error)
}

var getUser RIR.ReaderIOResult[ApiClient, User] = func(api ApiClient) func(context.Context) (User, error) {
    return func(ctx context.Context) (User, error) {
        u, err := api.GetUser(ctx)
        if err != nil {
            return User{}, fmt.Errorf("failed: %w", err)
        }
        return u, nil
    }
}
```

Usage:

```go
result, err := getUser(apiClient)(context.Background())
if err != nil {
    fmt.Println("Failed:", err)
} else {
    fmt.Println("User:", result.Name)
}
```

The idiomatic package uses Go's native `(A, error)` tuple instead of a
struct wrapper — it's faster and feels natural to Go developers, at the cost
of not having all the fp-ts combinators available.

This is the pattern for repository and service layers where dependencies,
asynchronous work, and expected failures all need to be composed.

---

## 19. State

`State[S, A]` represents a computation that receives a state `S` and produces
a result `A` alongside a new state `S`. In fp-go it's
`func(S) tuple.Tuple2[S, A]`.

**Without FP:**

```go
type Wallet struct{ balance int }

func (w *Wallet) addMoney(amount int) {
    w.balance += amount
}

func main() {
    w := &Wallet{balance: 100}
    w.addMoney(50)
    fmt.Println(w.balance) // 150
}
```

**With fp-go:**

```go
import (
    S "github.com/IBM/fp-go/v2/state"
    T "github.com/IBM/fp-go/v2/tuple"
)

var addMoney S.State[int, int] = func(balance int) T.Tuple2[int, int] {
    newBalance := balance + 50
    return T.Of2(newBalance, newBalance)
}

func main() {
    result := addMoney(100)
    fmt.Println(result.Second) // Result: 150
    fmt.Println(result.First)  // New state: 150
}
```

The original input state is not mutated. The computation returns the next
state and a result. `S.Map`, `S.Chain`, and `S.Evaluate` (run and discard the
state) are the combinators you'll use most.

**When to use:** Use `State` when you have a deterministic state machine
that's pure (parsing, simple games, simulations). For shared mutable state
across goroutines, use `IORef` (concept 22) or `sync/atomic` instead.

---

## 20. StateAsync

Go does not have a built-in async state monad in fp-go. The pattern is to
combine `State` with `IOResult` for async state transitions: build a
`func(S) IOResult[Tuple2[S, A]]`.

**Without FP:**

```go
type AppState struct{ loadingCount int }

func (s *AppState) loadData() error {
    s.loadingCount++
    time.Sleep(1 * time.Second)
    s.loadingCount--
    return nil
}
```

**With FP-style async state:**

```go
import (
    "context"
    "time"

    IOR "github.com/IBM/fp-go/v2/ioresult"
    T "github.com/IBM/fp-go/v2/tuple"
)

type AsyncState[S, A any] func(S) IOR.IOResult[T.Tuple2[S, A]]

func loadData() AsyncState[int, struct{}] {
    return func(count int) IOR.IOResult[T.Tuple2[int, struct{}]] {
        return func() IOR.Result[T.Tuple2[int, struct{}]] {
            next := count + 1
            time.Sleep(1 * time.Second)
            next = next - 1
            return IOR.Right[T.Tuple2[int, struct{}]](T.Of2(next, struct{}{}))
        }
    }
}
```

Learn ordinary `State` first. Async state abstractions are more advanced and
not needed in every Go project — `sync.Mutex` plus a struct field is often
the simpler answer.

---

## 21. IOOption revisited — chaining

Concept 16 introduced `IOOption`. The pattern that makes it powerful is
chaining with `IOO.Chain`:

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    IOO "github.com/IBM/fp-go/v2/iooption"
)

var (
    readEnv   = func(key string) IOO.IOOption[string] { /* ... */ }
    parsePort = func(s string) IOO.IOOption[int] {
        return func() O.Option[int] {
            n, err := strconv.Atoi(s)
            if err != nil || n < 0 || n > 65535 {
                return O.None[int]()
            }
            return O.Some(n)
        }
    }
)

var port = func(key string) IOO.IOOption[int] {
    return F.Pipe1(
        readEnv(key),
        IOO.Chain(parsePort),
    )
}
```

If `readEnv` returns `None`, `parsePort` is never called — the chain
short-circuits. This is the same shape as `Option.flatMap` lifted into the
`IO` context.

---

## 22. IORef

`IORef[A]` provides a mutable reference whose read/write operations are
represented as `IO` effects.

**Without FP:**

```go
counter := 0
counter++
counter++
fmt.Println(counter) // 2
```

**With fp-go:**

```go
import IORF "github.com/IBM/fp-go/v2/ioref"

func main() {
    counter := IORF.Of(0)

    initial := counter.Read()()
    fmt.Println(initial) // 0

    counter.Write(10)()

    result := counter.Read()()
    fmt.Println(result) // 10

    counter.Modify(func(n int) int { return n + 1 })()
    fmt.Println(counter.Read()()) // 11
}
```

`IORef` is useful when controlled mutation is necessary inside an otherwise
functional pipeline — counters in a long-running service, accumulators in a
stateful `IO.Chain`. FP does **not** mean mutation can never exist; it means
mutation should be visible in the types.

**Concurrency:** `IORef` is **not** safe for concurrent use. For atomic
shared state across goroutines, use `sync/atomic` directly or wrap an
`atomic.Int64` in your own `IO`-shaped helpers.

---

## 23. Predicate

`Predicate[A]` represents a condition that returns `true` or `false` for a
value. Predicates compose.

**Without FP:**

```go
func isAdult(age int) bool       { return age >= 18 }
func hasPermission(age int) bool { return age >= 21 }

func canAccess(age int) bool {
    return isAdult(age) && hasPermission(age)
}
```

**With fp-go:**

```go
import P "github.com/IBM/fp-go/v2/predicate"

var (
    isAdult       = P.FromPredicate(func(age int) bool { return age >= 18 })
    hasPermission = P.FromPredicate(func(age int) bool { return age >= 21 })
)

var canAccess = P.And(hasPermission)(isAdult)

func main() {
    fmt.Println(canAccess(25)) // true
    fmt.Println(canAccess(19)) // false
}
```

fp-go's `predicate` package ships `And`, `Or`, `Not`, and `Xor` combinators.
Predicates are useful for validation, authorization rules, and filtering
conditions — anywhere you'd otherwise scatter boolean logic across helpers.

---

## 24. Unit

`Unit` represents a value that carries no meaningful information. In Go this
is `struct{}` (the empty struct, zero bytes).

**Without FP:**

```go
func saveUser() {
    fmt.Println("User saved")
}
```

**With FP:**

```go
type Unit = struct{}

var unit Unit = Unit{}

func saveUser() Unit {
    fmt.Println("User saved")
    return unit
}
```

Use `Unit` when a computation needs to return a value even though there is no
meaningful result to provide — for example as the right operand of a generic
`Chain` that needs a uniform return type.

In practice, fp-go's `IO[Unit]` is rarely written explicitly: prefer
`IO[struct{}]` (or just `func()` with no return) when you mean "this effect
exists for its side effects only".

---

## 25. Do Notation

Go does not have generator syntax like JavaScript's `function*` or
Scala's `for`-comprehensions. The fp-go idiom for sequential chaining is
`F.Pipe1` with `Chain`, **not** a do-notation block. The pattern below shows
the equivalent of fp-ts's `Option.Do` for a chained "buy fruit" workflow.

**Without FP (ladder of nil checks):**

```go
func getMarket() *Market {
    return nil // or some market
}

func main() {
    market := getMarket()
    if market == nil {
        return
    }

    banana := market.buyBanana()
    if banana == nil {
        return
    }

    apple := market.buyApple()
    if apple == nil {
        return
    }

    pear := market.buyPear()
    if pear == nil {
        return
    }

    fmt.Println(banana, apple, pear)
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    O "github.com/IBM/fp-go/v2/option"
)

func getMarket() O.Option[Market]      { /* ... */ }
func buyBanana(Market) O.Option[string] { /* ... */ }
func buyApple(Market) O.Option[string]  { /* ... */ }
func buyPear(Market) O.Option[string]   { /* ... */ }

var buyAll O.Option[string] = F.Pipe3(
    getMarket(),
    O.Chain(func(m Market) O.Option[struct {
        market Market
        banana string
    }] {
        return O.Map(func(b string) struct {
            market Market
            banana string
        } {
            return struct {
                market Market
                banana string
            }{m, b}
        })(buyBanana(m))
    }),
    // ... and so on
)
```

That explicit shape gets awkward quickly. The idiomatic Go workaround is
`Option.Chain` with intermediate structs, **or** the **`Eitherize` / `FromOption`
helpers in the `result` package**, which let you short-circuit cleanly:

```go
import (
    R "github.com/IBM/fp-go/v2/result"
    E "github.com/IBM/fp-go/v2/errors"
)

var buyAll = func() R.Result[string] {
    market, ok := O.Unwrap(getMarket())
    if !ok {
        return R.Left[string](E.OnSome[Market]("market missing"))
    }
    banana, ok := O.Unwrap(buyBanana(market))
    if !ok {
        return R.Left[string](E.OnSome[string]("banana missing"))
    }
    apple, ok := O.Unwrap(buyApple(market))
    if !ok {
        return R.Left[string](E.OnSome[string]("apple missing"))
    }
    pear, ok := O.Unwrap(buyPear(market))
    if !ok {
        return R.Left[string](E.OnSome[string]("pear missing"))
    }
    return R.Right[string](fmt.Sprintf("%s %s %s", banana, apple, pear))
}
```

The pattern: short-circuit with early `return`, accumulate typed errors via
`Result.Left`, finish with `Result.Right`. If any step fails, the function
returns the typed error without running further steps. This is the closest
Go gets to do-notation — embrace it.

**Common pitfall:** Don't try to wrap a sequential chain in `F.Pipe1` when
each step depends on the previous step's result **and** you need to thread
multiple variables through. `Pipe1` shines for *transformations*; for
*chained dependent computations* the early-return pattern reads better.

---

## 26. Functor

A Functor is a structure that supports mapping a function over a value
inside a context. `Option` is a Functor.

**Without FP:**

```go
var number *int = ptrTo(10)

var result *int
if number != nil {
    n := *number * 2
    result = &n
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    O "github.com/IBM/fp-go/v2/option"
)

var number O.Option[int] = O.Some(10)

var result = F.Pipe1(
    number,
    O.Map(func(n int) int { return n * 2 }),
)
// Some(20)
```

The transformation happens inside the `Option` context. You do not need to
manually extract the value. `Either`, `Result`, `IO`, `IOEither`,
`Reader`, and `State` are all functors in fp-go.

---

## 27. Monad

A Monad is an abstraction that supports chaining computations where each next
computation returns a value in the same context. `Option` supports `Chain`.

**Without FP:**

```go
func parseNumber(value string) *int {
    n, err := strconv.Atoi(value)
    if err != nil {
        return nil
    }
    return &n
}

func doubleNumber(n int) *int {
    doubled := n * 2
    return &doubled
}

parsed := parseNumber("10")
var result *int
if parsed != nil {
    result = doubleNumber(*parsed)
}
```

**With fp-go:**

```go
import (
    F "github.com/IBM/fp-go/v2/function"
    O "github.com/IBM/fp-go/v2/option"
)

var parseNumber = func(value string) O.Option[int] {
    n, err := strconv.Atoi(value)
    if err != nil {
        return O.None[int]()
    }
    return O.Some(n)
}

var doubleNumber = func(n int) O.Option[int] { return O.Some(n * 2) }

var result = F.Pipe1(
    parseNumber("10"),
    O.Chain(doubleNumber),
)
// Some(20)
```

A useful starting point:

- Functor → `Map`
- Monad → `Chain` (also commonly called `flatMap` / `bind`)

These are abstractions describing behaviour, not separate types you must
instantiate.

---

## 28. Immutable Collections

FP encourages avoiding mutation of collections. Create new collections when
transforming data.

**Without FP:**

```go
numbers := []int{1, 2, 3}
numbers = append(numbers, 4)
for i, n := range numbers {
    if n == 1 {
        numbers = append(numbers[:i], numbers[i+1:]...)
        break
    }
}
```

The original slice's contents change (or share a backing array with the
caller — subtle bug territory).

**With FP-style transformations:**

```go
import A "github.com/IBM/fp-go/v2/array"

numbers := []int{1, 2, 3}

added := append(append([]int(nil), numbers...), 4)
// [1, 2, 3, 4]

removed := A.Filter(func(n int) bool { return n != 1 })(added)
// [2, 3, 4]
```

The original slice remains unchanged. The `append([]int(nil), numbers...)`
idiom guarantees a fresh backing array — use it whenever you want to be sure
a downstream consumer cannot mutate your data.

fp-go ships functional array helpers that always return new slices:

- `A.Map(f)` — pure transform.
- `A.Filter(p)` — keep matching.
- `A.Fold(f, seed)` — reduce.
- `A.Chunk(n)` — split into chunks of size `n`.
- `A.Reverse()` — non-mutating reverse.
- `A.Uniq()` — dedupe (requires `comparable` element type).

For *truly* immutable collections (persistent vectors, hash maps with
structural sharing), consider the separate `immutable` package
(`github.com/IBM/fp-go/v2/...` and external `github.com/mocafs/go-collections`).
A `[]int` declared with `var` is not immutable by itself — Go has no
`readonly` modifier, so discipline + copy-on-write is the practical answer.

---

## 29. IORef — atomic modify

This final example revisits `IORef` to emphasise the distinction between
describing an effect and executing it, and to show the atomic `Modify` /
`Update` API.

**Without FP (race-prone under concurrency):**

```go
counter := 0
counter = 10
counter++
fmt.Println(counter) // 11
```

**With fp-go:**

```go
import IORF "github.com/IBM/fp-go/v2/ioref"

func main() {
    // Create the ref (described, not yet run)
    counter := IORF.Of(0)

    // Write a new value (described as an IO)
    writeTen := counter.Write(10)
    writeTen() // execute

    // Atomically read-modify-write
    increment := counter.Modify(func(n int) int { return n + 1 })
    increment() // execute, counter is now 11

    // Read
    read := counter.Read()
    fmt.Println(read()) // 11
}
```

The central idea: reading, writing, and modifying are all represented as
`IO` operations. You compose them, pass them around, and execute them when
ready.

> **Concurrency caveat:** `IORef.Modify` is **not** atomic across goroutines.
> It is atomic *in the sense that the read and the write are part of a single
> described `IO`*, but fp-go's `IORef` uses a plain field underneath. For
> cross-goroutine shared mutable state, use `sync/atomic.Int64`,
> `atomic.Pointer[T]`, or wrap them in your own `IO`-shaped helpers. FP and
> Go's concurrency model are complementary — don't fight the language.

---

# Quick Reference

| Concept                  | Main purpose                                            |
| ------------------------ | ------------------------------------------------------- |
| Pure Functions           | Predictable functions                                   |
| Immutability             | Avoid modifying existing data (struct value semantics)  |
| Higher-Order Functions   | Pass functions as arguments or return them              |
| First-Class Functions    | Treat functions as values                               |
| Composition              | Combine smaller functions (`F.Pipe1`, `F.Flow3`)        |
| Map                      | Transform collection values (`A.Map`)                   |
| Where / Filter           | Filter collection values (`A.Filter`)                   |
| Fold / Reduce            | Reduce a collection to one result (`A.Fold`)            |
| Option                   | Handle missing values (`O.Some`, `O.None`)              |
| Either / Result          | Represent success or failure                            |
| Map vs Chain             | Chain contextual computations (`O.Chain`)               |
| IO                       | Represent synchronous side effects (`IO.Of`)            |
| Async (Task)             | Goroutines + `IOResult`; no native `Task` monad         |
| IOResult                 | Sync/async with typed failures (`IOR.TryCatch`)         |
| IOEither                 | Sync with custom typed failures                         |
| IOOption                 | Sync with optional results                              |
| Reader                   | Supply dependencies/environment (`RD.Reader`)           |
| ReaderIOResult           | Dependencies + IO + typed failures                      |
| State                    | Model state transitions (`S.State`)                     |
| StateAsync               | Model asynchronous state transitions (manual)           |
| IOOption (chained)       | Short-circuit chained optional sync effects             |
| IORef                    | Represent controlled mutable references                 |
| Predicate                | Represent and compose conditions (`P.And`, `P.Or`)      |
| Unit                     | Represent an empty result (`struct{}`)                  |
| Do notation              | Use early returns + `Result`; no generators in Go       |
| Functor                  | Map over a value inside a context                       |
| Monad                    | Chain computations inside a context (`Chain`)           |
| Immutable Collections    | Copy-on-write slices + `A.Map` / `A.Filter`             |
| IORef (atomic modify)    | Read-modify-write as a single described effect          |

---

# Recommended Learning Order

Do not try to memorise every abstraction at once. Learn them in this order:

1. **Foundation:** pure functions, immutability (struct value semantics),
   higher-order functions, composition (`F.Pipe1`, `F.Flow2`).
2. **Collection operations:** `A.Map`, `A.Filter`, `A.Fold`, and predicates
   (`P.And`, `P.Or`).
3. **Core FP types:** `Option`, `Either`, `Result`, `Map`, `Chain`, Functor,
   Monad.
4. **Side effects:** `IO.Of`, `IO.Map`, `IO.Chain`, then `IOEither` /
   `IOResult` for typed errors. Understand that fp-go has **no native
   `Task`** — async is goroutines + `IOResult`.
5. **Dependencies:** `Reader` for sync DI, then `ReaderIOResult` (prefer the
   `idiomatic/readerioresult` package for performance).
6. **State:** `State` for pure state machines, `IORef` for controlled
   mutation inside an `IO` pipeline. For cross-goroutine state, fall back to
   `sync/atomic`.
7. **Real Go application:** wrap HTTP handlers and repositories in
   `ReaderIOResult`, validate input with `Option`/`Result` chains, push
   side effects to the handler boundary.

---

# What to Prioritize in Production Go Projects

You do not need every fp-go abstraction in every project. Prioritize:

- **`Option`** for values that may be absent — especially at pointer
  boundaries, where `*T` would otherwise force nil-checks throughout your
  codebase. `O.FromNillable` / `O.ToNillable` make the migration painless.
- **`Result`** for explicit success/failure results — `R.Eitherize1` lifts
  existing `(A) (B, error)` functions without rewriting them.
- **`IOResult`** for sync (or async-completed) operations that can fail —
  the natural fit for repository methods and HTTP client wrappers.
- **`Map` and `Chain`** for composing transformations — replaces nested
  `if err != nil` blocks with a pipeline.
- **`ReaderIOResult`** (prefer `idiomatic/readerioresult`) for the service
  layer — gives you dependency injection, async work, and typed errors in
  one shape.
- **`Pipe1` / `Flow2`** for readable composition — left-to-right instead of
  inside-out function calls.
- **Immutability** for predictable application state — `readonly` discipline
  on slices via copy-on-write and the `A.*` helpers.
- **Pure functions** for business logic that is easy to test — validation,
  pricing, transformations.

**Adopt incrementally.** Start by wrapping your existing `(A, error)`
functions with `R.Eitherize1` at the repository boundary. Add `Option` for
lookups that may return nothing. Reach for `Pipe1` / `Flow2` when you have
3+ transformation steps. Save `Reader`, `State`, and optics for the modules
that genuinely benefit from them — usually a small minority.

**Common pitfalls in production:**

- **Overusing monads for simple code.** A function that does one HTTP call
  and returns is fine as `(A, error)`. Wrapping it in `IOResult` only helps
  when you compose it with other `IOResult`s.
- **Forgetting `IORef` is not concurrency-safe.** It's a plain field; use
  `sync/atomic` or `sync.Mutex` for cross-goroutine state.
- **Mixing `Either[CustomError, A]` and `Result[A]` (`Either[error, A]`) in
  the same layer.** Pick one error model per layer — usually `Result` at the
  edges (where Go's `error` is the lingua franca) and `Either[DomainError,
  A]` inside the domain.
- **Calling `IO` effects eagerly during construction.** `IO.Of(x)` is pure
  and safe, but `IO.TryCatch(...)` describes a side effect; only run it at
  the handler boundary.
- **Using the `generic` subpackages.** These were a v1 workaround for Go's
  pre-1.18 generics. On v2 (Go 1.24+) the top-level packages are the only
  ones you need.
- **Ignoring the `idiomatic/*` packages on hot paths.** The struct-based
  `Option[T]` allocates; the idiomatic `(T, bool)` tuple variant does not.
  For request handlers serving thousands of QPS, switch to idiomatic types
  at the boundary.

---

# Official Resources

- [fp-go on GitHub](https://github.com/IBM/fp-go) — source, issues, release
  notes.
- [fp-go v2 README](https://github.com/IBM/fp-go/blob/main/v2/README.md) —
  installation, migration guide from v1, idiomatic vs standard packages.
- [fp-go on pkg.go.dev](https://pkg.go.dev/github.com/IBM/fp-go/v2) —
  per-package API reference (look up `option`, `either`, `result`, `io`,
  `ioresult`, `reader`, `state`, `ioref`, `predicate`, `array`, `optics`).
- [fp-ts documentation](https://gcanti.github.io/fp-ts/) — fp-go's API
  closely mirrors fp-ts; the fp-ts learning materials are directly
  applicable.
- [Functional I/O in Go](https://github.com/IBM/fp-go/blob/main/v2/FUNCTIONAL_IO.md)
  — the maintainers' deep dive on `Context`, errors, and the `Reader` pattern
  for I/O in Go.
- [Idiomatic vs Standard comparison](https://github.com/IBM/fp-go/blob/main/v2/IDIOMATIC_COMPARISON.md)
  — when to use the tuple-based packages vs the struct-based ones.
- [samber/mo](https://github.com/samber/mo) — alternative, lighter-weight
  FP library (`Option`, `Result`, `Box`) if fp-go feels too heavy.

> **Version note:** This guide is written against fp-go **v2**, which
> requires Go 1.24+ for generic type aliases. On older Go versions use v1
> imports (`github.com/IBM/fp-go/option` without the `/v2`); some function
> signatures differ (notably `Ap` type-parameter ordering and the absence of
> the `idiomatic/*` packages). Always check
> `pkg.go.dev/github.com/IBM/fp-go/v2/<pkg>` for your installed version
> before relying on a snippet. Treat examples marked illustrative or
> conceptual as learning guides, not guaranteed copy-paste code.
