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

func TestEnvironmentAutoescapeGenericSlice(t *testing.T) {
	env := NewEnvironment()

	env.SetAutoescape([]interface{}{" html ", "XML", ".JINJA"})

	if !env.shouldAutoescape("page.HTML") {
		t.Fatalf("expected autoescape for html extension from generic slice")
	}
	if !env.shouldAutoescape("feed.xml") {
		t.Fatalf("expected autoescape for xml extension from generic slice")
	}
	if !env.shouldAutoescape("snippet.jinja") {
		t.Fatalf("expected autoescape for jinja extension from generic slice")
	}
	if env.shouldAutoescape("notes.txt") {
		t.Fatalf("did not expect txt extension to autoescape")
	}
}

func TestEnvironmentAutoescapeNilResetsToDefault(t *testing.T) {
	env := NewEnvironment()

	env.SetAutoescape(false)
	if env.shouldAutoescape("index.html") {
		t.Fatalf("expected html to be false after forcing autoescape off")
	}

	env.SetAutoescape(nil)
	if !env.shouldAutoescape("index.html") {
		t.Fatalf("expected html to autoescape after resetting to default")
	}
	if env.shouldAutoescape("plain.txt") {
		t.Fatalf("expected txt to remain disabled after resetting to default")
	}
}

func TestSelectAutoescape(t *testing.T) {
	selector := SelectAutoescape([]string{"html", ".xml"}, []string{"txt"}, true, false)

	if !selector("index.HTML") {
		t.Fatalf("expected selector to enable autoescape for html files")
	}
	if !selector("feed.xml") {
		t.Fatalf("expected selector to enable autoescape for xml files")
	}
	if selector("notes.txt") {
		t.Fatalf("expected selector to disable autoescape for txt files")
	}
	if selector("snippet.jinja") {
		t.Fatalf("expected selector to return default false for unmatched extensions")
	}
	if !selector("") {
		t.Fatalf("expected selector to use defaultForString when no name is provided")
	}
}

func TestSelectAutoescapeIntegration(t *testing.T) {
	env := NewEnvironment()
	env.SetAutoescape(SelectAutoescape([]string{"html"}, []string{"txt"}, false, false))

	if !env.shouldAutoescape("index.html") {
		t.Fatalf("expected environment to autoescape html via selector")
	}
	if env.shouldAutoescape("notes.txt") {
		t.Fatalf("expected environment selector to disable txt autoescape")
	}
	if env.shouldAutoescape("about.md") {
		t.Fatalf("expected environment selector to use default false for unmatched extensions")
	}
}
