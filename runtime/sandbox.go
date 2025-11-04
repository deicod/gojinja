package runtime

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/deicod/gojinja/nodes"
)

// SandboxEnvironment provides a secure execution environment for templates
type SandboxEnvironment struct {
	*Environment
	securityManager *SecurityManager
	policyName      string
}

// NewSandboxEnvironment creates a new sandboxed environment
func NewSandboxEnvironment(policyName string) *SandboxEnvironment {
	baseEnv := NewEnvironment()

	return &SandboxEnvironment{
		Environment:     baseEnv,
		securityManager: GetGlobalSecurityManager(),
		policyName:      policyName,
	}
}

// NewSecureEnvironment creates a new environment with default secure policy
func NewSecureEnvironment() *SandboxEnvironment {
	return NewSandboxEnvironment("default")
}

// NewDevelopmentEnvironment creates a new environment with development policy
func NewDevelopmentEnvironment() *SandboxEnvironment {
	return NewSandboxEnvironment("development")
}

// NewRestrictedEnvironment creates a new environment with restricted policy
func NewRestrictedEnvironment() *SandboxEnvironment {
	return NewSandboxEnvironment("restricted")
}

// SetSecurityPolicy sets the security policy for the sandbox
func (se *SandboxEnvironment) SetSecurityPolicy(policyName string) error {
	_, err := se.securityManager.GetPolicy(policyName)
	if err != nil {
		return err
	}

	se.policyName = policyName
	return nil
}

// GetSecurityPolicy returns the current security policy name
func (se *SandboxEnvironment) GetSecurityPolicy() string {
	return se.policyName
}

// ExecuteTemplate executes a template with security controls
func (se *SandboxEnvironment) ExecuteTemplate(template *Template, vars map[string]interface{}, writer io.Writer) error {
	// Create security context
	secCtx, err := se.securityManager.CreateSecurityContext(se.policyName, template.name)
	if err != nil {
		return fmt.Errorf("failed to create security context: %w", err)
	}
	defer se.securityManager.CleanupSecurityContext(fmt.Sprintf("%s_%d", template.name, time.Now().UnixNano()))

	// Create sandboxed context
	ctx := NewSandboxedContext(secCtx, vars, se.Environment, writer)

	// Execute template with timeout
	timeoutCtx, cancel := context.WithTimeout(context.Background(), secCtx.GetPolicy().MaxExecutionTime)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- template.ExecuteWithContext(ctx.Context)
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("template execution failed: %w", err)
		}

		// Check for security violations
		if secCtx.HasBlockedViolations() {
			violations := secCtx.GetViolations()
			return fmt.Errorf("template execution blocked due to %d security violations", len(violations))
		}

		return nil

	case <-timeoutCtx.Done():
		return fmt.Errorf("template execution timed out after %s", secCtx.GetPolicy().MaxExecutionTime)
	}
}

// ExecuteToString executes a template and returns the result as a string
func (se *SandboxEnvironment) ExecuteToString(template *Template, vars map[string]interface{}) (string, error) {
	var buf strings.Builder
	err := se.ExecuteTemplate(template, vars, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetSecurityViolations returns security violations from the last execution
// Note: This is a simplified version - in practice you'd want to track this per-execution
func (se *SandboxEnvironment) GetSecurityViolations() []*SecurityViolation {
	// This would need to be implemented with proper execution tracking
	return nil
}

// SandboxedContext provides a secure execution context
type SandboxedContext struct {
	*Context
	securityContext *SecurityContext
}

// NewSandboxedContext creates a new sandboxed context
func NewSandboxedContext(secCtx *SecurityContext, vars map[string]interface{}, env *Environment, writer io.Writer) *SandboxedContext {
	baseCtx := NewContextWithEnvironment(env, vars)

	if writer != nil {
		baseCtx.writer = writer
	}

	sc := &SandboxedContext{
		Context:         baseCtx,
		securityContext: secCtx,
	}

	// We need to embed the secure environment properly
	// For now, we'll set it as the environment and implement the interface methods
	baseCtx.environment = env

	return sc
}

// Resolve resolves a variable name with security checks
func (sc *SandboxedContext) Resolve(name string) (interface{}, error) {
	// Check if we're accessing a special variable that might be restricted
	if strings.HasPrefix(name, "__") || strings.HasPrefix(name, "_") {
		if !sc.securityContext.CheckAttributeAccess(name, sc.getCurrentTemplateName(), "variable_access") {
			return nil, fmt.Errorf("access to variable '%s' blocked by security policy", name)
		}
	}

	return sc.Context.Resolve(name)
}

// ResolveAttribute resolves an attribute access with security checks
func (sc *SandboxedContext) ResolveAttribute(obj interface{}, attr string) (interface{}, error) {
	// Build attribute path for security checking
	attributePath := sc.buildAttributePath(obj, attr)

	// Check security policy
	if !sc.securityContext.CheckAttributeAccess(attributePath, sc.getCurrentTemplateName(), "attribute_access") {
		return nil, fmt.Errorf("access to attribute '%s' blocked by security policy", attributePath)
	}

	// Call parent implementation
	return sc.Context.ResolveAttribute(obj, attr)
}

// ResolveIndex resolves an index access with security checks
func (sc *SandboxedContext) ResolveIndex(obj interface{}, index interface{}) (interface{}, error) {
	// Validate index access
	indexStr := fmt.Sprintf("%v", index)
	if !sc.securityContext.ValidateInput(indexStr, "index_access", sc.getCurrentTemplateName(), "index_resolution") {
		return nil, fmt.Errorf("index access blocked by security policy")
	}

	return sc.Context.ResolveIndex(obj, index)
}

// Set sets a variable with security checks
func (sc *SandboxedContext) Set(name string, value interface{}) {
	// Check if we're setting a restricted variable
	if strings.HasPrefix(name, "__") || strings.HasPrefix(name, "_") {
		if !sc.securityContext.CheckAttributeAccess(name, sc.getCurrentTemplateName(), "variable_assignment") {
			return // Silently block assignment
		}
	}

	sc.Context.Set(name, value)
}

// getCurrentTemplateName returns the current template name
func (sc *SandboxedContext) getCurrentTemplateName() string {
	if sc.current != nil {
		return sc.current.name
	}
	return "unknown"
}

// buildAttributePath builds an attribute path for security checking
func (sc *SandboxedContext) buildAttributePath(obj interface{}, attr string) string {
	objType := reflect.TypeOf(obj)
	if objType != nil {
		return fmt.Sprintf("%s.%s", objType.String(), attr)
	}
	return attr
}

// SecureEnvironmentWrapper wraps environment methods with security checks
type SecureEnvironmentWrapper struct {
	baseEnvironment *Environment
	securityContext *SecurityContext
}

// GetFilter returns a filter function with security checks
func (sew *SecureEnvironmentWrapper) GetFilter(name string) (FilterFunc, bool) {
	// Check security policy
	if !sew.securityContext.CheckFilterAccess(name, "unknown", "filter_lookup") {
		return nil, false
	}

	// Call parent implementation
	filter, ok := sew.baseEnvironment.GetFilter(name)
	if ok {
		// Wrap filter with security checks
		secureFilter := func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
			// Validate arguments
			for i, arg := range args {
				argStr := fmt.Sprintf("%v", arg)
				if !sew.securityContext.ValidateInput(argStr, "filter_argument", "unknown", fmt.Sprintf("filter_%s_arg_%d", name, i)) {
					return nil, fmt.Errorf("filter argument %d blocked by security policy", i)
				}
			}

			// Call original filter
			result, err := filter(ctx, value, args...)
			if err != nil {
				return nil, err
			}

			// Sanitize result if needed
			if resultStr, ok := result.(string); ok {
				result = sew.securityContext.SanitizeOutput(resultStr, "unknown")
			}

			return result, nil
		}
		return secureFilter, true
	}

	return nil, false
}

// GetTest returns a test function with security checks
func (sew *SecureEnvironmentWrapper) GetTest(name string) (TestFunc, bool) {
	// Check security policy
	if !sew.securityContext.CheckFunctionAccess(name, "unknown", "test_lookup") {
		return nil, false
	}

	// Call parent implementation
	test, ok := sew.baseEnvironment.GetTest(name)
	if ok {
		// Wrap test with security checks
		secureTest := func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error) {
			// Validate arguments
			for i, arg := range args {
				argStr := fmt.Sprintf("%v", arg)
				if !sew.securityContext.ValidateInput(argStr, "test_argument", "unknown", fmt.Sprintf("test_%s_arg_%d", name, i)) {
					return false, fmt.Errorf("test argument %d blocked by security policy", i)
				}
			}

			// Call original test
			return test(ctx, value, args...)
		}
		return secureTest, true
	}

	return nil, false
}

// GetGlobal returns a global function with security checks
func (sew *SecureEnvironmentWrapper) GetGlobal(name string) (GlobalFunc, bool) {
	// Check security policy
	if !sew.securityContext.CheckFunctionAccess(name, "unknown", "global_lookup") {
		return nil, false
	}

	// Call parent implementation
	global, ok := sew.baseEnvironment.GetGlobal(name)
	if ok {
		// Wrap global function with security checks
		secureGlobal := func(ctx *Context, args ...interface{}) (interface{}, error) {
			// Validate arguments
			for i, arg := range args {
				argStr := fmt.Sprintf("%v", arg)
				if !sew.securityContext.ValidateInput(argStr, "global_argument", "unknown", fmt.Sprintf("global_%s_arg_%d", name, i)) {
					return nil, fmt.Errorf("global argument %d blocked by security policy", i)
				}
			}

			// Call original global function
			result, err := global(ctx, args...)
			if err != nil {
				return nil, err
			}

			// Sanitize result if needed
			if resultStr, ok := result.(string); ok {
				result = sew.securityContext.SanitizeOutput(resultStr, "unknown")
			}

			return result, nil
		}
		return secureGlobal, true
	}

	return nil, false
}

// SandboxedEvaluator provides secure evaluation of template nodes
type SandboxedEvaluator struct {
	*Evaluator
	securityContext *SecurityContext
}

// NewSandboxedEvaluator creates a new sandboxed evaluator
func NewSandboxedEvaluator(ctx *Context, secCtx *SecurityContext) *SandboxedEvaluator {
	return &SandboxedEvaluator{
		Evaluator:       NewEvaluator(ctx),
		securityContext: secCtx,
	}
}

// Evaluate evaluates a node with security checks
func (se *SandboxedEvaluator) Evaluate(node nodes.Node) interface{} {
	// Check recursion limit
	templateName := "unknown"
	if se.Evaluator.ctx != nil {
		templateName = se.getTemplateName()
	}

	if !se.securityContext.CheckRecursionLimit(templateName) {
		return fmt.Errorf("recursion limit exceeded")
	}

	// Check execution time
	if !se.securityContext.CheckExecutionTime(templateName) {
		return fmt.Errorf("execution time limit exceeded")
	}

	// Delegate to parent evaluator
	result := se.Evaluator.Evaluate(node)

	// Sanitize string results
	if resultStr, ok := result.(string); ok {
		result = se.securityContext.SanitizeOutput(resultStr, templateName)

		// Update output size tracking
		se.securityContext.UpdateOutputSize(int64(len(resultStr)), templateName)
	}

	return result
}

// visitCall implements secure function calling
func (se *SandboxedEvaluator) visitCall(node *nodes.Call) interface{} {
	// Get the callable
	callable := se.Evaluate(node.Node)
	if err, ok := callable.(error); ok {
		return err
	}

	// Check if this is a method call
	if se.isMethodCall(callable, node) {
		methodName := se.extractMethodName(callable, node)
		templateName := se.getTemplateName()

		if !se.securityContext.CheckMethodCall(methodName, templateName, "method_call") {
			return fmt.Errorf("method call '%s' blocked by security policy", methodName)
		}
	}

	// Evaluate arguments securely
	args := make([]interface{}, len(node.Args))
	for i, arg := range node.Args {
		value := se.Evaluate(arg)
		if err, ok := value.(error); ok {
			return err
		}

		// Validate argument
		argStr := fmt.Sprintf("%v", value)
		if !se.securityContext.ValidateInput(argStr, "function_argument", se.getTemplateName(), fmt.Sprintf("arg_%d", i)) {
			return fmt.Errorf("function argument %d blocked by security policy", i)
		}

		args[i] = value
	}

	// Evaluate keyword arguments securely
	kwargs := make(map[string]interface{})
	for _, kwarg := range node.Kwargs {
		value := se.Evaluate(kwarg.Value)
		if err, ok := value.(error); ok {
			return err
		}

		// Validate keyword argument
		valueStr := fmt.Sprintf("%v", value)
		if !se.securityContext.ValidateInput(valueStr, "function_kwarg", se.getTemplateName(), kwarg.Key) {
			return fmt.Errorf("function keyword argument '%s' blocked by security policy", kwarg.Key)
		}

		kwargs[kwarg.Key] = value
	}

	// Call the function using the parent implementation
	return se.Evaluator.callFunction(callable, args, kwargs, node)
}

// Helper methods

func (se *SandboxedEvaluator) getTemplateName() string {
	if se.Evaluator.ctx == nil {
		return "unknown"
	}

	// Try to get template name from current template
	if se.Evaluator.ctx.current != nil {
		return se.Evaluator.ctx.current.name
	}

	// Try to get from SandboxedContext - check if current context has a method to get template name
	if se.Evaluator.ctx.current != nil {
		return se.Evaluator.ctx.current.name
	}

	return "unknown"
}

// Helper methods for method call detection

func (se *SandboxedEvaluator) isMethodCall(callable interface{}, node *nodes.Call) bool {
	// This is a simplified implementation
	// In practice, you'd want more sophisticated method call detection
	_, isMacro := callable.(*Macro)
	return !isMacro && callable != nil
}

func (se *SandboxedEvaluator) extractMethodName(callable interface{}, node *nodes.Call) string {
	// This is a simplified implementation
	// In practice, you'd want to extract the actual method name
	if nameNode, ok := node.Node.(*nodes.Getattr); ok {
		return nameNode.Attr
	}
	return "unknown"
}

// Template execution options
type ExecutionOption func(*executionConfig)

type executionConfig struct {
	timeout     time.Duration
	memoryLimit int64
	outputLimit int64
	enableAudit bool
	policyName  string
}

// WithTimeout sets the execution timeout
func WithTimeout(timeout time.Duration) ExecutionOption {
	return func(config *executionConfig) {
		config.timeout = timeout
	}
}

// WithMemoryLimit sets the memory limit
func WithMemoryLimit(limit int64) ExecutionOption {
	return func(config *executionConfig) {
		config.memoryLimit = limit
	}
}

// WithOutputLimit sets the output limit
func WithOutputLimit(limit int64) ExecutionOption {
	return func(config *executionConfig) {
		config.outputLimit = limit
	}
}

// WithAuditLogging enables audit logging
func WithAuditLogging(enable bool) ExecutionOption {
	return func(config *executionConfig) {
		config.enableAudit = enable
	}
}

// WithSecurityPolicy sets the security policy
func WithSecurityPolicy(policyName string) ExecutionOption {
	return func(config *executionConfig) {
		config.policyName = policyName
	}
}

// ExecuteTemplateWithOptions executes a template with custom options
func ExecuteTemplateWithOptions(template *Template, vars map[string]interface{}, writer io.Writer, options ...ExecutionOption) error {
	config := &executionConfig{
		timeout:     30 * time.Second,
		memoryLimit: 10 * 1024 * 1024, // 10MB
		outputLimit: 1024 * 1024,      // 1MB
		enableAudit: true,
		policyName:  "default",
	}

	// Apply options
	for _, option := range options {
		option(config)
	}

	// Create sandbox environment
	sandbox := NewSandboxEnvironment(config.policyName)

	// Create custom policy if limits are specified
	if config.timeout != 0 || config.memoryLimit != 0 || config.outputLimit != 0 {
		policy := sandbox.securityManager.policies[config.policyName]
		if policy != nil {
			customPolicy := policy.Clone()
			if config.timeout > 0 {
				customPolicy.MaxExecutionTime = config.timeout
			}
			if config.memoryLimit > 0 {
				customPolicy.MaxMemoryUsage = config.memoryLimit
			}
			if config.outputLimit > 0 {
				customPolicy.MaxOutputSize = config.outputLimit
			}
			customPolicy.EnableAuditLogging = config.enableAudit

			// Register custom policy
			policyName := fmt.Sprintf("custom_%d", time.Now().UnixNano())
			sandbox.securityManager.AddPolicy(policyName, customPolicy)
			sandbox.SetSecurityPolicy(policyName)
		}
	}

	// Execute template
	return sandbox.ExecuteTemplate(template, vars, writer)
}
