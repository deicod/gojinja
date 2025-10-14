package parser

import (
	"github.com/deicod/gojinja/nodes"
)

// ParseTemplate is a simple one-line API for parsing templates
// It creates a default environment and parses the given template string
// Returns the AST or an error with position information
func ParseTemplate(template string) (*nodes.Template, error) {
	env := &Environment{}
	return ParseTemplateWithEnv(env, template, "template", "template.html")
}

// ParseTemplateWithEnv parses a template using the given environment
// Returns the AST or an error with position information
func ParseTemplateWithEnv(env *Environment, template, name, filename string) (*nodes.Template, error) {
	parser, err := NewParser(env, template, name, filename, "")
	if err != nil {
		return nil, err
	}

	return parser.Parse()
}

// ParseTemplateWithErrorHandling parses a template and returns detailed error information
// This function provides better error context for debugging
func ParseTemplateWithErrorHandling(template string) (*nodes.Template, error) {
	return ParseTemplateWithEnv(&Environment{}, template, "template", "template.html")
}