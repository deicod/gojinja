// Package gojinja2 is a Go implementation of the Jinja2 template engine
package gojinja2

import (
	"path/filepath"

	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/runtime"
)

// Version of the gojinja2 library
const Version = "0.1.0"

// Template represents a compiled Jinja2 template
type Template = runtime.Template

// Environment represents the Jinja2 environment
type Environment = runtime.Environment

// Context represents the template rendering context
type Context = runtime.Context

// Loader represents a template loader
type Loader = runtime.Loader

// FileSystemLoader loads templates from the file system
type FileSystemLoader = runtime.FileSystemLoader

// MapLoader loads templates from an in-memory map
type MapLoader = runtime.MapLoader

// SecurityManager manages security policies
type SecurityManager = runtime.SecurityManager

// SecurityContext manages security during template execution
type SecurityContext = runtime.SecurityContext

// NewEnvironment creates a new Jinja2 environment
func NewEnvironment() *Environment {
	return runtime.NewEnvironment()
}

// NewFileSystemLoader creates a new filesystem loader
func NewFileSystemLoader(basePaths ...string) *FileSystemLoader {
	return runtime.NewFileSystemLoader(basePaths...)
}

// NewMapLoader creates a new map loader
func NewMapLoader(templates map[string]string) *MapLoader {
	return runtime.NewMapLoader(templates)
}

// NewSecureEnvironment creates a new environment with default secure policy
func NewSecureEnvironment() *runtime.SandboxEnvironment {
	return runtime.NewSecureEnvironment()
}

// NewDevelopmentEnvironment creates a new environment with development policy
func NewDevelopmentEnvironment() *runtime.SandboxEnvironment {
	return runtime.NewDevelopmentEnvironment()
}

// NewRestrictedEnvironment creates a new environment with restricted policy
func NewRestrictedEnvironment() *runtime.SandboxEnvironment {
	return runtime.NewRestrictedEnvironment()
}

// ParseString parses a template from a string
func ParseString(source string) (*Template, error) {
	env := runtime.NewEnvironment()
	return env.NewTemplate(source)
}

// ParseFile parses a template from a file
func ParseFile(filename string) (*Template, error) {
	if filename == "" {
		return nil, runtime.NewError(runtime.ErrorTypeTemplate, "filename must not be empty", nodes.Position{}, nil)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	env := runtime.NewEnvironment()
	env.SetLoader(runtime.NewFileSystemLoader(filepath.Dir(absPath)))

	tmpl, err := env.ParseFile(filepath.Base(absPath))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// Node access for AST manipulation

// Node represents an AST node
type Node = nodes.Node

// TemplateNode represents a template AST node
type TemplateNode = nodes.Template

// DumpAST returns a string representation of the AST for debugging
func DumpAST(node Node) string {
	return nodes.Dump(node)
}

// Walk traverses the AST using the visitor pattern
func Walk(visitor nodes.Visitor, node Node) {
	nodes.Walk(visitor, node)
}

// Error types

// Error represents a Jinja2 error
type Error = runtime.Error

// ErrorType represents the type of error
type ErrorType = runtime.ErrorType

// TemplateNotFoundError indicates a single missing template
type TemplateNotFoundError = runtime.TemplateNotFoundError

// TemplatesNotFoundError indicates all candidates were missing
type TemplatesNotFoundError = runtime.TemplatesNotFoundError

// NewTemplateNotFound creates a TemplateNotFoundError
func NewTemplateNotFound(name string, tried []string, cause error) *TemplateNotFoundError {
	return runtime.NewTemplateNotFound(name, tried, cause)
}

// NewTemplatesNotFound creates a TemplatesNotFoundError
func NewTemplatesNotFound(names []string, tried []string, cause error) *TemplatesNotFoundError {
	return runtime.NewTemplatesNotFound(names, tried, cause)
}

// Security access

// GetGlobalSecurityManager returns the global security manager
func GetGlobalSecurityManager() *SecurityManager {
	return runtime.GetGlobalSecurityManager()
}

// Security policies

// DefaultSecurityPolicy returns a secure default policy
func DefaultSecurityPolicy() *runtime.SecurityPolicy {
	return runtime.DefaultSecurityPolicy()
}

// DevelopmentSecurityPolicy returns a permissive development policy
func DevelopmentSecurityPolicy() *runtime.SecurityPolicy {
	return runtime.DevelopmentSecurityPolicy()
}

// RestrictedSecurityPolicy returns a highly restrictive policy
func RestrictedSecurityPolicy() *runtime.SecurityPolicy {
	return runtime.RestrictedSecurityPolicy()
}
