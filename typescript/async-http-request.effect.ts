import { Data, Effect, Fiber } from "effect";

// Effect.Effect<A, E, R> is a lazy description of a program:
//   A = value produced on success
//   E = expected error type (like the error side of Rust's Result)
//   R = services/dependencies needed to run (never = needs nothing)
//
// Creating an Effect does NOT run it. Nothing happens until runPromise,
// runSync, or runFork is called at the bottom of this file.

// Plain data shape returned by the API.
type Todo = {
  userId: number;
  id: number;
  title: string;
  completed: boolean;
};

// Data.TaggedError creates an error class with a literal "_tag" field.
// That tag is what Effect.catchTag("TimeoutError", ...) matches on.
// These errors are values, so they appear in the Effect's E type instead of
// being thrown and silently escaping the function's type.
class HttpError extends Data.TaggedError("HttpError")<{
  readonly status: number;
}> {}

class NetworkError extends Data.TaggedError("NetworkError")<{
  readonly cause: unknown;
}> {}

// fetchTodo has the type:
//   (id: number) => Effect<Todo, HttpError | NetworkError>
//
// This function is pure: calling it only builds a description. No network
// request is sent until the returned Effect is actually run.
const fetchTodo = (
  id: number,
): Effect.Effect<Todo, HttpError | NetworkError> =>
  // Effect.gen(function* () { ... }) is Effect's "do notation".
  // Read `yield*` as "await this Effect and unwrap its success value".
  Effect.gen(function* () {
    // Effect.tryPromise wraps a Promise that may reject.
    // - `try` receives an AbortSignal from Effect. If this Effect is
    //   interrupted (for example by Effect.timeout), the signal is aborted
    //   automatically, so fetch cancels the HTTP request.
    // - `catch` maps the rejected value into our typed NetworkError.
    const response = yield* Effect.tryPromise({
      try: (signal) =>
        fetch(`https://jsonplaceholder.typicode.com/todos/${id}`, { signal }),
      catch: (cause) => new NetworkError({ cause }),
    });

    // Effect.fail puts a value into the ERROR channel.
    // `return yield*` short-circuits this generator, similar to throwing,
    // but the failure is typed and visible in the function's return type.
    if (!response.ok) {
      return yield* Effect.fail(new HttpError({ status: response.status }));
    }

    // response.json() is another Promise, so wrap it the same way.
    // The cast tells TypeScript the parsed JSON follows the Todo shape.
    return yield* Effect.tryPromise({
      try: () => response.json() as Promise<Todo>,
      catch: (cause) => new NetworkError({ cause }),
    });
  });

// backgroundLog returns Effect<void> with error type never because it handles
// its own expected errors.
//
// pipe chains combinators left to right, passing the previous Effect's result
// into the next one.
const backgroundLog = (id: number): Effect.Effect<void> =>
  fetchTodo(id).pipe(
    // tap runs a side effect only if fetchTodo succeeds, then passes the
    // original Todo through unchanged. It is useful for logging.
    Effect.tap((todo) =>
      Effect.log(`[spawn] background todo ${todo.id}: ${todo.title}`),
    ),
    // catch handles any typed error. Here we log it instead of failing, which
    // is why this Effect's error type becomes never.
    Effect.catch((error) =>
      Effect.logError("[spawn] background request failed:", error),
    ),
  );

// Effect.log is an Effect that logs a message and succeeds with undefined.
const printTodo = (label: string, todo: Todo): Effect.Effect<void> =>
  Effect.log(
    `[${label}] ${todo.id}: ${todo.title} (completed: ${todo.completed})`,
  );

// `program` describes the whole application, but it is still just a value.
// Nothing inside this generator runs at module load time.
const program = Effect.gen(function* () {
  // Effect.forkDetach starts backgroundLog concurrently and returns a Fiber
  // handle. It is similar to `void backgroundLog(9)`, but the Fiber can be
  // joined or interrupted; no unhandled Promise is floating around.
  const background = yield* Effect.forkDetach(backgroundLog(9));

  // yield* fetchTodo(1) is like await, but a failure short-circuits the
  // generator into the typed error channel instead of throwing.
  yield* Effect.log("[await] fetching sequentially");
  const first = yield* fetchTodo(1);
  const second = yield* fetchTodo(2);
  yield* printTodo("await", first);
  yield* printTodo("await", second);

  // Effect.all is the applicative/concurrent combinator.
  // By default it runs sequentially. concurrency: "unbounded" starts both
  // fetches immediately and joins their results, like Promise.all.
  yield* Effect.log("[join] fetching concurrently");
  const [third, fourth] = yield* Effect.all(
    [fetchTodo(3), fetchTodo(4)],
    { concurrency: "unbounded" },
  );
  yield* printTodo("join", third);
  yield* printTodo("join", fourth);

  // Effect.timeout interrupts fetchTodo(5) if it takes longer than 2 seconds.
  // Because Effect.tryPromise received an AbortSignal, the HTTP request is
  // aborted too; no manual AbortController, timer, or clearTimeout is needed.
  //
  // tap prints the Todo only on success and passes it through.
  // catchTag handles only the TimeoutError; other errors still propagate.
  yield* Effect.log("[select] racing request against a 2s timeout");
  yield* fetchTodo(5).pipe(
    Effect.timeout("2 seconds"),
    Effect.tap((todo) => printTodo("select", todo)),
    Effect.catchTag("TimeoutError", (error) =>
      Effect.log("[select]", error.message),
    ),
  );

  // Effect.promise wraps a Promise that is not expected to reject.
  // This part is conceptually the same as the original async version:
  // WebCrypto does the work asynchronously, and Effect waits for it.
  yield* Effect.log("[offload] running CPU-bound work without blocking the event loop");
  const payload = new Uint8Array(5_000_000);
  const digest = yield* Effect.promise(() =>
    crypto.subtle.digest("SHA-256", payload),
  );
  const checksum = Array.from(new Uint8Array(digest).slice(0, 8), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
  yield* Effect.log("[offload] sha256 =", checksum);

  // Effect.sleep is a non-blocking sleep managed by the Effect runtime.
  yield* Effect.sleep("500 millis");

  // Fiber.join waits for the detached background fiber to finish before the
  // main program ends. This makes the background work observable.
  yield* Fiber.join(background);
});

// This is the boundary where the description becomes action.
// Everything above is pure construction; this line actually performs IO.
// runPromise returns a Promise, so we attach catch for unexpected defects
// or errors that escaped the typed error channel.
Effect.runPromise(program).catch((error) => console.error(error));
