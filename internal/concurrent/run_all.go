package concurrent

import (
	"fmt"
	"sync"
)

// TaskResult is the outcome of a single task.
type TaskResult[T any] struct {
	Value T
	Err   error
}

// RunAll runs all given tasks concurrently and waits for all of them to finish.
// It does not fail fast: even if some tasks return an error or panic, the others keep running.
// The returned slice preserves the order of the input tasks.
//
// Each task is a function returning (T, error). Panics inside tasks are recovered and
// exposed as errors in the corresponding Result with a message prefixed by "panic:".
//
// Concurrency safety: each goroutine writes to a distinct index in the results slice.
func RunAll[T any](tasks ...func() (T, error)) []TaskResult[T] {
	n := len(tasks)
	results := make([]TaskResult[T], n)

	var wg sync.WaitGroup
	wg.Add(n)

	for i, task := range tasks {
		go func(i int, task func() (T, error)) {
			defer wg.Done()

			var zero T
			// Recover panic and convert into error.
			defer func() {
				if rec := recover(); rec != nil {
					results[i] = TaskResult[T]{
						Value: zero,
						Err:   fmt.Errorf("panic: %v", rec),
					}
				}
			}()

			v, err := task()
			if err != nil {
				results[i] = TaskResult[T]{
					Value: zero,
					Err:   err,
				}
				return
			}
			results[i] = TaskResult[T]{
				Value: v,
				Err:   nil,
			}
		}(i, task)
	}

	wg.Wait()
	return results
}
