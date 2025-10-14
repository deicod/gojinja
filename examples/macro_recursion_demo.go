package examples

import (
	"fmt"
	"log"

	"github.com/deicod/gojinja/runtime"
)

func RunMacroRecursionDemo() {
	fmt.Println("=== Recursive Macro Demo ===")
	fmt.Println()

	// Demonstrate recursive macros (like tree rendering)
	recursiveMacroDemo()

	// Demonstrate mutual recursion
	mutualRecursionDemo()
}

func recursiveMacroDemo() {
	fmt.Println("1. Recursive Tree Rendering Macro")
	fmt.Println("=================================")

	// Template with recursive macro for rendering a tree structure
	treeTemplate := `
{%- macro render_tree(node, level=0) -%}
{%- if node -%}
{%- for _ in range(level) %}  {% endfor %}{{ node.name }}
{%- if node.children %} ({{ node.children|length }})
{%- for child in node.children %}
{{ render_tree(child, level + 1) }}
{%- endfor -%}
{%- endif -%}
{%- endif -%}
{%- endmacro -%}

{# Tree data structure #}
{% set tree_data = {
    "name": "root",
    "value": 100,
    "children": [
        {
            "name": "src",
            "children": [
                {"name": "main.go"},
                {"name": "utils.go"},
                {
                    "name": "models",
                    "children": [
                        {"name": "user.go"},
                        {"name": "product.go"}
                    ]
                }
            ]
        },
        {
            "name": "templates",
            "children": [
                {"name": "index.html"},
                {"name": "about.html"}
            ]
        },
        {"name": "README.md"}
    ]
} %}

Tree Structure:
{{ render_tree(tree_data) }}

{%- macro render_menu(items, level=0) -%}
{%- for item in items -%}
{%- for _ in range(level) %}  {% endfor %}- {{ item.title }}
{%- if item.children %}
{{ render_menu(item.children, level + 1) }}
{%- endif -%}
{%- endfor -%}
{%- endmacro -%}

Menu Structure:
{{ render_menu([
    {"title": "Home", "url": "/"},
    {"title": "Products", "children": [
        {"title": "Electronics", "children": [
            {"title": "Phones"},
            {"title": "Laptops"}
        ]},
        {"title": "Books", "children": [
            {"title": "Fiction"},
            {"title": "Non-Fiction"}
        ]}
    ]},
    {"title": "About", "url": "/about"},
    {"title": "Contact", "url": "/contact"}
]) }}
`

	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(treeTemplate, "recursive")
	if err != nil {
		log.Fatalf("Failed to create template: %v", err)
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("Recursive Template:")
	fmt.Println(treeTemplate)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func mutualRecursionDemo() {
	fmt.Println("2. Mutual Recursion Demo")
	fmt.Println("=========================")

	// Template demonstrating mutual recursion between two macros
	mutualTemplate := `
{%- macro is_even(n) -%}
{%- if n == 0 %}true
{%- else %}{{ is_odd(n - 1) }}
{%- endif -%}
{%- endmacro -%}

{%- macro is_odd(n) -%}
{%- if n == 0 %}false
{%- else %}{{ is_even(n - 1) }}
{%- endif -%}
{%- endmacro -%}

Even/Odd Results:
- is_even(0): {{ is_even(0) }}
- is_odd(0): {{ is_odd(0) }}
- is_even(1): {{ is_even(1) }}
- is_odd(1): {{ is_odd(1) }}
- is_even(2): {{ is_even(2) }}
- is_odd(2): {{ is_odd(2) }}
- is_even(3): {{ is_even(3) }}
- is_odd(3): {{ is_odd(3) }}

{%- macro fibonacci(n) -%}
{%- if n <= 1 %}{{ n }}
{%- else %}{{ fibonacci(n - 1) + fibonacci(n - 2) }}
{%- endif -%}
{%- endmacro -%}

Fibonacci Sequence (first 10):
{% for i in range(10) %}
F({{ i }}) = {{ fibonacci(i) }}
{% endfor %}

{%- macro factorial(n) -%}
{%- if n <= 1 %}1
{%- else %}{{ n }} * {{ factorial(n - 1) }}
{%- endif -%}
{%- endmacro -%}

Factorials:
- 0! = {{ factorial(0) }}
- 1! = {{ factorial(1) }}
- 3! = {{ factorial(3) }}
- 5! = {{ factorial(5) }}

{%- macro render_directory(path, depth=0) -%}
{%- set items = path.items if path.items else [] -%}
{%- for _ in range(depth) %}  {% endfor %}{{ path.name }}/
{%- for item in items %}
{%- if item.type == "directory" %}
{{ render_directory(item, depth + 1) }}
{%- else %}
{%- for _ in range(depth + 1) %}  {% endfor %}{{ item.name }}
{%- endif -%}
{%- endfor -%}
{%- endmacro -%}

File System Structure:
{{ render_directory({
    "name": "project",
    "type": "directory",
    "items": [
        {"name": "README.md", "type": "file"},
        {"name": "src", "type": "directory", "items": [
            {"name": "main.go", "type": "file"},
            {"name": "utils.go", "type": "file"},
            {"name": "config", "type": "directory", "items": [
                {"name": "database.yaml", "type": "file"},
                {"name": "server.yaml", "type": "file"}
            ]}
        ]},
        {"name": "tests", "type": "directory", "items": [
            {"name": "test_main.go", "type": "file"}
        ]}
    ]
}) }}
`

	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(mutualTemplate, "mutual_recursion")
	if err != nil {
		log.Fatalf("Failed to create template: %v", err)
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("Mutual Recursion Template:")
	fmt.Println(mutualTemplate)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()

	// Test recursion depth limits
	fmt.Println("3. Recursion Depth Test")
	fmt.Println("=========================")

	depthTestTemplate := `
{%- macro deep_recursion(n, max_depth=10) -%}
{%- if n > 0 %}{{ deep_recursion(n - 1, max_depth) }}{% endif %}
Depth: {{ n }}
{%- endmacro -%}

Testing recursion depth (should limit at 10):
{{ deep_recursion(15, 10) }}
`

	template2, err := env.NewTemplateFromSource(depthTestTemplate, "depth_test")
	if err != nil {
		log.Printf("Failed to create depth test template: %v", err)
		return
	}

	result2, err := template2.ExecuteToString(nil)
	if err != nil {
		fmt.Printf("Depth test failed (expected): %v\n", err)
	} else {
		fmt.Println("Depth test result:")
		fmt.Println(result2)
	}
}