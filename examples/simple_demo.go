package examples

import (
	"fmt"
	"log"
	"strings"

	"github.com/deicod/gojinja/runtime"
)

func RunSimpleDemo() {
	// Simple demonstration of the Go Jinja2 runtime engine

	// Example 1: Basic variable substitution
	fmt.Println("=== Example 1: Basic Variable Substitution ===")
	template1 := "Hello, {{ name }}!"
	context1 := map[string]interface{}{
		"name": "World",
	}

	result1, err := runtime.ExecuteToString(template1, context1)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Template: %s\n", template1)
	fmt.Printf("Result:   %s\n\n", result1)

	// Example 2: Simple filter usage
	fmt.Println("=== Example 2: Filter Usage ===")
	template2 := "Hello, {{ name|upper }}!"

	result2, err := runtime.ExecuteToString(template2, context1)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Template: %s\n", template2)
	fmt.Printf("Result:   %s\n\n", result2)

	// Example 3: For loop
	fmt.Println("=== Example 3: For Loop ===")
	template3 := "Items: {% for item in items %}{{ item }} {% endfor %}"
	context3 := map[string]interface{}{
		"items": []interface{}{"apple", "banana", "cherry"},
	}

	result3, err := runtime.ExecuteToString(template3, context3)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	fmt.Printf("Template: %s\n", template3)
	fmt.Printf("Result:   %s\n\n", result3)

	// Example 4: Conditional
	fmt.Println("=== Example 4: Conditional ===")
	template4 := "{% if user %}Welcome, {{ user }}!{% else %}Hello, Guest!{% endif %}"
	context4a := map[string]interface{}{"user": "Alice"}
	context4b := map[string]interface{}{}

	result4a, err := runtime.ExecuteToString(template4, context4a)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	result4b, err := runtime.ExecuteToString(template4, context4b)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Template: %s\n", template4)
	fmt.Printf("Result (with user): %s\n", result4a)
	fmt.Printf("Result (without):  %s\n\n", result4b)

	// Example 5: Using the Environment API
	fmt.Println("=== Example 5: Environment API ===")
	env := runtime.NewEnvironment()

	// Add a custom filter
	env.AddFilter("shout", func(ctx *runtime.Context, value interface{}, args ...interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return strings.ToUpper(str) + "!!!", nil
		}
		return value, nil
	})

	template5 := "Hello, {{ name|shout }}"

	tmpl, err := env.ParseString(template5, "greeting")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}

	result5, err := tmpl.ExecuteToString(context1)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}

	fmt.Printf("Template: %s\n", template5)
	fmt.Printf("Result:   %s\n\n", result5)

	fmt.Println("✅ Go Jinja2 Runtime Engine is working!")
}