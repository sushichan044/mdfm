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
