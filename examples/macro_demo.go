package examples

import (
	"fmt"
	"log"

	"github.com/deicod/gojinja/runtime"
)

func RunMacroDemo() {
	fmt.Println("=== Jinja2 Macro System Demo ===")
	fmt.Println()

	// Demo 1: Basic Macro Usage
	basicMacroDemo()

	// Demo 2: Macro with Default Arguments
	defaultArgumentsDemo()

	// Demo 3: Macro Import System
	importSystemDemo()

	// Demo 4: Complex Macro Scenarios
	complexScenariosDemo()

	// Demo 5: Advanced Macro Features
	advancedFeaturesDemo()

	// Demo 6: Error Handling
	errorHandlingDemo()
}

func basicMacroDemo() {
	fmt.Println("1. Basic Macro Usage")
	fmt.Println("=====================")

	templateStr := `
{%- macro greet(name) -%}
Hello, {{ name }}!
{%- endmacro -%}

{{ greet("World") }}
{{ greet("Alice") }}
{{ greet("Bob") }}
`

	// Create environment and template
	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(templateStr, "basic_macro")
	if err != nil {
		log.Printf("Failed to create template: %v", err)
		return
	}

	// Execute template
	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	fmt.Println("Template:")
	fmt.Println(templateStr)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func defaultArgumentsDemo() {
	fmt.Println("2. Macro with Default Arguments")
	fmt.Println("=================================")

	templateStr := `
{%- macro input(name, type="text", value="", required=false) -%}
<input type="{{ type }}" name="{{ name }}" value="{{ value }}" {%- if required %} required{%- endif %}>
{%- endmacro -%}

{{ input("username") }}
{{ input("email", type="email") }}
{{ input("password", type="password", required=true) }}
{{ input("age", type="number", value="25") }}
`

	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(templateStr, "default_args")
	if err != nil {
		log.Printf("Failed to create template: %v", err)
		return
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	fmt.Println("Template:")
	fmt.Println(templateStr)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func importSystemDemo() {
	fmt.Println("3. Macro Import System")
	fmt.Println("======================")

	// Create macro library
	macrosTemplate := `
{%- macro button(text, style="primary") -%}
<button class="btn btn-{{ style }}">{{ text }}</button>
{%- endmacro -%}

{%- macro alert(message, type="info") -%}
<div class="alert alert-{{ type }}">{{ message }}</div>
{%- endmacro -%}

{%- macro card(title, content) -%}
<div class="card">
  <div class="card-header">{{ title }}</div>
  <div class="card-body">{{ content }}</div>
</div>
{%- endmacro -%}
`

	// Create main template that imports macros
	mainTemplate := `
{% from 'macros.html' import button, alert, card %}

{{ button("Click Me") }}
{{ button("Cancel", style="secondary") }}

{{ alert("Success! Your changes have been saved.", type="success") }}
{{ alert("Warning: This action cannot be undone.", type="warning") }}

{{ card("Welcome", "This is the card content.") }}
{{ card("About", "Learn more about our product.") }}
`

	// Create environment with loader
	loader := runtime.NewMapLoader(map[string]string{
		"macros.html": macrosTemplate,
		"main.html":   mainTemplate,
	})

	env := runtime.NewEnvironment()
	env.SetLoader(loader)

	template, err := env.LoadTemplate("main.html")
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		return
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	fmt.Println("Macros Template (macros.html):")
	fmt.Println(macrosTemplate)
	fmt.Println("Main Template (main.html):")
	fmt.Println(mainTemplate)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func complexScenariosDemo() {
	fmt.Println("4. Complex Macro Scenarios")
	fmt.Println("==========================")

	// Create a macro that generates form fields
	formMacros := `
{%- macro form_field(name, field_type="text", label="", value="", required=false, options=none) -%}
<div class="form-group">
  {%- if label -%}
  <label for="{{ name }}">{{ label }}</label>
  {%- endif -%}

  {%- if field_type == "select" -%}
  <select id="{{ name }}" name="{{ name }}" {%- if required %} required{%- endif %}>
    {%- for option in options -%}
    <option value="{{ option.value }}" {%- if option.value == value %} selected{%- endif %}>
      {{ option.label }}
    </option>
    {%- endfor -%}
  </select>
  {%- elif field_type == "textarea" -%}
  <textarea id="{{ name }}" name="{{ name }}" {%- if required %} required{%- endif %}>{{ value }}</textarea>
  {%- else -%}
  <input type="{{ field_type }}" id="{{ name }}" name="{{ name }}" value="{{ value }}" {%- if required %} required{%- endif %}>
  {%- endif -%}
</div>
{%- endmacro -%}

{%- macro user_form(user_data=none) -%}
<form method="post">
  {{ form_field("name", "text", "Full Name", user_data.name if user_data else "", required=true) }}
  {{ form_field("email", "email", "Email Address", user_data.email if user_data else "", required=true) }}
  {{ form_field("role", "select", "Role", user_data.role if user_data else "", required=true, [
    {"value": "user", "label": "Regular User"},
    {"value": "admin", "label": "Administrator"}
  ]) }}
  {{ form_field("bio", "textarea", "Biography", user_data.bio if user_data else "") }}
  {{ form_field("active", "checkbox", "Active", "yes", user_data.active if user_data else "no") }}

  <button type="submit">Save User</button>
</form>
{%- endmacro -%}
`

	// Main template using the form macros
	formTemplate := `
{% from 'form_macros.html' import user_form %}

<h1>Create New User</h1>
{{ user_form() }}

<h1>Edit Existing User</h1>
{{ user_form({"name": "John Doe", "email": "john@example.com", "role": "admin", "bio": "Software developer", "active": "yes"}) }}
`

	loader := runtime.NewMapLoader(map[string]string{
		"form_macros.html": formMacros,
		"form.html":        formTemplate,
	})

	env := runtime.NewEnvironment()
	env.SetLoader(loader)

	template, err := env.LoadTemplate("form.html")
	if err != nil {
		log.Printf("Failed to load template: %v", err)
		return
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	fmt.Println("Form Macros:")
	fmt.Println(formMacros)
	fmt.Println("Form Template:")
	fmt.Println(formTemplate)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func advancedFeaturesDemo() {
	fmt.Println("5. Advanced Macro Features")
	fmt.Println("==========================")

	// Demo macro with conditional logic and loops
	advancedTemplate := `
{%- macro table(data, headers=none, class="table") -%}
<table class="{{ class }}">
  {%- if headers -%}
  <thead>
    <tr>
      {%- for header in headers -%}
      <th>{{ header }}</th>
      {%- endfor -%}
    </tr>
  </thead>
  {%- endif -%}
  <tbody>
    {%- for row in data -%}
    <tr>
      {%- for cell in row -%}
      <td>{{ cell }}</td>
      {%- endfor -%}
    </tr>
    {%- endfor -%}
  </tbody>
</table>
{%- endmacro -%}

{%- macro navigation(items, active_item=none) -%}
<nav class="navigation">
  <ul>
    {%- for item in items -%}
    <li>
      {%- if item.name == active_item -%}
      <span class="active">{{ item.label }}</span>
      {%- else -%}
      <a href="{{ item.url }}">{{ item.label }}</a>
      {%- endif -%}
    </li>
    {%- endfor -%}
  </ul>
</nav>
{%- endmacro -%}

{%- macro breadcrumb(items) -%}
<nav class="breadcrumb">
  <ol>
    {%- for item in items -%}
    <li>
      {%- if loop.last -%}
      <span>{{ item.label }}</span>
      {%- else -%}
      <a href="{{ item.url }}">{{ item.label }}</a>
      {%- endif -%}
    </li>
    {%- endfor -%}
  </ol>
</nav>
{%- endmacro -%}

{# Table example #}
{{ table([
  ["John", "Doe", "john@example.com"],
  ["Jane", "Smith", "jane@example.com"],
  ["Bob", "Johnson", "bob@example.com"]
], ["First Name", "Last Name", "Email"], "table-striped") }}

{# Navigation example #}
{{ navigation([
  {"name": "home", "label": "Home", "url": "/"},
  {"name": "about", "label": "About", "url": "/about"},
  {"name": "contact", "label": "Contact", "url": "/contact"}
], "about") }}

{# Breadcrumb example #}
{{ breadcrumb([
  {"label": "Home", "url": "/"},
  {"label": "Products", "url": "/products"},
  {"label": "Electronics", "url": "/products/electronics"},
  {"label": "Smartphones"}
]) }}
`

	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(advancedTemplate, "advanced_features")
	if err != nil {
		log.Printf("Failed to create template: %v", err)
		return
	}

	result, err := template.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		return
	}

	fmt.Println("Advanced Features Template:")
	fmt.Println(advancedTemplate)
	fmt.Println("Output:")
	fmt.Println(result)
	fmt.Println()
}

func errorHandlingDemo() {
	fmt.Println("6. Error Handling")
	fmt.Println("=================")

	// Template with various macro errors
	errorTemplate := `
{%- macro broken_macro(required_arg) -%}
This macro requires an argument: {{ required_arg }}
{%- endmacro -%}

{%- macro macro_with_bad_default(x, default="unclosed) -%}
{{ x }}: {{ default }}
{%- endmacro -%}

{# These will cause errors #}
{{ broken_macro() }}
{{ macro_with_bad_default("test") }}
{{ non_existent_macro("arg") }}
`

	env := runtime.NewEnvironment()
	template, err := env.NewTemplateFromSource(errorTemplate, "error_test")
	if err != nil {
		log.Printf("Failed to create template: %v", err)
		return
	}

	fmt.Println("Template with potential errors:")
	fmt.Println(errorTemplate)
	fmt.Println("Attempting execution...")

	result, err := template.ExecuteToString(nil)
	if err != nil {
		fmt.Printf("Expected error occurred: %v\n", err)

		// Check error type
		if runtime.IsMacroError(err) {
			fmt.Println("Error type: MacroError")
		}
	} else {
		fmt.Println("Result (unexpected):")
		fmt.Println(result)
	}

	fmt.Println()

	// Demonstrate proper error handling
	fmt.Println("Proper error handling example:")
	properTemplate := `
{%- macro safe_macro(name="World") -%}
Hello, {{ name }}!
{%- endmacro -%}

{# Safe usage with default values #}
{{ safe_macro() }}
{{ safe_macro("Alice") }}
`

	template2, err := env.NewTemplateFromSource(properTemplate, "safe_template")
	if err != nil {
		log.Printf("Failed to create safe template: %v", err)
		return
	}

	result2, err := template2.ExecuteToString(nil)
	if err != nil {
		log.Printf("Failed to execute safe template: %v", err)
		return
	}

	fmt.Println("Safe template output:")
	fmt.Println(result2)
	fmt.Println()
}

// Helper function to show macro registry statistics
func showMacroStats(env *runtime.Environment) {
	stats := env.GetMacroStats()
	fmt.Printf("Macro Registry Stats:\n")
	fmt.Printf("  Global macros: %d\n", stats["globals"])
	fmt.Printf("  Template macros: %d\n", stats["template_macros"])
	fmt.Printf("  Namespace macros: %d\n", stats["namespace_macros"])
	fmt.Printf("  Namespaces: %d\n", stats["namespaces"])
	fmt.Println()
}