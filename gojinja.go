// Package gojinja2 is a Go implementation of the Jinja2 template engine
package gojinja2

import (
	"io"
	"path/filepath"

	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/runtime"
)

// Version of the gojinja2 library
const Version = "0.1.0"

// Template represents a compiled Jinja2 template
type Template = runtime.Template
type TemplateStream = runtime.TemplateStream

// Awaitable represents a value that can be awaited inside async-enabled
// templates.
type Awaitable = runtime.Awaitable

// SimpleAwaitable mirrors Awaitable without requiring a rendering context.
type SimpleAwaitable = runtime.SimpleAwaitable

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

// SandboxEnvironment represents an environment protected by a security policy.
type SandboxEnvironment = runtime.SandboxEnvironment

// SecurityManager manages security policies
type SecurityManager = runtime.SecurityManager

// SecurityContext manages security during template execution
type SecurityContext = runtime.SecurityContext

// SecurityPolicy defines the restrictions enforced during template execution.
type SecurityPolicy = runtime.SecurityPolicy

// SecurityPolicyBuilder provides a fluent builder for constructing security policies.
type SecurityPolicyBuilder = runtime.SecurityPolicyBuilder

// SecurityLevel represents the enforcement level applied by a policy.
type SecurityLevel = runtime.SecurityLevel

const (
	// SecurityLevelDevelopment mirrors runtime.SecurityLevelDevelopment.
	SecurityLevelDevelopment SecurityLevel = runtime.SecurityLevelDevelopment
	// SecurityLevelStaging mirrors runtime.SecurityLevelStaging.
	SecurityLevelStaging SecurityLevel = runtime.SecurityLevelStaging
	// SecurityLevelProduction mirrors runtime.SecurityLevelProduction.
	SecurityLevelProduction SecurityLevel = runtime.SecurityLevelProduction
	// SecurityLevelRestricted mirrors runtime.SecurityLevelRestricted.
	SecurityLevelRestricted SecurityLevel = runtime.SecurityLevelRestricted
)

// SecurityViolation captures a policy violation that occurred during execution.
type SecurityViolation = runtime.SecurityViolation

// SecurityViolationType enumerates the classes of violations tracked by the policy.
type SecurityViolationType = runtime.SecurityViolationType

const (
	// ViolationTypeFilterAccess mirrors runtime.ViolationTypeFilterAccess.
	ViolationTypeFilterAccess SecurityViolationType = runtime.ViolationTypeFilterAccess
	// ViolationTypeFunctionAccess mirrors runtime.ViolationTypeFunctionAccess.
	ViolationTypeFunctionAccess SecurityViolationType = runtime.ViolationTypeFunctionAccess
	// ViolationTypeAttributeAccess mirrors runtime.ViolationTypeAttributeAccess.
	ViolationTypeAttributeAccess SecurityViolationType = runtime.ViolationTypeAttributeAccess
	// ViolationTypeMethodCall mirrors runtime.ViolationTypeMethodCall.
	ViolationTypeMethodCall SecurityViolationType = runtime.ViolationTypeMethodCall
	// ViolationTypeTemplateAccess mirrors runtime.ViolationTypeTemplateAccess.
	ViolationTypeTemplateAccess SecurityViolationType = runtime.ViolationTypeTemplateAccess
	// ViolationTypeRecursionLimit mirrors runtime.ViolationTypeRecursionLimit.
	ViolationTypeRecursionLimit SecurityViolationType = runtime.ViolationTypeRecursionLimit
	// ViolationTypeExecutionTimeout mirrors runtime.ViolationTypeExecutionTimeout.
	ViolationTypeExecutionTimeout SecurityViolationType = runtime.ViolationTypeExecutionTimeout
	// ViolationTypeMemoryLimit mirrors runtime.ViolationTypeMemoryLimit.
	ViolationTypeMemoryLimit SecurityViolationType = runtime.ViolationTypeMemoryLimit
	// ViolationTypeRestrictedContent mirrors runtime.ViolationTypeRestrictedContent.
	ViolationTypeRestrictedContent SecurityViolationType = runtime.ViolationTypeRestrictedContent
	// ViolationTypeInputValidation mirrors runtime.ViolationTypeInputValidation.
	ViolationTypeInputValidation SecurityViolationType = runtime.ViolationTypeInputValidation
)

// SecurityAuditEntry represents an audit log entry recorded by the security manager.
type SecurityAuditEntry = runtime.SecurityAuditEntry

// Macro represents a compiled Jinja2 macro
type Macro = runtime.Macro

// MacroNamespace represents the exported namespace returned by MakeModule
type MacroNamespace = runtime.MacroNamespace

// MacroCaller captures the caller state for call blocks
type MacroCaller = runtime.MacroCaller

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
func NewSecureEnvironment() *SandboxEnvironment {
	return runtime.NewSecureEnvironment()
}

// NewDevelopmentEnvironment creates a new environment with development policy
func NewDevelopmentEnvironment() *SandboxEnvironment {
	return runtime.NewDevelopmentEnvironment()
}

// NewRestrictedEnvironment creates a new environment with restricted policy
func NewRestrictedEnvironment() *SandboxEnvironment {
	return runtime.NewRestrictedEnvironment()
}

// NewContext creates a rendering context with the provided variables.
func NewContext(vars map[string]interface{}) *Context {
	return runtime.NewContext(vars)
}

// NewContextWithEnvironment creates a context bound to the provided environment.
func NewContextWithEnvironment(env *Environment, vars map[string]interface{}) *Context {
	return runtime.NewContextWithEnvironment(env, vars)
}

// ParseString parses a template from a string
func ParseString(source string) (*Template, error) {
	env := runtime.NewEnvironment()
	return env.NewTemplate(source)
}

// ParseStringWithName parses a template from a string with the provided name.
func ParseStringWithName(source, name string) (*Template, error) {
	return runtime.ParseStringWithName(source, name)
}

// ParseFileWithEnvironment parses a template using the provided environment's loader.
func ParseFileWithEnvironment(env *Environment, name string) (*Template, error) {
	return runtime.ParseFileWithEnvironment(env, name)
}

// GetTemplate retrieves a template by name using the provided environment.
func GetTemplate(env *Environment, name string) (*Template, error) {
	return runtime.GetTemplate(env, name)
}

// SelectTemplate resolves the first available template from the provided
// candidates using the environment loader.
func SelectTemplate(env *Environment, names []string) (*Template, error) {
	return runtime.SelectTemplate(env, names)
}

// GetOrSelectTemplate mirrors Jinja2's helper for resolving template names or
// template objects against the provided environment.
func GetOrSelectTemplate(env *Environment, target interface{}) (*Template, error) {
	return runtime.GetOrSelectTemplate(env, target)
}

// JoinPath combines a template path with its parent template name using the
// environment's loader semantics.
func JoinPath(env *Environment, template, parent string) (string, error) {
	return runtime.JoinPath(env, template, parent)
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

// ExecuteToString parses and renders a template string using a default environment.
func ExecuteToString(templateString string, vars map[string]interface{}) (string, error) {
	return runtime.ExecuteToString(templateString, vars)
}

// Execute parses and renders a template string, writing the result to the provided writer.
func Execute(templateString string, vars map[string]interface{}, writer io.Writer) error {
	return runtime.Execute(templateString, vars, writer)
}

// ExecuteToStringWithEnvironment renders a template string to a string using the provided environment.
func ExecuteToStringWithEnvironment(env *Environment, templateString string, vars map[string]interface{}) (string, error) {
	return runtime.ExecuteToStringWithEnvironment(env, templateString, vars)
}

// ExecuteWithEnvironment renders a template string to the provided writer using the supplied environment.
func ExecuteWithEnvironment(env *Environment, templateString string, vars map[string]interface{}, writer io.Writer) error {
	return runtime.ExecuteWithEnvironment(env, templateString, vars, writer)
}

// ParseAST creates a template from an AST using the default environment.
func ParseAST(ast *nodes.Template) (*Template, error) {
	return runtime.ParseAST(ast)
}

// ParseASTWithName creates a template from an AST with the specified name using the default environment.
func ParseASTWithName(ast *nodes.Template, name string) (*Template, error) {
	return runtime.ParseASTWithName(ast, name)
}

// ParseASTWithEnvironment creates a template from an AST using the provided environment.
func ParseASTWithEnvironment(env *Environment, ast *nodes.Template, name string) (*Template, error) {
	return runtime.ParseASTWithEnvironment(env, ast, name)
}

// FromString creates a template from a string using the default environment.
func FromString(templateString string) (*Template, error) {
	return runtime.FromString(templateString)
}

// FromStringWithEnvironment creates a template from a string using the provided environment.
func FromStringWithEnvironment(env *Environment, templateString string) (*Template, error) {
	return runtime.FromStringWithEnvironment(env, templateString)
}

// FromAST creates a template from an AST using the default environment.
func FromAST(ast *nodes.Template) (*Template, error) {
	return runtime.FromAST(ast)
}

// FromASTWithEnvironment creates a template from an AST using the provided environment.
func FromASTWithEnvironment(env *Environment, ast *nodes.Template) (*Template, error) {
	return runtime.FromASTWithEnvironment(env, ast)
}

// RenderTemplate renders a template string with the provided context using the default environment.
func RenderTemplate(templateString string, context map[string]interface{}) (string, error) {
	return runtime.RenderTemplate(templateString, context)
}

// RenderTemplateWithEnvironment renders a template string using the provided environment.
func RenderTemplateWithEnvironment(env *Environment, templateString string, context map[string]interface{}) (string, error) {
	return runtime.RenderTemplateWithEnvironment(env, templateString, context)
}

// RenderTemplateToWriter renders a template string to the provided writer using the default environment.
func RenderTemplateToWriter(templateString string, context map[string]interface{}, writer io.Writer) error {
	return runtime.RenderTemplateToWriter(templateString, context, writer)
}

// RenderTemplateToWriterWithEnvironment renders a template string to the provided writer using the supplied environment.
func RenderTemplateToWriterWithEnvironment(env *Environment, templateString string, context map[string]interface{}, writer io.Writer) error {
	return runtime.RenderTemplateToWriterWithEnvironment(env, templateString, context, writer)
}

// Generate renders a template string as a stream using the default environment.
func Generate(templateString string, context map[string]interface{}) (*TemplateStream, error) {
	return runtime.Generate(templateString, context)
}

// GenerateWithEnvironment renders a template string as a stream using the provided environment.
func GenerateWithEnvironment(env *Environment, templateString string, context map[string]interface{}) (*TemplateStream, error) {
	return runtime.GenerateWithEnvironment(env, templateString, context)
}

// GenerateToWriter renders a template string and streams the result to the
// provided writer using the default environment.
func GenerateToWriter(templateString string, context map[string]interface{}, writer io.Writer) (int64, error) {
	return runtime.GenerateToWriter(templateString, context, writer)
}

// GenerateToWriterWithEnvironment renders a template string using the provided
// environment and streams the output into the supplied writer.
func GenerateToWriterWithEnvironment(env *Environment, templateString string, context map[string]interface{}, writer io.Writer) (int64, error) {
	return runtime.GenerateToWriterWithEnvironment(env, templateString, context, writer)
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

// TemplateChain represents a chain of templates for inheritance-aware rendering.
type TemplateChain = runtime.TemplateChain

// NewTemplateChain creates a new template chain bound to the provided environment.
func NewTemplateChain(env *Environment) *TemplateChain {
	return runtime.NewTemplateChain(env)
}

// BatchRenderer renders multiple templates efficiently.
type BatchRenderer = runtime.BatchRenderer

// NewBatchRenderer creates a batch renderer associated with the provided environment.
func NewBatchRenderer(env *Environment) *BatchRenderer {
	return runtime.NewBatchRenderer(env)
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

// NewSecurityPolicyBuilder creates a new security policy builder.
func NewSecurityPolicyBuilder(name, description string) *SecurityPolicyBuilder {
	return runtime.NewSecurityPolicyBuilder(name, description)
}

// AddDangerousPatterns augments a policy with predefined restricted content patterns.
func AddDangerousPatterns(policy *SecurityPolicy) {
	runtime.AddDangerousPatterns(policy)
}
