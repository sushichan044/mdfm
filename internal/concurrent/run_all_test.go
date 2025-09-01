package concurrent_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sushichan044/fmx/internal/concurrent"
)

func TestRunAll_SuccessAndError(t *testing.T) {
	results := concurrent.RunAll(
		func() (int, error) { return 1, nil },
		func() (int, error) { return 0, errors.New("boom") },
		func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 42, nil
		},
	)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Value != 1 || results[0].Err != nil {
		t.Fatalf("unexpected result[0]: %+v", results[0])
	}

	if results[1].Err == nil {
		t.Fatalf("expected rejection at [1], got: %+v", results[1])
	}

	if results[2].Value != 42 || results[2].Err != nil {
		t.Fatalf("unexpected result[2]: %+v", results[2])
	}
}

func TestRunAll_PanicRecovery(t *testing.T) {
	results := concurrent.RunAll(
		func() (string, error) { return "ok", nil },
		func() (string, error) { panic("kaboom") },
	)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].Value != "ok" || results[0].Err != nil {
		t.Fatalf("unexpected result[0]: %+v", results[0])
	}

	if results[1].Err == nil || results[1].Err.Error() == "" {
		t.Fatalf("expected rejection with panic error at [1], got: %+v", results[1])
	}
}
