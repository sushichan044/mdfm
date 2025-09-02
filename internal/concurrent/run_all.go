package concurrent

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/semaphore"
)

type (
	// Task combines a task function with its metadata.
	Task[T, M any] struct {
		Metadata M
		Run      func() (T, error)
	}

	// TaskExecution combines task metadata with its execution result.
	TaskExecution[T, M any] struct {
		Metadata M
		Result   taskResult[T]
	}

	// taskResult represents the outcome of a task execution.
	taskResult[T any] struct {
		Value T
		Err   error
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
func RunAll[T, M any](tasks []Task[T, M], options ...ConcurrencyOptions) []TaskExecution[T, M] {
	opts := setOpts(options...)

	ctx := context.Background()
	sem := semaphore.NewWeighted(opts.maxConcurrency)

	var wg sync.WaitGroup
	results := make([]TaskExecution[T, M], len(tasks))

	for i, task := range tasks {
		wg.Add(1)

		go func(i int, task Task[T, M]) {
			defer wg.Done()
			var zero T

			if err := sem.Acquire(ctx, 1); err != nil {
				results[i] = TaskExecution[T, M]{
					Metadata: task.Metadata,
					Result: taskResult[T]{
						Value: zero,
						Err:   fmt.Errorf("semaphore acquire failed: %w", err),
					},
				}
				return
			}
			defer sem.Release(1)

			// Recover panic and convert into error.
			defer func() {
				if rec := recover(); rec != nil {
					results[i] = TaskExecution[T, M]{
						Metadata: task.Metadata,
						Result: taskResult[T]{
							Value: zero,
							Err:   fmt.Errorf("panic: %v", rec),
						},
					}
				}
			}()

			v, err := task.Run()
			results[i] = TaskExecution[T, M]{
				Metadata: task.Metadata,
				Result: taskResult[T]{
					Value: v,
					Err:   err,
				},
			}
		}(i, task)
	}

	wg.Wait()
	return results
}

// RunAllStream runs all given tasks with metadata concurrently and streams results as they complete.
// Unlike RunAll, this function returns a channel that receives results as soon as they are available,
// without waiting for all tasks to finish first. The channel is closed when all tasks complete.
//
// It does not fail fast: even if some tasks return an error or panic, the others keep running.
// Results are streamed in completion order, not input order.
//
// Each task includes metadata and a function returning (T, error). Panics inside tasks are recovered and
// exposed as errors in the corresponding Result with a message prefixed by "panic:".
//
// The returned channel should be consumed until it's closed to avoid goroutine leaks.
func RunAllStream[T, M any](tasks []Task[T, M], options ...ConcurrencyOptions) <-chan TaskExecution[T, M] {
	opts := setOpts(options...)

	ctx := context.Background()
	sem := semaphore.NewWeighted(opts.maxConcurrency)

	resultChan := make(chan TaskExecution[T, M], len(tasks))

	var wg sync.WaitGroup

	for _, task := range tasks {
		wg.Add(1)

		go func(task Task[T, M]) {
			defer wg.Done()
			var zero T

			if err := sem.Acquire(ctx, 1); err != nil {
				resultChan <- TaskExecution[T, M]{
					Metadata: task.Metadata,
					Result: taskResult[T]{
						Value: zero,
						Err:   fmt.Errorf("semaphore acquire failed: %w", err),
					},
				}
				return
			}
			defer sem.Release(1)

			// Recover panic and convert into error.
			defer func() {
				if rec := recover(); rec != nil {
					resultChan <- TaskExecution[T, M]{
						Metadata: task.Metadata,
						Result: taskResult[T]{
							Value: zero,
							Err:   fmt.Errorf("panic: %v", rec),
						},
					}
				}
			}()

			v, err := task.Run()
			resultChan <- TaskExecution[T, M]{
				Metadata: task.Metadata,
				Result: taskResult[T]{
					Value: v,
					Err:   err,
				},
			}
		}(task)
	}

	// Close the channel when all tasks are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}
