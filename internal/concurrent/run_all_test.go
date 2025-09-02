package concurrent_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sushichan044/mdfm/internal/concurrent"
)

func TestRunAll_SuccessAndError(t *testing.T) {
	tasks := []concurrent.Task[int, string]{
		{
			Metadata: "task-1",
			Run:      func() (int, error) { return 1, nil },
		},
		{
			Metadata: "task-2",
			Run:      func() (int, error) { return 0, errors.New("boom") },
		},
		{
			Metadata: "task-3",
			Run: func() (int, error) {
				time.Sleep(10 * time.Millisecond)
				return 42, nil
			},
		},
	}
	results := concurrent.RunAll(tasks, concurrent.WithMaxConcurrency(2))

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Metadata != "task-1" || results[0].Result.Value != 1 || results[0].Result.Err != nil {
		t.Fatalf("unexpected result[0]: %+v", results[0])
	}

	if results[1].Metadata != "task-2" || results[1].Result.Err == nil {
		t.Fatalf("expected rejection at [1], got: %+v", results[1])
	}

	if results[2].Metadata != "task-3" || results[2].Result.Value != 42 || results[2].Result.Err != nil {
		t.Fatalf("unexpected result[2]: %+v", results[2])
	}
}

func TestRunAll_PanicRecovery(t *testing.T) {
	tasks := []concurrent.Task[string, string]{
		{
			Metadata: "good-task",
			Run:      func() (string, error) { return "ok", nil },
		},
		{
			Metadata: "panic-task",
			Run:      func() (string, error) { panic("kaboom") },
		},
	}
	results := concurrent.RunAll(tasks, concurrent.WithMaxConcurrency(2))

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Metadata != "good-task" || results[0].Result.Value != "ok" || results[0].Result.Err != nil {
		t.Fatalf("unexpected result[0]: %+v", results[0])
	}

	if results[1].Metadata != "panic-task" || results[1].Result.Err == nil || results[1].Result.Err.Error() == "" {
		t.Fatalf("expected rejection with panic error at [1], got: %+v", results[1])
	}
}

func TestRunAllStream_SuccessAndError(t *testing.T) {
	tasks := []concurrent.Task[int, string]{
		{
			Metadata: "task-1",
			Run:      func() (int, error) { return 1, nil },
		},
		{
			Metadata: "task-2",
			Run:      func() (int, error) { return 0, errors.New("boom") },
		},
		{
			Metadata: "task-3",
			Run: func() (int, error) {
				time.Sleep(10 * time.Millisecond)
				return 42, nil
			},
		},
	}

	resultChan := concurrent.RunAllStream(tasks, concurrent.WithMaxConcurrency(2))

	results := make([]concurrent.TaskExecution[int, string], 0, 3)
	for result := range resultChan {
		results = append(results, result)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Results may come in any order due to concurrency, so we need to check by metadata
	resultMap := make(map[string]concurrent.TaskExecution[int, string])
	for _, result := range results {
		resultMap[result.Metadata] = result
	}

	if result, ok := resultMap["task-1"]; !ok || result.Result.Value != 1 || result.Result.Err != nil {
		t.Fatalf("unexpected result for task-1: %+v", result)
	}

	if result, ok := resultMap["task-2"]; !ok || result.Result.Err == nil {
		t.Fatalf("expected rejection for task-2, got: %+v", result)
	}

	if result, ok := resultMap["task-3"]; !ok || result.Result.Value != 42 || result.Result.Err != nil {
		t.Fatalf("unexpected result for task-3: %+v", result)
	}
}

func TestRunAllStream_PanicRecovery(t *testing.T) {
	tasks := []concurrent.Task[string, string]{
		{
			Metadata: "good-task",
			Run:      func() (string, error) { return "ok", nil },
		},
		{
			Metadata: "panic-task",
			Run:      func() (string, error) { panic("kaboom") },
		},
	}

	resultChan := concurrent.RunAllStream(tasks, concurrent.WithMaxConcurrency(2))

	results := make([]concurrent.TaskExecution[string, string], 0, 2)
	for result := range resultChan {
		results = append(results, result)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Results may come in any order due to concurrency, so we need to check by metadata
	resultMap := make(map[string]concurrent.TaskExecution[string, string])
	for _, result := range results {
		resultMap[result.Metadata] = result
	}

	if result, ok := resultMap["good-task"]; !ok || result.Result.Value != "ok" || result.Result.Err != nil {
		t.Fatalf("unexpected result for good-task: %+v", result)
	}

	if result, ok := resultMap["panic-task"]; !ok || result.Result.Err == nil || result.Result.Err.Error() == "" {
		t.Fatalf("expected rejection with panic error for panic-task, got: %+v", result)
	}
}

func TestRunAllStream_ChannelClosure(t *testing.T) {
	tasks := []concurrent.Task[int, string]{
		{
			Metadata: "task-1",
			Run:      func() (int, error) { return 1, nil },
		},
	}

	resultChan := concurrent.RunAllStream(tasks)

	// Consume the channel
	count := 0
	for range resultChan {
		count++
	}

	if count != 1 {
		t.Fatalf("expected 1 result, got %d", count)
	}

	// Channel should be closed now, so receiving should return zero value and false
	if result, ok := <-resultChan; ok {
		t.Fatalf("expected channel to be closed, but received: %+v", result)
	}
}

func TestRunAllStream_EmptyTasks(t *testing.T) {
	var tasks []concurrent.Task[int, string]

	resultChan := concurrent.RunAllStream(tasks)

	// Channel should be closed immediately
	count := 0
	for range resultChan {
		count++
	}

	if count != 0 {
		t.Fatalf("expected 0 results, got %d", count)
	}
}
