package runtime

import (
	"fmt"
	"sync"
)

// Namespace provides a mutable attribute container similar to Jinja's namespace helper.
type Namespace struct {
	mu     sync.RWMutex
	values map[string]interface{}
}

// NewNamespace creates a namespace populated with the provided values.
func NewNamespace(initial map[string]interface{}) *Namespace {
	ns := &Namespace{values: make(map[string]interface{})}
	if initial != nil {
		for k, v := range initial {
			ns.values[k] = v
		}
	}
	return ns
}

// Get returns the stored value for a key and a flag indicating if it existed.
func (ns *Namespace) Get(name string) (interface{}, bool) {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	value, ok := ns.values[name]
	return value, ok
}

// Set stores a value under the given key and returns nil for template compatibility.
func (ns *Namespace) Set(name string, value interface{}) interface{} {
	ns.mu.Lock()
	ns.values[name] = value
	ns.mu.Unlock()
	return nil
}

// Update merges the provided mapping into the namespace and returns nil.
func (ns *Namespace) Update(values interface{}) interface{} {
	if values == nil {
		return nil
	}
	if m, ok := toStringInterfaceMap(values); ok {
		ns.mu.Lock()
		for k, v := range m {
			ns.values[k] = v
		}
		ns.mu.Unlock()
	}
	return nil
}

// Items returns a shallow copy of the namespace values.
func (ns *Namespace) Items() map[string]interface{} {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	copyMap := make(map[string]interface{}, len(ns.values))
	for k, v := range ns.values {
		copyMap[k] = v
	}
	return copyMap
}

func (ns *Namespace) String() string {
	ns.mu.RLock()
	defer ns.mu.RUnlock()
	return fmt.Sprintf("namespace(%v)", ns.values)
}
