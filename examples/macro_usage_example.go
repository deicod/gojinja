package examples

import (
	"fmt"
	"log"

	"github.com/deicod/gojinja/runtime"
)

func RunMacroUsageExample() {
	fmt.Println("=== Jinja2 Macro Usage Example ===")
	fmt.Println()

	// This example demonstrates the exact usage pattern requested
	demonstrateMacroUsage()
}

func demonstrateMacroUsage() {
	fmt.Println("Creating templates as specified in the requirements...")
	fmt.Println()

	// Create the macros.html template
	macrosHTML := `{% macro input(name, type="text", value="") %}
    <input type="{{ type }}" name="{{ name }}" value="{{ value }}">
{% endmacro %}

{% macro user_list(users) %}
    <ul>
    {% for user in users %}
        <li>{{ user.name|escape }}</li>
    {% endfor %}
    </ul>
{% endmacro %}`

	// Create the template.html that imports from macros.html
	templateHTML := `{% from 'macros.html' import input, user_list %}

<form>
    {{ input("username", value=user.name) }}
    {{ input("email", type="email") }}
</form>

{{ user_list(users) }}`

	// Set up the environment with a map loader
	loader := runtime.NewMapLoader(map[string]string{
		"macros.html":  macrosHTML,
		"template.html": templateHTML,
	})

	env := runtime.NewEnvironment()
	env.SetLoader(loader)

	// Load and render the template
	template, err := env.LoadTemplate("template.html")
	if err != nil {
		log.Fatalf("Failed to load template: %v", err)
	}

	// Prepare context data
	context := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "john_doe",
		},
		"users": []map[string]interface{}{
			{"name": "Alice"},
			{"name": "Bob"},
			{"name": "Charlie"},
		},
	}

	// Execute the template
	result, err := template.ExecuteToString(context)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	fmt.Println("=== Template: macros.html ===")
	fmt.Println(macrosHTML)
	fmt.Println()

	fmt.Println("=== Template: template.html ===")
	fmt.Println(templateHTML)
	fmt.Println()

	fmt.Println("=== Context Data ===")
	fmt.Printf("user: %+v\n", context["user"])
	fmt.Printf("users: %+v\n", context["users"])
	fmt.Println()

	fmt.Println("=== Rendered Output ===")
	fmt.Println(result)
	fmt.Println()

	// Demonstrate advanced usage
	fmt.Println("=== Advanced Macro Features ===")
	demonstrateAdvancedFeatures()
}

func demonstrateAdvancedFeatures() {
	// More complex macro usage
	advancedMacros := `
{%- macro card(title, content, footer="") -%}
<div class="card">
    <div class="card-header">
        <h3>{{ title }}</h3>
    </div>
    <div class="card-body">
        {{ content }}
    </div>
    {%- if footer -%}
    <div class="card-footer">
        {{ footer }}
    </div>
    {%- endif -%}
</div>
{%- endmacro -%}

{%- macro alert(message, type="info") -%}
<div class="alert alert-{{ type }}">
    {{ message }}
</div>
{%- endmacro -%}

{%- macro nav_item(url, label, active=false) -%}
<li class="nav-item{% if active %} active{% endif %}">
    <a class="nav-link" href="{{ url }}">{{ label }}</a>
</li>
{%- endmacro -%}

{%- macro nav(items, active_item="") -%}
<nav class="navbar">
    <ul class="nav">
        {%- for item in items -%}
        {{ nav_item(item.url, item.label, item.name == active_item) }}
        {%- endfor -%}
    </ul>
</nav>
{%- endmacro -%}
`

	advancedTemplate := `
{% from 'advanced_macros.html' import card, alert, nav %}

{{ nav([
    {"name": "home", "url": "/", "label": "Home"},
    {"name": "about", "url": "/about", "label": "About"},
    {"name": "contact", "url": "/contact", "label": "Contact"}
], "about") }}

{{ alert("Welcome to our website!", "success") }}

{{ card("About Us", "We are a company that makes awesome products.", "© 2023 Our Company") }}

{{ card("Features",
    card("Fast", "Lightning fast performance", "footer") ~
    card("Secure", "Enterprise-grade security", "footer") ~
    card("Scalable", "Grows with your needs", "footer")
) }}
`

	loader := runtime.NewMapLoader(map[string]string{
		"advanced_macros.html": advancedMacros,
		"advanced.html":        advancedTemplate,
	})

	env := runtime.NewEnvironment()
	env.SetLoader(loader)

	template, err := env.LoadTemplate("advanced.html")
	if err != nil {
		log.Printf("Failed to load advanced template: %v", err)
		return
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute advanced template: %v", err)
		return
	}

	fmt.Println("=== Advanced Macros Template ===")
	fmt.Println(advancedTemplate)
	fmt.Println()

	fmt.Println("=== Advanced Rendered Output ===")
	fmt.Println(result)
	fmt.Println()

	// Show macro statistics
	stats := env.GetMacroStats()
	fmt.Printf("Macro Statistics: %+v\n", stats)
}