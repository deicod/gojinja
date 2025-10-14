package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFileSystemLoaderSearchPathFallback(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	expected := "hello from second path"
	if err := os.WriteFile(filepath.Join(dir2, "greeting.txt"), []byte(expected), 0o644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	loader := NewFileSystemLoader(dir1)
	loader.AddSearchPath(dir2)

	content, err := loader.Load("greeting.txt")
	if err != nil {
		t.Fatalf("expected to load template, got error: %v", err)
	}
	if content != expected {
		t.Fatalf("expected content %q, got %q", expected, content)
	}

	paths := loader.SearchPath()
	if len(paths) != 2 || paths[0] != dir1 || paths[1] != dir2 {
		t.Fatalf("unexpected search path order: %v", paths)
	}

	// Ensure SearchPath returns a copy that can be mutated by the caller.
	paths[0] = "mutated"
	if loader.SearchPath()[0] != dir1 {
		t.Fatal("SearchPath should return a defensive copy")
	}
}

func TestFileSystemLoaderTemplateNotFoundTracksAllPaths(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	loader := NewFileSystemLoader(dir1, dir2)

	_, err := loader.Load("missing.txt")
	if err == nil {
		t.Fatal("expected error for missing template")
	}

	var notFound *TemplateNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected TemplateNotFoundError, got %T", err)
	}

	expectedTried := []string{
		filepath.Join(dir1, "missing.txt"),
		filepath.Join(dir2, "missing.txt"),
	}

	if len(notFound.Tried) != len(expectedTried) {
		t.Fatalf("expected tried paths %v, got %v", expectedTried, notFound.Tried)
	}

	for i, path := range expectedTried {
		if notFound.Tried[i] != path {
			t.Fatalf("expected tried path %q at index %d, got %q", path, i, notFound.Tried[i])
		}
	}
}

func TestFileSystemLoaderSetSearchPath(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	loader := NewFileSystemLoader()
	loader.SetSearchPath(dir1, dir2)

	paths := loader.SearchPath()
	if len(paths) != 2 || paths[0] != dir1 || paths[1] != dir2 {
		t.Fatalf("unexpected search paths after SetSearchPath: %v", paths)
	}

	loader.SetSearchPath("", dir2)
	paths = loader.SearchPath()
	if len(paths) != 1 || paths[0] != dir2 {
		t.Fatalf("expected empty paths to be ignored, got %v", paths)
	}
}
