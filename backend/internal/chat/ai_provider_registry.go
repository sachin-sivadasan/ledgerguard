package chat

import (
	"fmt"
	"sync"
)

// AIProviderRegistry manages multiple AI providers and resolves them by name.
type AIProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]AIClient
	fallback  string // default provider name
}

// NewAIProviderRegistry creates a registry with the given default provider name.
func NewAIProviderRegistry(fallback string) *AIProviderRegistry {
	return &AIProviderRegistry{
		providers: make(map[string]AIClient),
		fallback:  fallback,
	}
}

// Register adds an AI provider by name.
func (r *AIProviderRegistry) Register(name string, client AIClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = client
}

// Get returns the provider by name, falling back to the default if name is empty.
func (r *AIProviderRegistry) Get(name string) (AIClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if name == "" {
		name = r.fallback
	}

	client, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("AI provider %q not registered", name)
	}
	return client, nil
}

// Default returns the fallback provider.
func (r *AIProviderRegistry) Default() (AIClient, error) {
	return r.Get("")
}
