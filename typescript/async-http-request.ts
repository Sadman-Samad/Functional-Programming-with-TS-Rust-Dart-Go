type Todo = {
  userId: number;
  id: number;
  title: string;
  completed: boolean;
};

async function fetchTodo(id: number, signal?: AbortSignal): Promise<Todo> {
  const response = await fetch(`https://jsonplaceholder.typicode.com/todos/${id}`, { signal });

  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  return (await response.json()) as Todo;
}

async function backgroundLog(id: number): Promise<void> {
  try {
    const todo = await fetchTodo(id);
    console.log(`[spawn] background todo ${todo.id}: ${todo.title}`);
  } catch (error) {
    console.error("[spawn] background request failed:", error);
  }
}

function printTodo(label: string, todo: Todo): void {
  console.log(`[${label}] ${todo.id}: ${todo.title} (completed: ${todo.completed})`);
}

async function main(): Promise<void> {
  void backgroundLog(9);

  console.log("[await] fetching sequentially");
  const first = await fetchTodo(1);
  const second = await fetchTodo(2);
  printTodo("await", first);
  printTodo("await", second);

  console.log("[join] fetching concurrently");
  const [third, fourth] = await Promise.all([fetchTodo(3), fetchTodo(4)]);
  printTodo("join", third);
  printTodo("join", fourth);

  console.log("[select] racing request against a 2s timeout");
  const controller = new AbortController();
  let timer: ReturnType<typeof setTimeout>;
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error("timeout after 2s")), 2000);
  });

  try {
    const winner = await Promise.race([fetchTodo(5, controller.signal), timeout]);
    printTodo("select", winner);
  } catch (error) {
    controller.abort();
    console.log("[select]", error instanceof Error ? error.message : error);
  } finally {
    clearTimeout(timer!);
  }

  console.log("[offload] running CPU-bound work without blocking the event loop");
  const payload = new Uint8Array(5_000_000);
  const digest = await crypto.subtle.digest("SHA-256", payload);
  const checksum = Array.from(new Uint8Array(digest).slice(0, 8), (byte) =>
    byte.toString(16).padStart(2, "0"),
  ).join("");
  console.log("[offload] sha256 =", checksum);

  await new Promise((resolve) => setTimeout(resolve, 500));
}

main().catch((error) => console.error(error));
