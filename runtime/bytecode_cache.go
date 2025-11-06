package runtime

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sync"
	"time"

	"github.com/deicod/gojinja/nodes"
)

// BytecodeArtifact represents the serialized template data stored in a bytecode cache.
// It contains the processed AST along with metadata needed to validate cache entries.
type BytecodeArtifact struct {
	AST          *nodes.Template
	ParentBlocks map[string]*nodes.Block
	Dependencies map[string]time.Time
	EnvSignature string
	GeneratedAt  time.Time
}

// BytecodeCache provides an abstraction similar to Jinja2's bytecode cache API.
// Implementations are responsible for persisting compiled template data between
// environment instances or process runs.
type BytecodeCache interface {
	// Load retrieves the cached artifact for the given key. A nil artifact with
	// a nil error indicates a cache miss.
	Load(key string) (*BytecodeArtifact, error)

	// Store persists the artifact for the given key.
	Store(key string, artifact *BytecodeArtifact) error

	// Remove deletes the cached artifact for the given key, ignoring missing
	// entries.
	Remove(key string) error

	// Clear removes all cached artifacts.
	Clear() error
}

// MemoryBytecodeCache stores serialized templates in-process using gob encoding.
// It mirrors Jinja2's default in-memory cache and offers deterministic cloning of
// the AST by serializing entries when stored.
type MemoryBytecodeCache struct {
	mu    sync.RWMutex
	items map[string][]byte
}

// NewMemoryBytecodeCache creates an empty bytecode cache backed by memory.
func NewMemoryBytecodeCache() *MemoryBytecodeCache {
	return &MemoryBytecodeCache{
		items: make(map[string][]byte),
	}
}

// Load retrieves a cached artifact from memory.
func (c *MemoryBytecodeCache) Load(key string) (*BytecodeArtifact, error) {
	c.mu.RLock()
	data, ok := c.items[key]
	c.mu.RUnlock()

	if !ok {
		return nil, nil
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	var artifact BytecodeArtifact
	if err := decoder.Decode(&artifact); err != nil {
		return nil, fmt.Errorf("decode bytecode artifact: %w", err)
	}

	return &artifact, nil
}

// Store persists a compiled artifact in memory.
func (c *MemoryBytecodeCache) Store(key string, artifact *BytecodeArtifact) error {
	if artifact == nil {
		return fmt.Errorf("bytecode artifact cannot be nil")
	}

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(artifact); err != nil {
		return fmt.Errorf("encode bytecode artifact: %w", err)
	}

	c.mu.Lock()
	c.items[key] = buf.Bytes()
	c.mu.Unlock()

	return nil
}

// Remove deletes the cached artifact for the provided key.
func (c *MemoryBytecodeCache) Remove(key string) error {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
	return nil
}

// Clear removes all cached artifacts.
func (c *MemoryBytecodeCache) Clear() error {
	c.mu.Lock()
	c.items = make(map[string][]byte)
	c.mu.Unlock()
	return nil
}
