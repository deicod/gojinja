# Jinja2 Go Security Guide

This guide covers the comprehensive security and sandboxing features implemented in the Go-based Jinja2 template engine.

## Overview

The security system provides:
- **Configurable security policies** with whitelist/blacklist approach
- **Sandboxed execution environment** for untrusted templates
- **Resource limits** to prevent DoS attacks
- **Input validation and output sanitization**
- **Comprehensive audit logging**
- **Protection against common vulnerabilities** (XSS, injection attacks, etc.)

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/deicod/gojinja/runtime"
)

func main() {
    // Create a secure environment (uses default security policy)
    env := runtime.NewSecureEnvironment()

    // Parse and execute template
    template, err := env.NewTemplateFromSource("Hello {{ name|upper }}!", "greeting")
    if err != nil {
        panic(err)
    }

    result, err := env.ExecuteToString(template, map[string]interface{}{
        "name": "World",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // Output: HELLO WORLD!
}
```

### Custom Security Policy

```go
// Create custom security policy
policy := runtime.NewSecurityPolicyBuilder("blog", "Blog platform policy").
    AllowFilters("upper", "lower", "title", "trim", "escape").
    AllowFunctions("range", "dict").
    AllowAttributes("user.name", "user.email", "post.title").
    BlockAllMethodCalls().
    SetMaxExecutionTime(5 * time.Second).
    SetMaxRecursionDepth(25).
    Build()

// Create environment with custom policy
env := runtime.NewEnvironment()
env.SetSecurityPolicy(policy)
env.SetSandboxed(true)
```

## Security Policies

### Predefined Policies

#### Default Policy (Production)
- Whitelist mode for maximum security
- Allows only safe text manipulation filters
- Blocks all method calls
- Strict resource limits
- Auto-escaping enabled

#### Development Policy
- Blacklist mode for flexibility
- Allows most filters and functions
- Permits some method calls
- Higher resource limits
- Auto-escaping optional

#### Restricted Policy
- Maximum security for untrusted templates
- Minimal filter and function access
- Very strict resource limits
- Blocks all potentially dangerous operations

### Custom Policy Builder

```go
policy := runtime.NewSecurityPolicyBuilder("custom", "Description").
    SetLevel(runtime.SecurityLevelProduction).
    // Filter controls
    AllowFilters("safe1", "safe2").
    BlockFilters("dangerous1", "dangerous2").
    FilterWhitelist(true). // true = whitelist mode
    // Function controls
    AllowFunctions("func1", "func2").
    BlockFunctions("sys_func1", "sys_func2").
    FunctionWhitelist(true).
    // Attribute access controls
    AllowAttributes("obj.attr1", "obj.attr2").
    AllowAttributePattern("^user\\.(name|email)$").
    BlockAttributes("password", "secret").
    // Method call controls
    AllowMethods("safe_method1").
    BlockMethods("dangerous_method1").
    BlockAllMethodCalls().
    // Resource limits
    SetMaxExecutionTime(10 * time.Second).
    SetMaxRecursionDepth(50).
    SetMaxMemoryUsage(50 * 1024 * 1024). // 50MB
    SetMaxOutputSize(5 * 1024 * 1024).    // 5MB
    // Security options
    EnableAuditLogging(true).
    BlockOnViolation(true).
    AutoEscapeOutput(true).
    ValidateAllInputs(true).
    Build()
```

### Security Levels

- **SecurityLevelDevelopment**: Permissive, suitable for development
- **SecurityLevelStaging**: Moderate security, suitable for staging
- **SecurityLevelProduction**: High security, suitable for production
- **SecurityLevelRestricted**: Maximum security, for untrusted templates

## Sandboxed Execution

### Creating a Sandbox Environment

```go
// Using predefined policy
sandbox := runtime.NewSecureEnvironment()      // Default policy
sandbox := runtime.NewDevelopmentEnvironment()  // Development policy
sandbox := runtime.NewRestrictedEnvironment()  // Restricted policy

// Using custom policy
sandbox := runtime.NewSandboxEnvironment("custom_policy_name")
```

### Template Execution with Security

```go
// Execute with security controls
err := sandbox.ExecuteTemplate(template, variables, writer)
if err != nil {
    // Handle security violations
    fmt.Printf("Template execution blocked: %v", err)
}

// Execute to string with options
result, err := runtime.ExecuteTemplateWithOptions(
    template,
    variables,
    writer,
    runtime.WithTimeout(10*time.Second),
    runtime.WithMemoryLimit(20*1024*1024),
    runtime.WithSecurityPolicy("custom_policy"),
)
```

## Input Validation and Sanitization

### Input Validation

```go
policy := runtime.NewSecurityPolicyBuilder("input", "Input validation").
    SetMaxInputLength(10 * 1024). // 10KB
    AllowRestrictedContentPattern(`^[a-zA-Z0-9\s.,!?-]+$`).
    ValidateAllInputs(true).
    Build()
```

### Output Sanitization

The security system automatically sanitizes output based on policy settings:

- **HTML escaping** for XSS protection
- **Content filtering** for dangerous patterns
- **Size limits** for output content

```go
// Manual output sanitization
secCtx, _ := manager.CreateSecurityContext("policy", "template")
sanitized := secCtx.SanitizeOutput(userInput, "template_name")
```

## Resource Limits

### Supported Limits

- **Execution Time**: Prevents long-running templates
- **Recursion Depth**: Prevents stack overflow
- **Memory Usage**: Prevents memory exhaustion
- **Output Size**: Prevents response bloat

### Configuration

```go
policy := runtime.NewSecurityPolicyBuilder("limits", "Resource limits").
    SetMaxExecutionTime(10 * time.Second).
    SetMaxRecursionDepth(100).
    SetMaxMemoryUsage(50 * 1024 * 1024). // 50MB
    SetMaxOutputSize(5 * 1024 * 1024).   // 5MB
    Build()
```

## Audit Logging

### Configuration

```go
// File-based audit logging
logger, err := runtime.NewFileAuditLogger("audit.log", 100*1024*1024, 5)
if err != nil {
    panic(err)
}

// Console audit logging
consoleLogger := runtime.NewConsoleAuditLogger(runtime.AuditLevelInfo)

// Memory audit logging (for testing)
memoryLogger := runtime.NewMemoryAuditLogger(1000)

// Multiple loggers
multiLogger := runtime.NewMultiAuditLogger(logger, consoleLogger)

// Configure global audit logging
runtime.ConfigureAuditLogging(multiLogger, runtime.AuditLevelInfo)
```

### Audit Events

The system logs various security events:

- Template access
- Filter/function access
- Security violations
- Resource limit exceeded
- Input validation failures

```go
// Get audit manager
auditManager := runtime.GetGlobalAuditManager()

// Log custom events
auditManager.LogTemplateAccess("template.html", "user123", "session456", true)
auditManager.LogResourceAccess("filter", "upper", "template.html", "rendering", true)
```

## Security Best Practices

### 1. Use Principle of Least Privilege

```go
// Good: Restrictive policy for untrusted templates
policy := runtime.NewSecurityPolicyBuilder("user_templates", "User templates").
    AllowFilters("upper", "lower", "escape").
    AllowFunctions("range").
    BlockAllMethodCalls().
    Build()

// Bad: Permissive policy for untrusted templates
policy := runtime.DevelopmentSecurityPolicy()
```

### 2. Validate All Inputs

```go
policy := runtime.NewSecurityPolicyBuilder("input", "Input validation").
    ValidateAllInputs(true).
    SetMaxInputLength(1024).
    AllowRestrictedContentPattern(`^[a-zA-Z0-9\s]+$`).
    Build()
```

### 3. Enable Auto-Escaping

```go
policy := runtime.NewSecurityPolicyBuilder("autoescape", "Autoescaping").
    AutoEscapeOutput(true).
    AllowHTMLContent(false).
    Build()
```

### 4. Set Reasonable Resource Limits

```go
policy := runtime.NewSecurityPolicyBuilder("limits", "Resource limits").
    SetMaxExecutionTime(5 * time.Second).
    SetMaxRecursionDepth(25).
    SetMaxMemoryUsage(20 * 1024 * 1024). // 20MB
    Build()
```

### 5. Enable Audit Logging

```go
// Configure comprehensive audit logging
logger := runtime.NewFileAuditLogger("/var/log/jinja2/security.log", 100*1024*1024, 10)
runtime.ConfigureAuditLogging(logger, runtime.AuditLevelInfo)
```

## Common Security Scenarios

### Blog Platform

```go
blogPolicy := runtime.NewSecurityPolicyBuilder("blog", "Blog platform").
    // Allow text manipulation
    AllowFilters("upper", "lower", "title", "trim", "striptags", "escape").
    // Allow date formatting
    AllowFilters("date", "time").
    // Safe functions
    AllowFunctions("range", "dict").
    // User content attributes
    AllowAttributes("user.name", "user.bio", "post.title", "post.content").
    // Block dangerous operations
    BlockFilters("eval", "attr", "globals").
    BlockAllMethodCalls().
    // Limits
    SetMaxExecutionTime(10 * time.Second).
    SetMaxOutputSize(1024 * 1024). // 1MB
    Build()
```

### Email Templates

```go
emailPolicy := runtime.NewSecurityPolicyBuilder("email", "Email templates").
    // Text filters only
    AllowFilters("upper", "lower", "title", "trim", "escape").
    // No functions except basic ones
    AllowFunctions("range").
    // Limited attributes
    AllowAttributes("user.name", "user.email", "order.id").
    // Strict limits
    SetMaxExecutionTime(5 * time.Second).
    SetMaxOutputSize(100 * 1024). // 100KB
    Build()
```

### User-Generated Templates

```go
userPolicy := runtime.NewSecurityPolicyBuilder("user_templates", "User templates").
    // Very restrictive
    AllowFilters("upper", "lower", "escape").
    BlockAllMethodCalls().
    // Very strict limits
    SetMaxExecutionTime(2 * time.Second).
    SetMaxRecursionDepth(10).
    SetMaxMemoryUsage(1024 * 1024). // 1MB
    SetMaxOutputSize(10 * 1024).    // 10KB
    Build()
```

## Security Violations

### Types of Violations

1. **Filter Access Violations**: Attempting to use blocked filters
2. **Function Access Violations**: Attempting to use blocked functions
3. **Attribute Access Violations**: Attempting to access blocked attributes
4. **Method Call Violations**: Attempting to call blocked methods
5. **Template Access Violations**: Attempting to access blocked templates
6. **Resource Limit Violations**: Exceeding execution time, memory, etc.
7. **Input Validation Violations**: Invalid input detected
8. **Restricted Content Violations**: Dangerous content detected

### Handling Violations

```go
// Check for violations
if secCtx.HasViolations() {
    violations := secCtx.GetViolations()
    for _, violation := range violations {
        log.Printf("Security violation: %s", violation.Description)

        if violation.Blocked {
            // Template execution was blocked
            return fmt.Errorf("template blocked due to security violation")
        }
    }
}

// Check for blocked violations specifically
if secCtx.HasBlockedViolations() {
    return fmt.Errorf("template execution blocked by security policy")
}
```

## Migration Guide

### From Unsecure to Secure

1. **Start with Default Policy**:
```go
env := runtime.NewSecureEnvironment() // Instead of runtime.NewEnvironment()
```

2. **Add Allowed Operations**:
```go
policy := runtime.DefaultSecurityPolicy()
policy.AllowedFilters["custom_filter"] = true
env.SetSecurityPolicy(policy)
```

3. **Test Templates**:
```go
result, err := env.ExecuteToString(template, vars)
if err != nil {
    // Check if security violation
    if strings.Contains(err.Error(), "security violation") {
        // Add missing permissions to policy
    }
}
```

4. **Gradually Increase Permissions**:
```go
// Start restrictive, then add what's needed
policy := runtime.RestrictedSecurityPolicy()
policy.AllowedFilters["needed_filter"] = true
```

## Performance Considerations

- Security checks add minimal overhead (~5-10% for typical templates)
- Resource limits protect against DoS attacks
- Audit logging can be disabled in production if not needed
- Security context reuse reduces allocation overhead

## Troubleshooting

### Common Issues

1. **Templates Blocked Unexpectedly**:
   - Check security policy permissions
   - Review audit logs for specific violations
   - Verify resource limits aren't too restrictive

2. **Performance Issues**:
   - Disable audit logging if not needed
   - Adjust resource limits appropriately
   - Consider policy complexity

3. **False Positives**:
   - Refine attribute patterns
   - Adjust input validation rules
   - Review blocked content patterns

### Debug Mode

```go
// Enable debug logging
runtime.ConfigureAuditLogging(
    runtime.NewConsoleAuditLogger(runtime.AuditLevelDebug),
    runtime.AuditLevelDebug,
)

// Show security violations
violations := secCtx.GetViolations()
for _, violation := range violations {
    fmt.Printf("Violation: %s\n", violation.Description)
    fmt.Printf("  Type: %s\n", violation.Type.String())
    fmt.Printf("  Context: %s\n", violation.Context)
    fmt.Printf("  Blocked: %v\n", violation.Blocked)
}
```

## Contributing to Security

When contributing to the security system:

1. **Default to Secure**: New features should be secure by default
2. **Add Tests**: Include comprehensive security tests
3. **Document**: Update security documentation
4. **Consider Edge Cases**: Think about malicious use cases
5. **Performance**: Consider performance impact of security checks

For security issues, please report them privately to maintain security.