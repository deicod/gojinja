package runtime

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/deicod/gojinja/nodes"
)

// ErrorType represents different types of runtime errors
type ErrorType string

const (
	ErrorTypeTemplate   ErrorType = "template_error"
	ErrorTypeUndefined  ErrorType = "undefined_error"
	ErrorTypeSyntax     ErrorType = "syntax_error"
	ErrorTypeSecurity   ErrorType = "security_error"
	ErrorTypeFilter     ErrorType = "filter_error"
	ErrorTypeTest       ErrorType = "test_error"
	ErrorTypeRange      ErrorType = "range_error"
	ErrorTypeAssignment ErrorType = "assignment_error"
	ErrorTypeContext    ErrorType = "context_error"
	ErrorTypeMacro      ErrorType = "macro_error"
	ErrorTypeImport     ErrorType = "import_error"
)

// Error represents a runtime error with position information
type Error struct {
	Type     ErrorType
	Message  string
	Position nodes.Position
	Node     nodes.Node
	Cause    error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Position.Line > 0 {
		if e.Position.Column > 0 {
			return fmt.Sprintf("%s at line %d, column %d: %s", e.Type, e.Position.Line, e.Position.Column, e.Message)
		}
		return fmt.Sprintf("%s at line %d: %s", e.Type, e.Position.Line, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying cause
func (e *Error) Unwrap() error {
	return e.Cause
}

// NewError creates a new runtime error
func NewError(errorType ErrorType, message string, position nodes.Position, node nodes.Node) *Error {
	return &Error{
		Type:     errorType,
		Message:  message,
		Position: position,
		Node:     node,
	}
}

// NewErrorWithCause creates a new runtime error with an underlying cause
func NewErrorWithCause(errorType ErrorType, message string, position nodes.Position, node nodes.Node, cause error) *Error {
	return &Error{
		Type:     errorType,
		Message:  message,
		Position: position,
		Node:     node,
		Cause:    cause,
	}
}

// WrapError wraps an existing error with position information
func WrapError(err error, position nodes.Position, node nodes.Node) error {
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	case *Error:
		if position.Line != 0 {
			e.Position = position
		}
		if node != nil {
			e.Node = node
		}
		return e
	case *TemplateNotFoundError:
		if base := e.runtimeError(); base != nil {
			if position.Line != 0 {
				base.Position = position
			}
			if node != nil {
				base.Node = node
			}
		}
		return e
	case *TemplatesNotFoundError:
		if base := e.runtimeError(); base != nil {
			if position.Line != 0 {
				base.Position = position
			}
			if node != nil {
				base.Node = node
			}
		}
		return e
	default:
		return &Error{
			Type:     ErrorTypeTemplate,
			Message:  err.Error(),
			Position: position,
			Node:     node,
			Cause:    err,
		}
	}
}

// UndefinedError represents an undefined variable error
type UndefinedError struct {
	error
	Name string
}

// NewUndefinedError creates a new undefined variable error
func NewUndefinedError(name string, position nodes.Position, node nodes.Node) *UndefinedError {
	return &UndefinedError{
		error: NewError(ErrorTypeUndefined, fmt.Sprintf("'%s' is undefined", name), position, node),
		Name:  name,
	}
}

// SecurityError represents a security-related error
type SecurityError struct {
	error
	Operation string
}

// NewSecurityError creates a new security error
func NewSecurityError(operation, message string, position nodes.Position, node nodes.Node) *SecurityError {
	return &SecurityError{
		error:     NewError(ErrorTypeSecurity, message, position, node),
		Operation: operation,
	}
}

// FilterError represents a filter-related error
type FilterError struct {
	error
	FilterName string
}

// NewFilterError creates a new filter error
func NewFilterError(filterName, message string, position nodes.Position, node nodes.Node, cause error) *FilterError {
	return &FilterError{
		error:      NewErrorWithCause(ErrorTypeFilter, fmt.Sprintf("filter '%s': %s", filterName, message), position, node, cause),
		FilterName: filterName,
	}
}

// TestError represents a test-related error
type TestError struct {
	error
	TestName string
}

// NewTestError creates a new test error
func NewTestError(testName, message string, position nodes.Position, node nodes.Node, cause error) *TestError {
	return &TestError{
		error:    NewErrorWithCause(ErrorTypeTest, fmt.Sprintf("test '%s': %s", testName, message), position, node, cause),
		TestName: testName,
	}
}

// AssignmentError represents an assignment-related error
type AssignmentError struct {
	error
	Target string
}

// NewAssignmentError creates a new assignment error
func NewAssignmentError(target, message string, position nodes.Position, node nodes.Node) *AssignmentError {
	return &AssignmentError{
		error:  NewError(ErrorTypeAssignment, fmt.Sprintf("cannot assign to %s: %s", target, message), position, node),
		Target: target,
	}
}

// ContextError represents a context-related error
type ContextError struct {
	error
	Context string
}

// NewContextError creates a new context error
func NewContextError(context, message string, position nodes.Position, node nodes.Node) *ContextError {
	return &ContextError{
		error:   NewError(ErrorTypeContext, fmt.Sprintf("%s: %s", context, message), position, node),
		Context: context,
	}
}

// ErrorWithCallStack adds call stack information to an error
func ErrorWithCallStack(err error) *Error {
	if runtimeErr, ok := err.(*Error); ok {
		// Add caller information to the error message
		_, file, line, ok := runtime.Caller(1)
		if ok {
			runtimeErr.Message = fmt.Sprintf("%s (called from %s:%d)", runtimeErr.Message, file, line)
		}
		return runtimeErr
	}

	// Wrap non-runtime errors
	_, file, line, ok := runtime.Caller(1)
	if ok {
		return &Error{
			Type:    ErrorTypeTemplate,
			Message: fmt.Sprintf("%s (called from %s:%d)", err.Error(), file, line),
			Cause:   err,
		}
	}

	return &Error{
		Type:    ErrorTypeTemplate,
		Message: err.Error(),
		Cause:   err,
	}
}

// IsUndefinedError checks if an error is an undefined variable error
func IsUndefinedError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*UndefinedError)
	return ok
}

// TemplateNotFoundError represents an error when a single template cannot be located.
type TemplateNotFoundError struct {
	base  *Error
	Name  string
	Tried []string
}

// NewTemplateNotFound creates a TemplateNotFoundError with optional tried locations and cause.
func NewTemplateNotFound(name string, tried []string, cause error) *TemplateNotFoundError {
	message := fmt.Sprintf("template %s not found", name)
	if len(tried) > 0 {
		message = fmt.Sprintf("%s (tried: %s)", message, strings.Join(tried, ", "))
	}

	return &TemplateNotFoundError{
		base:  NewErrorWithCause(ErrorTypeTemplate, message, nodes.Position{}, nil, cause),
		Name:  name,
		Tried: append([]string(nil), tried...),
	}
}

// TemplatesNotFoundError represents an error when none of a set of templates can be located.
type TemplatesNotFoundError struct {
	base  *Error
	Names []string
	Tried []string
}

// NewTemplatesNotFound creates a TemplatesNotFoundError with optional tried locations and cause.
func NewTemplatesNotFound(names []string, tried []string, cause error) *TemplatesNotFoundError {
	message := "no templates found"
	if len(names) > 0 {
		message = fmt.Sprintf("no templates found among [%s]", strings.Join(names, ", "))
	}
	if len(tried) > 0 {
		message = fmt.Sprintf("%s (tried: %s)", message, strings.Join(tried, ", "))
	}

	return &TemplatesNotFoundError{
		base:  NewErrorWithCause(ErrorTypeTemplate, message, nodes.Position{}, nil, cause),
		Names: append([]string(nil), names...),
		Tried: append([]string(nil), tried...),
	}
}

// Error returns the message for TemplateNotFoundError.
func (e *TemplateNotFoundError) Error() string {
	if e == nil {
		return "template not found"
	}
	if e.base != nil {
		return e.base.Error()
	}
	return fmt.Sprintf("template %s not found", e.Name)
}

// Error returns the message for TemplatesNotFoundError.
func (e *TemplatesNotFoundError) Error() string {
	if e == nil {
		return "no templates found"
	}
	if e.base != nil {
		return e.base.Error()
	}
	return "no templates found"
}

// Unwrap returns the underlying cause for TemplatesNotFoundError.
func (e *TemplatesNotFoundError) Unwrap() error {
	if e == nil || e.base == nil {
		return nil
	}
	return e.base.Cause
}

// Unwrap returns the underlying cause for TemplateNotFoundError.
func (e *TemplateNotFoundError) Unwrap() error {
	if e == nil || e.base == nil {
		return nil
	}
	return e.base.Cause
}

func (e *TemplateNotFoundError) runtimeError() *Error {
	if e == nil {
		return nil
	}
	return e.base
}

func (e *TemplatesNotFoundError) runtimeError() *Error {
	if e == nil {
		return nil
	}
	return e.base
}

// IsSecurityError checks if an error is a security error
func IsSecurityError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*SecurityError)
	return ok
}

// IsFilterError checks if an error is a filter error
func IsFilterError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*FilterError)
	return ok
}

// IsAssignmentError checks if an error is an assignment error
func IsAssignmentError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*AssignmentError)
	return ok
}

// MacroError represents a macro-related error
type MacroError struct {
	error
	MacroName string
}

// NewMacroError creates a new macro error
func NewMacroError(macroName, message string, position nodes.Position, node nodes.Node) *MacroError {
	return &MacroError{
		error:     NewError(ErrorTypeMacro, fmt.Sprintf("macro '%s': %s", macroName, message), position, node),
		MacroName: macroName,
	}
}

// ImportError represents an import-related error
type ImportError struct {
	error
	TemplateName string
}

// NewImportError creates a new import error
func NewImportError(templateName, message string, position nodes.Position, node nodes.Node) *ImportError {
	return &ImportError{
		error:        NewError(ErrorTypeImport, fmt.Sprintf("import '%s': %s", templateName, message), position, node),
		TemplateName: templateName,
	}
}

// IsMacroError checks if an error is a macro error
func IsMacroError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*MacroError)
	return ok
}

// IsImportError checks if an error is an import error
func IsImportError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*ImportError)
	return ok
}
