package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Todo struct {
	UserID    int    `json:"userId"`
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func fetchTodo(ctx context.Context, client *http.Client, id int) (Todo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://jsonplaceholder.typicode.com/todos/%d", id),
		nil,
	)
	if err != nil {
		return Todo{}, err
	}

	response, err := client.Do(request)
	if err != nil {
		return Todo{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Todo{}, fmt.Errorf("request failed with status %s", response.Status)
	}

	var todo Todo
	if err := json.NewDecoder(response.Body).Decode(&todo); err != nil {
		return Todo{}, err
	}

	return todo, nil
}

func backgroundLog(client *http.Client, id int) {
	todo, err := fetchTodo(context.Background(), client, id)
	if err != nil {
		fmt.Println("[spawn] background request failed:", err)
		return
	}
	fmt.Printf("[spawn] background todo %d: %s\n", todo.ID, todo.Title)
}

func heavyCPUWork(limit int) uint64 {
	var sum uint64
	for i := 0; i <= limit; i++ {
		sum += uint64(i) * uint64(i)
	}
	return sum
}

func printTodo(label string, todo Todo) {
	fmt.Printf("[%s] %d: %s (completed: %t)\n", label, todo.ID, todo.Title, todo.Completed)
}

func main() {
	client := &http.Client{}
	ctx := context.Background()

	go backgroundLog(client, 9)

	fmt.Println("[await] fetching sequentially")
	first, err := fetchTodo(ctx, client, 1)
	if err != nil {
		fmt.Println("[await] error:", err)
		return
	}
	second, err := fetchTodo(ctx, client, 2)
	if err != nil {
		fmt.Println("[await] error:", err)
		return
	}
	printTodo("await", first)
	printTodo("await", second)

	fmt.Println("[join] fetching concurrently")
	ids := []int{3, 4}
	todos := make([]Todo, len(ids))
	errs := make([]error, len(ids))

	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func(index, todoID int) {
			defer wg.Done()
			todos[index], errs[index] = fetchTodo(ctx, client, todoID)
		}(i, id)
	}
	wg.Wait()

	for i := range ids {
		if errs[i] != nil {
			fmt.Println("[join] error:", errs[i])
			continue
		}
		printTodo("join", todos[i])
	}

	fmt.Println("[select] racing request against a 2s timeout")
	raceCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	type raceResult struct {
		todo Todo
		err  error
	}
	results := make(chan raceResult, 1)
	go func() {
		todo, err := fetchTodo(raceCtx, client, 5)
		results <- raceResult{todo: todo, err: err}
	}()

	select {
	case result := <-results:
		if result.err != nil {
			fmt.Println("[select] request error:", result.err)
		} else {
			printTodo("select", result.todo)
		}
	case <-raceCtx.Done():
		fmt.Println("[select] timeout:", raceCtx.Err())
	}

	fmt.Println("[blocking] running CPU-bound work in a goroutine")
	done := make(chan uint64, 1)
	go func() {
		done <- heavyCPUWork(10_000_000)
	}()
	fmt.Println("[blocking] checksum =", <-done)

	time.Sleep(500 * time.Millisecond)
}
