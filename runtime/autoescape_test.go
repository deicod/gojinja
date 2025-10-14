package runtime

import "testing"

func TestEnvironmentAutoescapeDefaults(t *testing.T) {
	env := NewEnvironment()
	cases := []struct {
		name string
		file string
		want bool
	}{
		{"HTML", "index.html", true},
		{"Uppercase extension", "INDEX.HTML", true},
		{"HTM extension", "index.htm", true},
		{"XML extension", "feed.xml", true},
		{"Text file", "notes.txt", false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := env.shouldAutoescape(tt.file); got != tt.want {
				t.Fatalf("shouldAutoescape(%q) = %v, want %v", tt.file, got, tt.want)
			}
		})
	}
}

func TestEnvironmentAutoescapeBool(t *testing.T) {
	env := NewEnvironment()

	env.SetAutoescape(true)
	if !env.shouldAutoescape("index.txt") {
		t.Fatal("expected autoescape true")
	}

	env.SetAutoescape(false)
	if env.shouldAutoescape("index.html") {
		t.Fatal("expected autoescape false")
	}
}

func TestEnvironmentAutoescapeSlice(t *testing.T) {
	env := NewEnvironment()
	env.SetAutoescape([]string{".txt", "foo"})

	if !env.shouldAutoescape("readme.txt") {
		t.Fatalf("expected .txt extension to autoescape")
	}
	if !env.shouldAutoescape("about.foo") {
		t.Fatalf("expected .foo extension to autoescape")
	}
	if env.shouldAutoescape("index.html") {
		t.Fatalf("did not expect html extension to autoescape when not in list")
	}
}

func TestEnvironmentAutoescapeCallable(t *testing.T) {
	env := NewEnvironment()
	env.SetAutoescape(func(name string) bool {
		return len(name) > 0 && name[0] == 's'
	})

	if !env.shouldAutoescape("snippet.txt") {
		t.Fatalf("expected callable to enable autoescape for snippet.txt")
	}
	if env.shouldAutoescape("index.txt") {
		t.Fatalf("expected callable to disable autoescape for index.txt")
	}
}

func TestEnvironmentAutoescapeStringAliases(t *testing.T) {
	env := NewEnvironment()

	env.SetAutoescape("true")
	if !env.shouldAutoescape("index.txt") {
		t.Fatal("expected string alias 'true' to enable autoescape")
	}

	env.SetAutoescape("off")
	if env.shouldAutoescape("index.html") {
		t.Fatal("expected string alias 'off' to disable autoescape")
	}

	env.SetAutoescape("default")
	if !env.shouldAutoescape("index.html") || env.shouldAutoescape("notes.txt") {
		t.Fatal("expected default behaviour for string value 'default'")
	}
}
