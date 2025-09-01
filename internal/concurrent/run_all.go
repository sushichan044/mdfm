package concurrent

import (
	"fmt"
	"sync"
)

type (
	// result represents the outcome of a task execution.
	result[T any] struct {
		Value T
		Err   error
	}

	// TaskResult combines task metadata with its execution result.
	TaskResult[T, M any] struct {
		Metadata M
		Result   result[T]
	}

	// Task combines a task function with its metadata.
	Task[T, M any] struct {
		Metadata M
		Run      func() (T, error)
	}
)

// RunAll runs all given tasks with metadata concurrently and waits for all of them to finish.
// It does not fail fast: even if some tasks return an error or panic, the others keep running.
// The returned slice preserves the order of the input tasks.
//
// Each task includes metadata and a function returning (T, error). Panics inside tasks are recovered and
// exposed as errors in the corresponding Result with a message prefixed by "panic:".
//
// Concurrency safety: each goroutine writes to a distinct index in the results slice.
func RunAll[T, M any](tasks ...Task[T, M]) []TaskResult[T, M] {
	n := len(tasks)
	results := make([]TaskResult[T, M], n)

	var wg sync.WaitGroup
	wg.Add(n)

	for i, task := range tasks {
		go func(i int, task Task[T, M]) {
			defer wg.Done()

			var zero T
			// Recover panic and convert into error.
			defer func() {
				if rec := recover(); rec != nil {
					results[i] = TaskResult[T, M]{
						Metadata: task.Metadata,
						Result: result[T]{
							Value: zero,
							Err:   fmt.Errorf("panic: %v", rec),
						},
					}
				}
			}()

			v, err := task.Run()
			results[i] = TaskResult[T, M]{
				Metadata: task.Metadata,
				Result: result[T]{
					Value: v,
					Err:   err,
				},
			}
		}(i, task)
	}

	wg.Wait()
	return results
}
