package examples

import (
	"fmt"
	"log"

	"github.com/deicod/gojinja/parser"
	"github.com/deicod/gojinja/nodes"
)

func RunParserDemo() {
	// Create a simple environment (we'll implement this fully later)
	env := &parser.Environment{}

	// Test templates
	templates := []string{
		// Simple variable
		"Hello {{ name }}!",

		// Basic control structures
		"{% if user %}Welcome {{ user.name }}!{% else %}Please log in{% endif %}",

		// For loop
		"{% for item in items %}{{ item }}{% if not loop.last %}, {% endif %}{% endfor %}",

		// Filters and expressions
		"{{ title | upper | truncate(20) }}",

		// Complex expressions
		"{{ (user.age + 1) * factor if user else 'Unknown' }}",

		// List and dict literals
		"{{ [1, 2, 3] + {'key': 'value'}.keys() | list }}",

		// Method calls and attribute access
		"{{ users.filter(active=true).map('name') | join(', ') }}",

		// Set statements
		"{% set navigation = ['Home', 'About', 'Contact'] %}",

		// Macro definition
		"{% macro greeting(name, greeting='Hello') %}{{ greeting }}, {{ name }}!{% endmacro %}",

		// Include and extends
		"{% extends 'base.html' %}\n{% block content %}<p>Page content</p>{% endblock %}",

		// With statement
		"{% with alpha = 1, beta = 2 %}{{ alpha + beta }}{% endwith %}",
	}

	fmt.Println("GoJinja2 Parser Demo")
	fmt.Println("===================")

	for i, tmpl := range templates {
		fmt.Printf("\nTemplate %d:\n", i+1)
		fmt.Printf("Source: %q\n", tmpl)

		// Create parser
		parser, err := parser.NewParser(env, tmpl, fmt.Sprintf("test%d", i+1), "demo.html", "")
		if err != nil {
			log.Printf("Failed to create parser: %v", err)
			continue
		}

		// Parse template
		ast, err := parser.Parse()
		if err != nil {
			log.Printf("Failed to parse template: %v", err)
			continue
		}

		// Print AST structure
		fmt.Printf("AST: %s\n", nodes.Dump(ast))
	}

	fmt.Println("\nDemo completed!")
}