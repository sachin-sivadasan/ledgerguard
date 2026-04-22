package queue

import (
	"context"
	"fmt"
	"sync"
)

// SyncProcessor is the interface that all sync job processors must implement
type SyncProcessor interface {
	// Type returns the job type this processor handles (e.g., "transaction_sync")
	Type() string

	// Process executes the sync job. Should check IsCancelled periodically.
	Process(ctx context.Context, payload *SyncJobPayload) error
}

// ProcessorRegistry manages sync processors by job type
type ProcessorRegistry struct {
	mu         sync.RWMutex
	processors map[string]SyncProcessor
}

// NewProcessorRegistry creates an empty processor registry
func NewProcessorRegistry() *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: make(map[string]SyncProcessor),
	}
}

// Register adds a processor for a given job type
func (r *ProcessorRegistry) Register(p SyncProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[p.Type()] = p
}

// Get returns the processor for the given job type, or an error if not found
func (r *ProcessorRegistry) Get(jobType string) (SyncProcessor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.processors[jobType]
	if !ok {
		return nil, fmt.Errorf("no processor registered for job type %q", jobType)
	}
	return p, nil
}
