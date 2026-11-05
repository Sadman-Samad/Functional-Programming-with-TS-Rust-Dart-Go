# TypeScript Functional Programming with Effect

## 29 Concepts: Traditional TypeScript vs Functional Programming with Effect

This guide compares traditional imperative/object-oriented TypeScript with
functional programming (FP) using the **[Effect](https://effect.website)**
library (`effect` package, v3 / v4).

Effect is the spiritual successor to `fp-ts`. It provides the same FP
abstractions (`Option`, `Either`, monadic `map` / `flatMap`) plus a unified
`Effect<R, E, A>` type that replaces `IO`, `Task`, `IOEither`, `TaskEither`,
`ReaderTaskEither` and friends from `fp-ts` with a single composable value.

The examples use common application scenarios: transforming data, handling
missing values, managing API failures, and composing asynchronous operations.

> **Important:** FP is a programming style, not a requirement to use a
> particular package. TypeScript is multi-paradigm, so you can mix OOP,
> imperative programming, and FP in the same application. Some examples
> below use native TypeScript features (`Array.map`, `Array.filter`,
> `Array.reduce`); others use the `effect` package.

---

## Installation

```bash
npm install effect
# or
pnpm add effect
# or
yarn add effect
```

```typescript
import { Effect, Option, Either, pipe, flow, Ref, Layer, Context } from "effect"
```

> **Version note:** This guide targets Effect v3 / v4 (`effect@3.x` or newer).
> APIs that changed between versions are flagged inline. Always check the
> installed version's [reference](https://effect.website/docs) before
> copy-pasting into production.

---

## 1. Pure Functions

A pure function:

- Returns the same output for the same input.
- Does not modify external state.
- Does not perform side effects.

**Without FP:**

```typescript
function calculateDiscount(price: number): number {
  console.log("Calculating discount") // Side effect
  return price - Math.floor(price * 0.1)
}
```

**With FP:**

```typescript
function calculateDiscount(price: number): number {
  return price - Math.floor(price * 0.1)
}

const result = calculateDiscount(1000)
console.log(result) // 900
```

**When to use:** Use pure functions for business logic that benefits from
predictability and easy testing (price calculations, validation rules,
transformations). Push side effects (logging, network, DOM mutation) to the
edges of your application.

**Common pitfall:** A function that returns `void` and mutates an object passed
by reference is **not** pure, even if it doesn't read external state.

---

## 2. Immutability

Immutability means not changing existing data. Instead, create a new value
with the desired changes.

**Without FP:**

```typescript
class User {
  constructor(public name: string, public age: number) {}
}

const user = new User("John", 25)
user.age = 26 // Mutates the existing object
```

**With FP:**

```typescript
interface User {
  readonly name: string
  readonly age: number
}

const User = (name: string, age: number): User => ({ name, age })

const user = User("John", 25)
const updatedUser: User = { ...user, age: 26 }
// `user` is unchanged
```

Effect ships `Data.struct` / `Data.tagged` for ergonomic immutable records with
value-based equality:

```typescript
import { Data } from "effect"

class User extends Data.Class<{ readonly name: string; readonly age: number }> {}

const user = User.make({ name: "John", age: 25 })
const updated = User.make({ ...user, age: 26 })

console.log(User.equal(user, User.make({ name: "John", age: 25 }))) // true
```

**Common pitfall:** `const x = [...]` only makes the binding immutable; the
array contents can still be mutated. Use `readonly`/`ReadonlyArray<T>` for
real immutability at the type level.

---

## 3. Higher-Order Functions

A higher-order function accepts another function as an argument or returns a
function.

**Without FP:**

```typescript
function calculate(a: number, b: number): number {
  return a + b
}

function execute(a: number, b: number): number {
  return calculate(a, b)
}
```

**With FP:**

```typescript
function execute(
  a: number,
  b: number,
  operation: (a: number, b: number) => number,
): number {
  return operation(a, b)
}

const sum = execute(10, 20, (a, b) => a + b)
const product = execute(10, 20, (a, b) => a * b)
```

The same function can execute different behaviours supplied by the caller.

**When to use:** Anywhere you would otherwise branch on a flag or duplicate
near-identical functions. The classic wins: `Array.map`, retry policies,
logging hooks.

---

## 4. First-Class Functions

TypeScript functions are first-class values. You can store them in variables,
pass them to other functions, and return them.

**Without FP / OOP-style:**

```typescript
class Calculator {
  add(a: number, b: number): number {
    return a + b
  }
}

const calculator = new Calculator()
const result = calculator.add(10, 20)
```

**With FP:**

```typescript
const add: (a: number, b: number) => number = (a, b) => a + b

const result = add(10, 20)
```

Both styles are valid. The functional version treats the behaviour itself as a
value, which makes it trivial to compose and test.

---

## 5. Function Composition

Effect ships two composition helpers worth memorising:

- `pipe(input, f1, f2, ..., fN)` — feeds `input` left-to-right through a list
  of unary functions.
- `flow(f1, f2, ..., fN)` — returns a new function equivalent to
  `(x) => pipe(x, f1, f2, ..., fN)`.

**Without FP:**

```typescript
function processName(name: string): string {
  const trimmed = name.trim()
  const upper = trimmed.toUpperCase()
  return `Hello ${upper}`
}
```

**With FP-style composition:**

```typescript
import { flow } from "effect"

const trimName = (s: string): string => s.trim()
const uppercase = (s: string): string => s.toUpperCase()
const greet = (s: string): string => `Hello ${s}`

const processName = flow(trimName, uppercase, greet)

console.log(processName(" John ")) // Hello JOHN
```

Each function has one responsibility and can be reused independently.

**Common pitfall:** Functions passed to `pipe` / `flow` must be unary. To
compose a binary function, partially apply it first: `(b) => add(10, b)`.

---

## 6. Map

`map` transforms every item in a collection into another value.

**Without FP (imperative loop):**

```typescript
const numbers = [1, 2, 3, 4]
const results: number[] = []

for (const n of numbers) {
  results.push(n * 2)
}
```

**With FP-style collection processing:**

```typescript
const numbers: ReadonlyArray<number> = [1, 2, 3, 4]

const results = numbers.map((n) => n * 2)
// [2, 4, 6, 8]
```

TypeScript's built-in `Array.prototype.map` is already a functional-style
operation; it does not require Effect.

Effect also exposes `Chunk.map` for immutable, structural-sharing sequences,
and `ReadonlyArray.map` re-exports for parity.

---

## 7. Where / Filter

`filter` keeps only the elements that satisfy a condition.

**Without FP (imperative loop):**

```typescript
const numbers = [1, 2, 3, 4, 5, 6]
const results: number[] = []

for (const n of numbers) {
  if (n % 2 === 0) {
    results.push(n)
  }
}
```

**With FP-style collection processing:**

```typescript
const numbers: ReadonlyArray<number> = [1, 2, 3, 4, 5, 6]

const results = numbers.filter((n) => n % 2 === 0)
// [2, 4, 6]
```

Effect's `Array.filter` is identical in shape but returns a `ReadonlyArray`
when the input is `readonly`.

---

## 8. Fold / Reduce

`reduce` processes a collection and combines its elements into one result.

**Without FP:**

```typescript
const prices = [100, 200, 300]
let total = 0

for (const p of prices) {
  total += p
}
```

**With FP-style collection processing:**

```typescript
const prices: ReadonlyArray<number> = [100, 200, 300]

const total = prices.reduce((sum, price) => sum + price, 0)
// 600
```

The initial value is `0`; each iteration combines the accumulated value with
the next price.

Effect's `Chunk.reduce` is the immutable-friendly equivalent. Prefer it when
you want to keep all data flowing through `pipe`.

---

## 9. Option

`Option<A>` represents a value that may or may not exist.

It has two cases:

- `Some(value)`: a value exists.
- `None()`: no value exists.

**Without FP:**

```typescript
interface User {
  readonly name: string
  readonly age: number
}

function findUser(id: number): User | null {
  if (id === 1) return { name: "John", age: 25 }
  return null
}

const user = findUser(10)

if (user !== null) {
  console.log(user.name)
} else {
  console.log("User not found")
}
```

**With Effect:**

```typescript
import { Option } from "effect"

interface User {
  readonly name: string
  readonly age: number
}

function findUser(id: number): Option.Option<User> {
  if (id === 1) return Option.some({ name: "John", age: 25 })
  return Option.none()
}

const result = findUser(10)

console.log(
  Option.match(result, {
    onNone: () => "User not found",
    onSome: (user) => user.name,
  }),
)
```

`Option<User>` makes the possibility of a missing user explicit in the return
type — callers cannot forget to handle the absence.

**Common pitfall:** `Option.fromNullable(undefined)` returns `None`, which is
useful at the boundary between JS libs and your domain. Avoid using
`Option.some(null)` — that defeats the type.

---

## 10. Either

`Either<L, R>` represents one of two possible outcomes:

- `Left(error)`: conventionally used for failure.
- `Right(value)`: conventionally used for success.

> Effect's `Either` is a discriminated union — `Right` is a subtype of
> `Effect` so you can yield it directly inside `Effect.gen`. For richer error
> metadata Effect prefers `Exit`, but `Either` is fine for typed domain
> errors.

**Without FP:**

```typescript
function fetchUser(): User {
  try {
    return api.getUser()
  } catch (error) {
    throw new Error("Failed to fetch user")
  }
}
```

The caller must catch the exception.

**With Effect:**

```typescript
import { Either } from "effect"

type ApiError = { readonly _tag: "FetchFailed"; readonly cause: unknown }

function fetchUser(): Either.Either<User, ApiError> {
  try {
    const user = api.getUser()
    return Either.right(user)
  } catch (error) {
    return Either.left({ _tag: "FetchFailed", cause: error })
  }
}

const result = fetchUser()

console.log(
  Either.match(result, {
    onLeft: (err) => `Error: ${err._tag}`,
    onRight: (user) => `User: ${user.name}`,
  }),
)
```

The caller handles both outcomes explicitly. In real applications, prefer
meaningful typed error models (`ApiError` with a `_tag` literal) over plain
strings — they survive refactoring and pattern-match cleanly.

---

## 11. Map vs FlatMap

These operations are central to working with values inside containers such as
`Option` and `Either`.

- `map` transforms the contained value using a function that returns a plain
  value.
- `flatMap` (also called `Chain` in fp-ts lineage) transforms the contained
  value using a function that returns another container of the same kind,
  avoiding nested containers.

**Without FP:**

```typescript
function parseAge(value: string): number | null {
  const n = Number.parseInt(value, 10)
  return Number.isNaN(n) ? null : n
}

function validateAge(age: number): number | null {
  return age < 18 ? null : age
}

const parsed = parseAge("25")
const result = parsed === null ? null : validateAge(parsed)
```

**With Effect:**

```typescript
import { Option, pipe } from "effect"

const parseAge = (value: string): Option.Option<number> =>
  Option.fromNullable(Number.parseInt(value, 10))

const validateAge = (age: number): Option.Option<number> =>
  age < 18 ? Option.none() : Option.some(age)

const result = pipe(parseAge("25"), Option.flatMap(validateAge))
// Some(25)
```

Using `map` for a plain transformation:

```typescript
const doubled = pipe(parseAge("25"), Option.map((age) => age * 2))
// Some(50)
```

Both `map` and `flatMap` preserve the `Option` context. If you `map` a
function that *returns* an `Option`, you get `Option<Option<T>>` — almost
always a bug. Reach for `flatMap` instead.

---

## 12. Effect (synchronous side effects)

`Effect<A, E = never, R = never>` is Effect's unified type. For a
synchronous side-effecting computation it plays the role `IO<A>` plays in
fp-ts.

**Without FP:**

```typescript
function readUsername(): string {
  console.log("Reading username")
  return "John"
}

const username = readUsername() // Runs immediately
```

**With Effect:**

```typescript
import { Effect } from "effect"

const readUsername: Effect.Effect<string> = Effect.sync(() => {
  console.log("Reading username")
  return "John"
})

// Describes the computation; nothing runs yet
const username = Effect.runSync(readUsername) // Executes it
```

Other constructors worth knowing:

- `Effect.succeed(value)` — pure value, no side effect.
- `Effect.sync(thunk)` — synchronous side effect, cannot fail (typed `E =
  never`).
- `Effect.try(() => A)` — synchronous, captures thrown errors as
  `UnknownException`.

`Effect` separates *describing* a side-effecting operation from *running* it.
It does not make the side effect disappear.

**Common pitfall:** `Effect.runSync` throws if the effect is async or fails
with a non-defect error. Use `Effect.runPromise` for production code paths.

---

## 13. Effect async (Task)

`Effect<A, E, R>` also covers asynchronous computation, replacing fp-ts's
`Task<A>`.

**Without FP:**

```typescript
async function fetchUser(): Promise<User> {
  return await api.getUser()
}
```

**With Effect:**

```typescript
import { Effect } from "effect"

const fetchUser: Effect.Effect<User> = Effect.promise(() => api.getUser())
```

Usage:

```typescript
const user = await Effect.runPromise(fetchUser)
```

You can compose operations:

```typescript
import { pipe, Effect } from "effect"

const program = pipe(
  Effect.succeed(10),
  Effect.map((n) => n * 2),
)

const result = await Effect.runPromise(program) // 20
```

`Effect.promise` assumes the promise cannot reject. Use `Effect.tryPromise`
when the promise can fail and you want the error captured.

---

## 14. Effect with an error channel (TaskEither)

`Effect<A, E, R>` carries a typed error channel `E`, replacing fp-ts's
`TaskEither<L, R>`. Failure is part of the type — callers cannot ignore it.

**Without FP:**

```typescript
async function fetchUser(): Promise<User> {
  try {
    return await api.getUser()
  } catch (error) {
    throw new Error("Unable to fetch user")
  }
}
```

**With Effect:**

```typescript
import { Effect } from "effect"

class FetchError {
  readonly _tag = "FetchError"
  constructor(readonly cause: unknown) {}
}

const fetchUser: Effect.Effect<User, FetchError> = Effect.tryPromise({
  try: () => api.getUser(),
  catch: (cause) => new FetchError(cause),
})
```

Usage:

```typescript
const result = await Effect.runPromise(Effect.either(fetchUser))

console.log(
  Either.match(result, {
    onLeft: (err) => `Failed: ${err._tag}`,
    onRight: (user) => `Success: ${user.name}`,
  }),
)
```

Or handle the error channel directly on the Effect:

```typescript
const recovered = pipe(
  fetchUser,
  Effect.catchTag("FetchError", (err) =>
    Effect.succeed({ name: "anonymous", age: 0 } as User),
  ),
)
```

This is the pattern to reach for in API clients, repositories, and
authentication flows.

**Common pitfall:** Throwing inside `Effect.try` / `Effect.tryPromise` is the
*only* place Effect will catch a throw for you. Inside `Effect.sync` /
`Effect.promise` a thrown value becomes a *defect* (an unexpected crash), not
a typed error.

---

## 15. Sync with typed error (IOEither)

`Effect<A, E, R>` also replaces fp-ts's `IOEither<L, R>` — a synchronous
computation that can fail with a typed error.

**Without FP:**

```typescript
function readConfiguration(): number {
  try {
    return Number.parseInt("123", 10)
  } catch (error) {
    throw new Error("Invalid configuration")
  }
}
```

**With Effect:**

```typescript
import { Effect } from "effect"

class ConfigError {
  readonly _tag = "ConfigError"
}

const readConfiguration: Effect.Effect<number, ConfigError> = Effect.try({
  try: () => Number.parseInt("123", 10),
  catch: () => new ConfigError(),
})
```

Usage:

```typescript
const result = Effect.runSync(Effect.either(readConfiguration))

console.log(
  Either.match(result, {
    onLeft: (err) => `Error: ${err._tag}`,
    onRight: (v) => `Value: ${v}`,
  }),
)
```

The rule of thumb: if the work is sync and can fail with a typed error,
`Effect.try` is the constructor.

---

## 16. Effect Option (TaskOption)

`Effect<Option<A>>` covers async computations whose result may be absent —
the `TaskOption<A>` slot from fp-ts. Effect also has `Effect.option` /
`Option` combinators to make this ergonomic.

For example: a remote lookup may succeed but find no matching user.

**Without FP:**

```typescript
async function findUser(id: number): Promise<User | null> {
  return await database.findUser(id)
}
```

**With Effect:**

```typescript
import { Effect, Option } from "effect"

const findUser = (id: number): Effect.Effect<Option.Option<User>, FindError> =>
  pipe(
    Effect.tryPromise({
      try: () => database.findUser(id),
      catch: (cause) => new FindError(cause),
    }),
    Effect.map(Option.fromNullable),
  )
```

Usage:

```typescript
const result = await Effect.runPromise(findUser(10))

console.log(
  Option.match(result, {
    onNone: () => "User not found",
    onSome: (user) => user.name,
  }),
)
```

If absence is a normal outcome and you don't need a typed failure channel,
consider `Effect<Option<A>>`. If both absence *and* failure matter, model them
as two different `_tag`s on the error channel instead — do not collapse them.

---

## 17. Layer / Context (Reader)

`Effect<A, E, R>` carries a *requirements* channel `R` — the set of services
the effect needs to run. This replaces fp-ts's `Reader<E, A>` and gives you
compile-time-checked dependency injection.

**Without FP:**

```typescript
class UserRepository {
  constructor(private readonly api: ApiClient) {}

  getUser(): User {
    return this.api.getUser()
  }
}

const repository = new UserRepository(apiClient)
const user = repository.getUser()
```

**With Effect:**

```typescript
import { Effect, Context } from "effect"

class ApiClient extends Context.Tag("ApiClient")<ApiClient, { getUser: () => User }>() {}

const getUser: Effect.Effect<User, never, ApiClient> = Effect.gen(function* () {
  const api = yield* ApiClient
  return api.getUser()
})
```

The effect *requires* `ApiClient` in its type. You cannot run it without
providing one — the type system enforces it.

To run, provide the dependency through a `Layer`:

```typescript
import { Layer, Effect } from "effect"

const ApiClientLive = Layer.succeed(ApiClient, ApiClient.of({
  getUser: () => ({ name: "John", age: 25 }),
}))

const program = pipe(
  getUser,
  Effect.provide(ApiClientLive),
)

const user = await Effect.runPromise(program)
```

`Layer` is useful for passing dependencies explicitly, making code testable
with fakes, and tearing down resources (DB connections, HTTP clients) on
scope exit.

---

## 18. Layer + Effect + typed errors (ReaderTaskEither)

`Effect<R, E, A>` is a single type that already covers everything
`ReaderTaskEither<E, L, R>` covered in fp-ts:

- `R` is the dependency environment (`Reader`).
- async work happens automatically (`Task`).
- `E` is the typed error channel (`Either`).

**Without FP:**

```typescript
async function getUser(api: ApiClient): Promise<User> {
  try {
    return await api.getUser()
  } catch (error) {
    throw new Error(`Failed: ${error}`)
  }
}
```

**With Effect:**

```typescript
class FetchError {
  readonly _tag = "FetchError"
  constructor(readonly cause: unknown) {}
}

const getUser: Effect.Effect<User, FetchError, ApiClient> = Effect.gen(
  function* () {
    const api = yield* ApiClient
    return yield* Effect.tryPromise({
      try: () => api.getUser(),
      catch: (cause) => new FetchError(cause),
    })
  },
)
```

Usage:

```typescript
const program = pipe(
  getUser,
  Effect.provide(ApiClientLive),
)

const result = await Effect.runPromise(Effect.either(program))

console.log(
  Either.match(result, {
    onLeft: (err) => `Failed: ${err._tag}`,
    onRight: (user) => `User: ${user.name}`,
  }),
)
```

This is the pattern for repository and service layers where dependencies,
asynchronous work, and expected failures all need to be composed in one
pipeline.

---

## 19. Ref (State)

`Ref<A>` is a mutable reference whose reads and writes are `Effect`s. It
replaces fp-ts's `State<S, A>` for stateful application code.

**Without FP:**

```typescript
class Wallet {
  balance = 100

  addMoney(amount: number): void {
    this.balance += amount
  }
}

const wallet = new Wallet()
wallet.addMoney(50)
console.log(wallet.balance) // 150
```

**With Effect:**

```typescript
import { Effect, Ref } from "effect"

const addMoney = (amount: number): Effect.Effect<number, never, Ref.Ref<number>> =>
  Effect.gen(function* () {
    const ref = yield* Ref.get<Ref.Ref<number>>() // when ref is provided via env
    return yield* Ref.updateAndGet(ref, (b) => b + amount)
  })

// Simpler: create the ref and update it
const program = Effect.gen(function* () {
  const balanceRef = yield* Ref.make(100)
  yield* Ref.update(balanceRef, (b) => b + 50)
  return yield* Ref.get(balanceRef) // 150
})

console.log(Effect.runSync(program)) // 150
```

The original input state is not mutated from the outside: every update goes
through a typed `Effect`, which is inspectable, retriable, and interruptible.

**Common pitfall:** `Ref` is mutable state. Treat each `Ref` as a single
source of truth; do not copy `Ref` values around — pass the `Effect` that
reads them.

---

## 20. SynchronizedRef (async state)

`SynchronizedRef<A>` is the async-safe counterpart of `Ref`: its `update`
takes an `Effect`, so you can read from a database or another async source
inside the update callback without races.

**Without FP:**

```typescript
class AppState {
  loadingCount = 0

  async loadData(): Promise<void> {
    this.loadingCount++
    await delay(1000)
    this.loadingCount--
  }
}
```

**With Effect:**

```typescript
import { Effect, SynchronizedRef } from "effect"

const program = Effect.gen(function* () {
  const ref = yield* SynchronizedRef.make(0)

  yield* SynchronizedRef.updateEffect(ref, (count) =>
    pipe(
      Effect.succeed(count + 1),
      Effect.tap(() => Effect.sleep("1 seconds")),
      Effect.map((c) => c - 1),
    ),
  )

  return yield* SynchronizedRef.get(ref)
})

console.log(await Effect.runPromise(program)) // 0
```

Learn ordinary `Ref` first; reach for `SynchronizedRef` only when the
update itself needs to do async work.

---

## 21. Sync with Option (IOOption)

For a synchronous effect that may produce a value or none — fp-ts's
`IOOption<A>` slot — use `Effect<Option<A>>` or, more idiomatically, model the
absence as a typed error.

**Without FP:**

```typescript
function readSetting(key: string): string | null {
  return localStorage.get(key) ?? null
}
```

**With Effect (option style):**

```typescript
import { Effect, Option } from "effect"

const readSetting = (key: string): Effect.Effect<Option.Option<string>> =>
  Effect.sync(() => Option.fromNullable(localStorage.get(key) ?? null))
```

Usage:

```typescript
const result = Effect.runSync(readSetting("username"))

console.log(
  Option.match(result, {
    onNone: () => "Setting missing",
    onSome: (v) => v,
  }),
)
```

**With Effect (error-tag style):**

```typescript
class MissingSetting {
  readonly _tag = "MissingSetting"
  constructor(readonly key: string) {}
}

const readSettingOrError = (key: string): Effect.Effect<string, MissingSetting> =>
  Effect.gen(function* () {
    const raw = localStorage.get(key) ?? null
    if (raw === null) {
      return yield* Effect.fail(new MissingSetting(key))
    }
    return raw
  })
```

Useful when reading a synchronous source where a value may legitimately be
absent (env vars, in-memory cache, query params).

---

## 22. Ref — controlled mutable references (revisited)

`Ref` is Effect's answer to fp-ts's `IORef`: a mutable reference whose read
and write operations are themselves `Effect`s.

**Without FP:**

```typescript
let counter = 0

counter++
counter++

console.log(counter) // 2
```

**With Effect:**

```typescript
import { Effect, Ref } from "effect"

const program = Effect.gen(function* () {
  const counter = yield* Ref.make(0)
  yield* Ref.update(counter, (n) => n + 1)
  yield* Ref.update(counter, (n) => n + 1)
  return yield* Ref.get(counter)
})

console.log(Effect.runSync(program)) // 2
```

The central idea is that reading and writing are represented as `Effect`
operations: every mutation is visible in the type, in trace logs, and in test
assertions.

**When not to use Ref:** If the state truly belongs to a single function's
scope and never escapes, a plain `let` is fine. Reach for `Ref` when state is
shared between effects, fibers, or services.

---

## 23. Predicate

A `Predicate<A>` is a function `(a: A) => boolean`. Effect does not ship a
dedicated `Predicate` class, but the pattern is core FP — and combinators
make predicates composable.

**Without FP:**

```typescript
const isAdult = (age: number): boolean => age >= 18
const hasPermission = (age: number): boolean => age >= 21

function canAccess(age: number): boolean {
  return isAdult(age) && hasPermission(age)
}
```

**With FP combinators:**

```typescript
type Predicate<A> = (a: A) => boolean

const and = <A>(p: Predicate<A>, q: Predicate<A>): Predicate<A> =>
  (a) => p(a) && q(a)

const isAdult: Predicate<number> = (age) => age >= 18
const hasPermission: Predicate<number> = (age) => age >= 21

const canAccess = and(isAdult, hasPermission)

console.log(canAccess(25)) // true
console.log(canAccess(19)) // false
```

Predicates are useful for validation, authorization rules, and filtering
conditions. They compose with `and`, `or`, `not` — define each combinator once
and reuse it everywhere.

---

## 24. Unit / void

`void` in TypeScript is the absence of a return value, but it is not a value.
FP needs an actual value for "no meaningful result" so that types line up in
generic combinators. Effect ships `void` (the type) and the `unit` value
(`undefined`).

**Without FP:**

```typescript
function saveUser(): void {
  console.log("User saved")
}
```

**With Effect:**

```typescript
import { Effect } from "effect"

const saveUser: Effect.Effect<void> = Effect.sync(() => {
  console.log("User saved")
})
```

Use `Effect.void` when a computation needs to return a value even though
there is no meaningful result to provide — for example as the right operand of
`Effect.tap`.

```typescript
const program = pipe(
  Effect.succeed(42),
  Effect.tap((n) => Effect.sync(() => console.log(`got ${n}`))),
  Effect.as(undefined),
)
```

---

## 25. Do Notation

Do notation lets you write sequential-looking code while composing operations
that may fail or produce no value. In Effect this is `Effect.gen` with
generator functions.

**Without FP (ladder of null checks):**

```typescript
function getMarket(): Market | null {
  return null /* ... */
}

const market = getMarket()
if (market === null) process.exit(0)

const banana = market.buyBanana()
if (banana === null) process.exit(0)

const apple = market.buyApple()
if (apple === null) process.exit(0)

const pear = market.buyPear()
if (pear === null) process.exit(0)

console.log(`${banana} ${apple} ${pear}`)
```

**With Effect `Option.Do` / `Effect.gen`:**

```typescript
import { Option, Effect } from "effect"

const buyBanana = (m: Market): Option.Option<string> => Option.some("banana")
const buyApple = (m: Market): Option.Option<string> => Option.some("apple")
const buyPear = (m: Market): Option.Option<string> => Option.some("pear")

const buyAll: Effect.Effect<string, MissingItem> = Effect.gen(function* () {
  const market = yield* Effect.fromOption(getMarketOption(), () => new MissingItem("market"))
  const banana = yield* Effect.fromOption(buyBanana(market), () => new MissingItem("banana"))
  const apple = yield* Effect.fromOption(buyApple(market), () => new MissingItem("apple"))
  const pear = yield* Effect.fromOption(buyPear(market), () => new MissingItem("pear"))
  return `${banana} ${apple} ${pear}`
})

const result = await Effect.runPromise(Effect.either(buyAll))

console.log(
  Either.match(result, {
    onLeft: (err) => `Purchase failed: ${err.item}`,
    onRight: (s) => s,
  }),
)
```

`Effect.gen` is available for `Effect` and `Either`; `Option.gen` is also
supported. If any `yield*` produces a failure, the computation short-circuits
and the error propagates to the caller.

**Common pitfall:** `yield*` requires an `Effect` (or compatible yieldable)
on the right-hand side. Plain values must be wrapped: `yield* Effect.succeed(x)`.

---

## 26. Functor

A Functor is a structure that supports mapping a function over a value inside
a context. `Option` is a Functor.

**Without FP:**

```typescript
const number: number | null = 10

const result = number === null ? null : number * 2
```

**With Effect:**

```typescript
import { Option, pipe } from "effect"

const number: Option.Option<number> = Option.some(10)

const result = pipe(number, Option.map((n) => n * 2))
// Some(20)
```

The transformation happens inside the `Option` context. You do not need to
manually extract the value. `Either`, `Effect`, `Chunk`, and `ReadonlyArray`
are all functors in Effect.

---

## 27. Monad

A Monad is an abstraction that supports chaining computations where each next
computation returns a value in the same context. `Option` supports
`flatMap`.

**Without FP:**

```typescript
function parseNumber(value: string): number | null {
  const n = Number.parseInt(value, 10)
  return Number.isNaN(n) ? null : n
}

function doubleNumber(n: number): number | null {
  return n * 2
}

const parsed = parseNumber("10")
const result = parsed === null ? null : doubleNumber(parsed)
```

**With Effect:**

```typescript
import { Option, pipe } from "effect"

const parseNumber = (value: string): Option.Option<number> =>
  Option.fromNullable(Number.parseInt(value, 10))

const doubleNumber = (n: number): Option.Option<number> => Option.some(n * 2)

const result = pipe(parseNumber("10"), Option.flatMap(doubleNumber))
// Some(20)
```

A useful starting point:

- Functor → `map`
- Monad → `flatMap` (also commonly called `bind` / `Chain`)

These are abstractions describing behaviour, not separate classes you must
instantiate.

---

## 28. Immutable Collections

FP encourages avoiding mutation of collections. Create new collections when
transforming data.

**Without FP:**

```typescript
const numbers = [1, 2, 3]

numbers.push(4)
numbers.splice(numbers.indexOf(1), 1)
// `numbers` has changed
```

**With FP-style transformations:**

```typescript
const numbers: ReadonlyArray<number> = [1, 2, 3]

const added: ReadonlyArray<number> = [...numbers, 4]
// [1, 2, 3, 4]

const removed: ReadonlyArray<number> = added.filter((n) => n !== 1)
// [2, 3, 4]
```

The original list remains unchanged.

For structural-sharing immutable collections Effect ships:

- `Chunk<A>` — indexed sequence with O(1) prepend/append, good for stream
  pipelines.
- `HashSet<A>` / `HashMap<K, V>` — hash-based collections with value equality.
- `ReadonlyArray<A>` / `ReadonlyRecord<K, V>` — type-only immutability,
  zero runtime cost.

```typescript
import { Chunk } from "effect"

const c = Chunk.make(1, 2, 3)
const appended = pipe(c, Chunk.append(4)) // Chunk(1, 2, 3, 4)
const filtered = pipe(appended, Chunk.filter((n) => n !== 1)) // Chunk(2, 3, 4)
```

A `const` reference alone does **not** make a `[]` immutable; the `readonly`
modifier does.

---

## 29. Ref — `Ref.modify` for atomic read-and-write

The final example revisits `Ref` to show the atomic update-and-return pattern,
which replaces the "read-mutate-write" idiom that's unsafe under concurrency.

**Without FP (race-prone):**

```typescript
let counter = 0

function incrementAndGet(): number {
  counter += 1 // Two threads doing this concurrently can lose updates
  return counter
}
```

**With Effect:**

```typescript
import { Effect, Ref } from "effect"

const program = Effect.gen(function* () {
  const counter = yield* Ref.make(0)

  const before = yield* Ref.modify(counter, (n) => [n, n + 1] as const) // returns old value
  const after = yield* Ref.get(counter)

  return { before, after }
})

console.log(Effect.runSync(program)) // { before: 0, after: 1 }
```

`Ref.modify` returns `[result, newState]` atomically: the read and the write
happen as a single step, so no other fiber can observe an intermediate
state. `Ref.updateAndGet` is the convenience wrapper when you want the new
value instead of the old one.

**When to use which:**

- `Ref.get` + `Ref.set` — fine for independent reads and writes.
- `Ref.update` — fire-and-forget state change.
- `Ref.modify` / `Ref.updateAndGet` — when you need to return a value derived
  from the state *and* update it in one step.

---

# Quick Reference

| Concept                  | Main purpose                                            |
| ------------------------ | ------------------------------------------------------- |
| Pure Functions           | Predictable functions                                   |
| Immutability             | Avoid modifying existing data                           |
| Higher-Order Functions   | Pass functions as arguments or return them              |
| First-Class Functions    | Treat functions as values                               |
| Composition              | Combine smaller functions with `pipe` / `flow`         |
| Map                      | Transform collection values                             |
| Where / Filter           | Filter collection values                                |
| Fold / Reduce            | Reduce a collection to one result                       |
| Option                   | Handle missing values                                   |
| Either                   | Represent success or failure                            |
| FlatMap                  | Chain contextual computations                           |
| Effect (sync)            | Synchronous side effects (`Effect.sync`)                |
| Effect (async)           | Asynchronous computations (`Effect.promise`)            |
| Effect (typed error)     | Async with typed failures (`Effect.tryPromise`)         |
| Effect (sync typed err)  | Sync with typed failures (`Effect.try`)                 |
| Effect Option            | Async with optional results                             |
| Layer / Context          | Supply dependencies / environment                       |
| Layer + Effect           | Dependencies + async work + typed failures              |
| Ref                      | Model state transitions                                 |
| SynchronizedRef          | Model asynchronous state transitions                    |
| IOOption-style           | Synchronous effect with an optional result              |
| Ref (controlled)         | Represent controlled mutable references                 |
| Predicate                | Represent and compose conditions                        |
| Unit / void              | Represent an empty/meaningless result                   |
| Do notation              | Make functional chains easier to read (`Effect.gen`)    |
| Functor                  | Map over a value inside a context                       |
| Monad                    | Chain computations inside a context                     |
| Immutable Collections    | Avoid mutating collection contents                      |
| Ref.modify              | Atomic read-and-write                                   |

---

# Recommended Learning Order

Do not try to memorise every abstraction at once. Learn them in this order:

1. **Foundation:** pure functions, immutability, higher-order functions,
   composition (`pipe`, `flow`).
2. **Collection operations:** `map`, `filter`, `reduce`, predicates.
3. **Core FP types:** `Option`, `Either`, `map`, `flatMap`, Functor, Monad.
4. **Effects:** `Effect.succeed`, `Effect.sync`, `Effect.try`,
   `Effect.promise`, `Effect.tryPromise`. Understand the `Effect<A, E, R>`
   triple and how `E` and `R` are inferred from your code.
5. **Composing effects:** `Effect.gen` (do notation), `Effect.map`,
   `Effect.flatMap`, `Effect.catchTag`, `Effect.all`.
6. **Dependencies:** `Context.Tag`, `Layer.succeed`, `Layer.effect`,
   `Effect.provide`.
7. **State:** `Ref.make`, `Ref.update`, `Ref.modify`, then
   `SynchronizedRef` for async updates.
8. **Real TypeScript application:** build an API repository, authentication
   flow, error handling, and dependency management around `Effect`.

---

# What to Prioritize in Production TypeScript Projects

You do not need every Effect abstraction in every project. Prioritize:

- `Option` for values that may be absent — at boundaries with `null` /
  `undefined` and in domain models.
- `Either` for explicit success/failure results in pure code paths.
- `Effect<A, E, R>` for anything async that can fail — replaces `Promise` +
  `try/catch` + manual DI.
- `Effect.gen` (`yield*`) for readable sequencing instead of nested
  `pipe(...flatMap(...))`.
- `Layer` + `Context.Tag` for dependency injection that survives refactoring
  and is checked at compile time.
- `Ref` / `SynchronizedRef` for state shared across fibers.
- Immutability (`readonly`, `ReadonlyArray`, `Chunk`) for predictable
  application state.
- Pure functions for business logic that is easy to test.

**Adopt incrementally.** Start by using `Option` / `Either` at the boundary
of an existing Promise-based codebase. Add `Effect` for new modules where
typed errors and DI matter most. Do not rewrite working code just to make it
"more functional".

**Common pitfalls in production:**

- Calling `Effect.runPromise` too early inside business logic — keep effects
  unrun until the program's entry point.
- Using `Effect.fail(new Error("..."))` instead of a tagged class — you lose
  `Effect.catchTag` and exhaustiveness checking.
- Providing the same `Layer` twice — Effect will memoise it, but providing it
  in nested scopes can cause unexpected teardown ordering.
- Mixing `async/await` with `Effect.gen` — prefer `yield*` inside generators
  and reserve `await` for the outer `Effect.runPromise` boundary.

---

# Official Resources

- [Effect website](https://effect.website) — official docs and tutorials.
- [Effect API reference](https://effect.website/docs/reference/effect/Effect).
- [Effect on npm](https://www.npmjs.com/package/effect).
- [Effect GitHub repository](https://github.com/Effect-TS/effect).
- [Effect Discord](https://discord.gg/effect-ts).
- [Effect Days conference talks](https://www.youtube.com/@effect-ts).

> **Version note:** This guide is written against Effect v3 / v4. Some
> constructor names changed between major versions (notably the move from
> `Effect.tryCatch` to `Effect.try` with the `{ try, catch }` object form).
> Always check the API docs for your installed `effect` version before
> relying on a snippet, and treat the examples marked illustrative as
> learning guides rather than guaranteed copy-paste code.
