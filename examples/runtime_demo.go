package examples

import (
	"fmt"
	"log"
	"strings"

	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/parser"
	"github.com/deicod/gojinja/runtime"
)

func RunRuntimeDemo() {
	// Example 1: Basic template rendering
	fmt.Println("=== Example 1: Basic Template Rendering ===")
	basicExample()

	// Example 2: Complex template with loops and conditionals
	fmt.Println("\n=== Example 2: Complex Template ===")
	complexExample()

	// Example 3: Custom filters and functions
	fmt.Println("\n=== Example 3: Custom Filters and Functions ===")
	customFiltersExample()

	// Example 4: Environment configuration
	fmt.Println("\n=== Example 4: Environment Configuration ===")
	environmentExample()

	// Example 5: Error handling
	fmt.Println("\n=== Example 5: Error Handling ===")
	errorHandlingExample()

	// Example 6: Batch rendering
	fmt.Println("\n=== Example 6: Batch Rendering ===")
	batchRenderingExample()

	// Example 7: Template from AST
	fmt.Println("\n=== Example 7: Template from AST ===")
	astExample()

	// Example 8: Autoescaping
	fmt.Println("\n=== Example 8: Autoescaping ===")
	autoescapeExample()
}

func basicExample() {
	// Simple template rendering
	template := "Hello {{ name }}! You have {{ count }} new messages."
	context := map[string]interface{}{
		"name":  "Alice",
		"count": 5,
	}

	result, err := runtime.ExecuteToString(template, context)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Printf("Template: %s\n", template)
	fmt.Printf("Result:   %s\n", result)
}

func complexExample() {
	template := `
<!DOCTYPE html>
<html>
<head>
    <title>{{ title|title }}</title>
</head>
<body>
    <h1>{{ heading|upper }}</h1>

    {% if user %}
    <p>Welcome, {{ user.name|capitalize }}!</p>
    {% endif %}

    {% if items %}
    <h2>Your Items ({{ items|length }} total):</h2>
    <ul>
    {% for item in items %}
        <li>
            {{ loop.index }}. {{ item.title|title }}
            {% if item.description %}
            - {{ item.description|truncate(50) }}
            {% endif %}
            ({{ item.tags|length }} tags)
        </li>
    {% endfor %}
    </ul>

    {% if items|length > 5 %}
    <p>Showing all items. Use pagination for better performance.</p>
    {% endif %}
    {% else %}
    <p>No items found.</p>
    {% endif %}

    <hr>
    <p>Generated with Go Jinja2 Runtime Engine</p>
</body>
</html>
`

	context := map[string]interface{}{
		"title": "welcome to your dashboard",
		"heading": "user dashboard",
		"user": map[string]interface{}{
			"name": "john doe",
		},
		"items": []interface{}{
			map[string]interface{}{
				"title":       "first item",
				"description": "This is the description of the first item which is quite long and will be truncated.",
				"tags":        []interface{}{"important", "work"},
			},
			map[string]interface{}{
				"title":       "second item",
				"description": "Short description",
				"tags":        []interface{}{"personal"},
			},
			map[string]interface{}{
				"title": "third item",
				"tags":  []interface{}{"urgent", "work", "review"},
			},
		},
	}

	result, err := runtime.ExecuteToString(template, context)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Println("Complex template rendered successfully!")
	fmt.Printf("Result length: %d characters\n", len(result))
	fmt.Println("First 200 characters:")
	if len(result) > 200 {
		fmt.Println(result[:200] + "...")
	} else {
		fmt.Println(result)
	}
}

func customFiltersExample() {
	// Create a custom environment
	env := runtime.NewEnvironment()

	// Add a custom filter
	env.AddFilter("slugify", func(ctx *runtime.Context, value interface{}, args ...interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			// Simple slugify implementation
			slug := strings.ToLower(str)
			slug = strings.ReplaceAll(slug, " ", "-")
			slug = strings.ReplaceAll(slug, "_", "-")
			// Remove non-alphanumeric characters except hyphens
			var result strings.Builder
			for _, r := range slug {
				if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
					result.WriteRune(r)
				}
			}
			return result.String(), nil
		}
		return value, nil
	})

	// Add a custom global function
	env.AddGlobal("format_price", func(ctx *runtime.Context, args ...interface{}) (interface{}, error) {
		if len(args) == 0 {
			return "$0.00", nil
		}
		if price, ok := args[0].(float64); ok {
			return fmt.Sprintf("$%.2f", price), nil
		}
		return "$0.00", nil
	})

	// Use the custom filter and function
	template := `
Product: {{ name|slugify }}
Price: {{ format_price(price) }}
Tags: {{ tags|join(", ")|upper }}
`

	context := map[string]interface{}{
		"name":  "Awesome Product",
		"price": 29.99,
		"tags":  []interface{}{"new", "sale", "popular"},
	}

	result, err := env.ParseString(template, "product")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	output, err := result.ExecuteToString(context)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Println("Template with custom filters:")
	fmt.Println(output)
}

func environmentExample() {
	// Create environment with custom configuration
	env := runtime.NewEnvironment()
	env.SetAutoescape(true)
	env.SetTrimBlocks(true)
	env.SetLstripBlocks(true)

	// Configure template loader (in-memory)
	templates := map[string]string{
		"base.html": `
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    <header>
        <h1>{{ site_name }}</h1>
    </header>
    <main>
        {% block content %}{% endblock %}
    </main>
    <footer>
        <p>&copy; 2024 {{ site_name }}</p>
    </footer>
</body>
</html>
`,
		"home.html": `
{% extends "base.html" %}

{% block title %}Home - {{ site_name }}{% endblock %}

{% block content %}
<h2>Welcome to {{ site_name }}!</h2>
<p>This is the home page content.</p>

{% if featured_items %}
<h3>Featured Items:</h3>
<ul>
{% for item in featured_items %}
    <li>{{ item.title }} - {{ item.price }}</li>
{% endfor %}
</ul>
{% endif %}
{% endblock %}
`,
	}

	loader := runtime.NewMapLoader(templates)
	env.SetLoader(loader)

	// Add site-wide variables
	env.AddGlobal("site_name", "My Awesome Website")

	// Add a custom date filter
	env.AddFilter("date", func(ctx *runtime.Context, value interface{}, args ...interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return str, nil // Simplified - would use proper date formatting
		}
		return "Unknown date", nil
	})

	// Try to load and render a template
	fmt.Println("Environment configuration:")

	// Check autoescaping by creating a simple template
	testTemplate, err := env.NewTemplateFromSource("test", "home.html")
	if err != nil {
		log.Printf("Error creating test template: %v", err)
	} else {
		fmt.Printf("Autoescape: %v\n", testTemplate.Autoescape())
	}
	fmt.Printf("Available templates: %v\n", []string{"base.html", "home.html"})

	// Note: Template inheritance would be implemented here
	// For now, just render a simple template
	simpleTemplate := `
<h1>{{ site_name }}</h1>
<p>Rendered at: {{ '2024-01-01'|date }}</p>
`

	result, err := env.ParseString(simpleTemplate, "simple")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	output, err := result.ExecuteToString(nil)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Println("Rendered template:")
	fmt.Println(output)
}

func errorHandlingExample() {
	examples := []struct {
		name     string
		template string
		context  map[string]interface{}
	}{
		{
			name:     "Undefined variable",
			template: "Hello {{ undefined_var }}!",
			context:  map[string]interface{}{},
		},
		{
			name:     "Division by zero",
			template: "{{ 10 / 0 }}",
			context:  nil,
		},
		{
			name:     "Unknown filter",
			template: "{{ 'hello'|unknown_filter }}",
			context:  nil,
		},
		{
			name:     "Invalid syntax",
			template: "{% if %}",
			context:  nil,
		},
		{
			name:     "Index out of bounds",
			template: "{{ [1, 2, 3][10] }}",
			context:  nil,
		},
	}

	for _, example := range examples {
		fmt.Printf("\n%s:\n", example.name)
		fmt.Printf("Template: %s\n", example.template)

		result, err := runtime.ExecuteToString(example.template, example.context)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Result: %s\n", result)
		}
	}
}

func batchRenderingExample() {
	// Create batch renderer
	renderer := runtime.NewBatchRenderer(runtime.NewEnvironment())

	// Add multiple templates
	templates := map[string]string{
		"welcome": "Welcome, {{ name }}!",
		"goodbye": "Goodbye, {{ name }}!",
		"counter": "Count: {{ count }}",
		"list":    "Items: {{ items|join(', ') }}",
	}

	for name, tmpl := range templates {
		err := renderer.AddTemplate(name, tmpl)
		if err != nil {
			log.Printf("Error adding template %s: %v", name, err)
			continue
		}
	}

	fmt.Printf("Added %d templates to batch renderer\n", renderer.Size())

	// Render all templates with different contexts
	contexts := []map[string]interface{}{
		{"name": "Alice", "count": 1, "items": []interface{}{"a", "b"}},
		{"name": "Bob", "count": 2, "items": []interface{}{"x", "y", "z"}},
		{"name": "Charlie", "count": 3, "items": []interface{}{}},
	}

	for i, ctx := range contexts {
		fmt.Printf("\nContext %d:\n", i+1)
		for name := range templates {
			result, err := renderer.Render(name, ctx)
			if err != nil {
				fmt.Printf("  %s: Error - %v\n", name, err)
			} else {
				fmt.Printf("  %s: %s\n", name, result)
			}
		}
	}
}

func astExample() {
	// Parse template to AST
	templateStr := "Hello {{ name|upper }}! You have {{ count }} {{ 'message'|pluralize(count) }}."

	fmt.Printf("Original template: %s\n", templateStr)

	// Parse using the parser
	ast, err := parser.ParseTemplate(templateStr)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	fmt.Printf("AST type: %T\n", ast)
	fmt.Printf("AST dump:\n%s\n", nodes.Dump(ast))

	// Create template from AST
	template, err := runtime.ParseASTWithName(ast, "greeting")
	if err != nil {
		log.Printf("Error creating template from AST: %v", err)
		return
	}

	// Render the template
	context := map[string]interface{}{
		"name":  "world",
		"count": 5,
	}

	result, err := template.ExecuteToString(context)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Printf("Rendered result: %s\n", result)
}

func autoescapeExample() {
	// Test with autoescaping enabled and disabled
	input := `<script>alert('XSS')</script>`

	// With autoescaping enabled (default for HTML files)
	env1 := runtime.NewEnvironment()
	env1.SetAutoescape(true)

	template1 := "Safe output: {{ content }}"
	tmpl1, err := env1.ParseString(template1, "test.html")
	if err != nil {
		log.Printf("Error creating template: %v", err)
		return
	}

	result1, err := tmpl1.ExecuteToString(map[string]interface{}{"content": input})
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	// With autoescaping disabled
	env2 := runtime.NewEnvironment()
	env2.SetAutoescape(false)

	template2 := "Raw output: {{ content }}"
	tmpl2, err := env2.ParseString(template2, "test.txt")
	if err != nil {
		log.Printf("Error creating template: %v", err)
		return
	}

	result2, err := tmpl2.ExecuteToString(map[string]interface{}{"content": input})
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Printf("Input: %s\n", input)
	fmt.Printf("With autoescaping: %s\n", result1)
	fmt.Printf("Without autoescaping: %s\n", result2)

	// Test with safe filter
	template3 := "Safe filter: {{ content|safe }}"
	tmpl3, err := env1.ParseString(template3, "test.html")
	if err != nil {
		log.Printf("Error creating template: %v", err)
		return
	}

	result3, err := tmpl3.ExecuteToString(map[string]interface{}{"content": input})
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		return
	}

	fmt.Printf("With safe filter: %s\n", result3)
}