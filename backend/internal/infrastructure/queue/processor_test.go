package queue

import (
	"context"
	"testing"
)

type stubProcessor struct {
	jobType    string
	processErr error
	called     bool
}

func (s *stubProcessor) Type() string { return s.jobType }
func (s *stubProcessor) Process(_ context.Context, _ *SyncJobPayload) error {
	s.called = true
	return s.processErr
}

func TestProcessorRegistry(t *testing.T) {
	reg := NewProcessorRegistry()

	txProc := &stubProcessor{jobType: "transaction_sync"}
	reviewProc := &stubProcessor{jobType: "review_sync"}

	reg.Register(txProc)
	reg.Register(reviewProc)

	// Get existing processor
	got, err := reg.Get("transaction_sync")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != txProc {
		t.Error("Got wrong processor")
	}

	// Get non-existent processor
	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent processor")
	}
}

func TestProcessorRegistryOverwrite(t *testing.T) {
	reg := NewProcessorRegistry()

	proc1 := &stubProcessor{jobType: "transaction_sync"}
	proc2 := &stubProcessor{jobType: "transaction_sync"}

	reg.Register(proc1)
	reg.Register(proc2)

	got, _ := reg.Get("transaction_sync")
	if got != proc2 {
		t.Error("Expected second registration to overwrite first")
	}
}
