package runtime

import (
	"sync"
	"testing"
	"time"
)

type countingLoader struct {
	mu      sync.Mutex
	source  string
	modTime time.Time
	loads   int
}

func newCountingLoader(source string) *countingLoader {
	return &countingLoader{
		source:  source,
		modTime: time.Now().UTC(),
	}
}

func (l *countingLoader) Load(name string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loads++
	return l.source, nil
}

func (l *countingLoader) TemplateModTime(name string) (time.Time, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.modTime, nil
}

func (l *countingLoader) setSource(source string) {
	l.mu.Lock()
	l.source = source
	l.modTime = l.modTime.Add(time.Minute)
	l.mu.Unlock()
}

func (l *countingLoader) loadCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.loads
}

func TestMemoryBytecodeCacheReuse(t *testing.T) {
	loader := newCountingLoader("Hello {{ name }}")

	env := NewEnvironment()
	env.SetLoader(loader)
	env.SetBytecodeCache(NewMemoryBytecodeCache())

	tmpl, err := env.LoadTemplate("greeting.html")
	if err != nil {
		t.Fatalf("LoadTemplate error: %v", err)
	}

	out, err := tmpl.ExecuteToString(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("ExecuteToString error: %v", err)
	}
	if got, want := out, "Hello Go"; got != want {
		t.Fatalf("unexpected render output: got %q want %q", got, want)
	}

	if count := loader.loadCount(); count != 1 {
		t.Fatalf("expected loader to be invoked once, got %d", count)
	}

	env.ClearCache()

	tmpl, err = env.LoadTemplate("greeting.html")
	if err != nil {
		t.Fatalf("LoadTemplate (cached) error: %v", err)
	}

	out, err = tmpl.ExecuteToString(map[string]interface{}{"name": "Gopher"})
	if err != nil {
		t.Fatalf("ExecuteToString after cache error: %v", err)
	}
	if got, want := out, "Hello Gopher"; got != want {
		t.Fatalf("unexpected cached render output: got %q want %q", got, want)
	}

	if count := loader.loadCount(); count != 1 {
		t.Fatalf("expected bytecode cache to prevent reload, got %d loads", count)
	}

	env.ClearCache()
	loader.setSource("Hi {{ name }}")

	tmpl, err = env.LoadTemplate("greeting.html")
	if err != nil {
		t.Fatalf("LoadTemplate after change error: %v", err)
	}

	out, err = tmpl.ExecuteToString(map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("ExecuteToString after change error: %v", err)
	}
	if got, want := out, "Hi Go"; got != want {
		t.Fatalf("unexpected reloaded output: got %q want %q", got, want)
	}

	if count := loader.loadCount(); count != 2 {
		t.Fatalf("expected loader to reload after modification, got %d loads", count)
	}
}
