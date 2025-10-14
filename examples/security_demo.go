package examples

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/deicod/gojinja/runtime"
	"github.com/deicod/gojinja/parser"
)

func RunSecurityDemo() {
	fmt.Println("=== Jinja2 Security System Demo ===")
	fmt.Println()

	// Demo 1: Basic Security Policies
	fmt.Println("1. Basic Security Policies:")
	demoBasicPolicies()

	// Demo 2: Custom Security Policy Builder
	fmt.Println("\n2. Custom Security Policy Builder:")
	demoCustomPolicyBuilder()

	// Demo 3: Security Context and Violations
	fmt.Println("\n3. Security Context and Violations:")
	demoSecurityContext()

	// Demo 4: Sandbox Environment
	fmt.Println("\n4. Sandbox Environment:")
	demoSandboxEnvironment()

	// Demo 5: Input Validation and Sanitization
	fmt.Println("\n5. Input Validation and Sanitization:")
	demoInputValidation()

	// Demo 6: Audit Logging
	fmt.Println("\n6. Audit Logging:")
	demoAuditLogging()

	// Demo 7: Resource Limits
	fmt.Println("\n7. Resource Limits:")
	demoResourceLimits()

	// Demo 8: Real-world Template Security
	fmt.Println("\n8. Real-world Template Security:")
	demoRealWorldSecurity()
}

func demoBasicPolicies() {
	// Show predefined security policies
	fmt.Println("  Default Policy (Production):")
	defaultPolicy := runtime.DefaultSecurityPolicy()
	fmt.Printf("    - Filter Whitelist: %v\n", defaultPolicy.FilterWhitelist)
	fmt.Printf("    - Block on Violation: %v\n", defaultPolicy.BlockOnViolation)
	fmt.Printf("    - Max Execution Time: %v\n", defaultPolicy.MaxExecutionTime)
	fmt.Printf("    - Allowed Filters: %v\n", getMapKeys(defaultPolicy.AllowedFilters))

	fmt.Println("\n  Development Policy:")
	devPolicy := runtime.DevelopmentSecurityPolicy()
	fmt.Printf("    - Filter Whitelist: %v\n", devPolicy.FilterWhitelist)
	fmt.Printf("    - Block on Violation: %v\n", devPolicy.BlockOnViolation)
	fmt.Printf("    - Max Execution Time: %v\n", devPolicy.MaxExecutionTime)

	fmt.Println("\n  Restricted Policy:")
	restrictedPolicy := runtime.RestrictedSecurityPolicy()
	fmt.Printf("    - Block All Methods: %v\n", restrictedPolicy.BlockAllMethods)
	fmt.Printf("    - Max Recursion Depth: %d\n", restrictedPolicy.MaxRecursionDepth)
	fmt.Printf("    - Max Memory Usage: %d bytes\n", restrictedPolicy.MaxMemoryUsage)
}

func demoCustomPolicyBuilder() {
	// Create custom security policy
	policy := runtime.NewSecurityPolicyBuilder("custom-blog", "Policy for blog templates").
		SetLevel(runtime.SecurityLevelProduction).
		// Allow safe filters
		AllowFilters("upper", "lower", "title", "trim", "length", "escape", "safe").
		// Block dangerous filters
		BlockFilters("eval", "attr", "globals", "locals", "vars").
		// Allow safe functions
		AllowFunctions("range", "dict", "cycler", "joiner").
		// Block system functions
		BlockFunctions("open", "file", "exec", "import").
		// Allow user object attributes
		AllowAttributes("user.name", "user.email", "user.avatar", "user.bio").
		AllowAttributePattern("^post\\.(title|content|date|author)$").
		// Block all method calls
		BlockAllMethodCalls().
		// Set reasonable limits
		SetMaxExecutionTime(5 * time.Second).
		SetMaxRecursionDepth(25).
		SetMaxMemoryUsage(20 * 1024 * 1024). // 20MB
		SetMaxOutputSize(1024 * 1024).       // 1MB
		// Security options
		EnableAuditLogging(true).
		BlockOnViolation(true).
		AutoEscapeOutput(true).
		ValidateAllInputs(true).
		Build()

	fmt.Printf("  Custom Policy '%s':\n", policy.Name)
	fmt.Printf("    - Description: %s\n", policy.Description)
	fmt.Printf("    - Security Level: %s\n", policy.Level.String())
	fmt.Printf("    - Allowed Filters: %v\n", getMapKeys(policy.AllowedFilters))
	fmt.Printf("    - Blocked Filters: %v\n", getMapKeys(policy.BlockedFilters))
	fmt.Printf("    - Allowed Functions: %v\n", getMapKeys(policy.AllowedFunctions))
	fmt.Printf("    - Block All Methods: %v\n", policy.BlockAllMethods)
}

func demoSecurityContext() {
	// Create security manager and context
	manager := runtime.NewSecurityManager()

	// Add custom policy
	policy := runtime.NewSecurityPolicyBuilder("demo", "Demo policy").
		AllowFilters("upper", "lower").
		BlockFilters("eval").
		Build()

	manager.AddPolicy("demo", policy)

	// Create security context
	secCtx, err := manager.CreateSecurityContext("demo", "test_template")
	if err != nil {
		log.Fatalf("Failed to create security context: %v", err)
	}
	defer manager.CleanupSecurityContext("test_template_123")

	// Test allowed operations
	fmt.Println("  Testing allowed operations:")
	allowed := secCtx.CheckFilterAccess("upper", "test_template", "demo")
	fmt.Printf("    - Filter 'upper' access: %v\n", allowed)

	allowed = secCtx.CheckFunctionAccess("range", "test_template", "demo")
	fmt.Printf("    - Function 'range' access: %v\n", allowed)

	// Test blocked operations
	fmt.Println("\n  Testing blocked operations:")
	blocked := secCtx.CheckFilterAccess("eval", "test_template", "demo")
	fmt.Printf("    - Filter 'eval' access: %v\n", blocked)

	// Show violations
	violations := secCtx.GetViolations()
	fmt.Printf("\n  Security Violations: %d\n", len(violations))
	for i, violation := range violations {
		fmt.Printf("    %d. %s: %s\n", i+1, violation.Type.String(), violation.Description)
		fmt.Printf("       Severity: %s, Blocked: %v\n", violation.Severity, violation.Blocked)
	}
}

func demoSandboxEnvironment() {
	// Create sandbox environment
	sandbox := runtime.NewSecureEnvironment()

	// Create a simple template
	templateContent := "Hello {{ name|upper }}!"

	// Parse template
	parserEnv := &parser.Environment{}
	ast, err := parser.ParseTemplateWithEnv(parserEnv, templateContent, "test", "test")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	// Create template
	template, err := sandbox.NewTemplateFromAST(ast, "test")
	if err != nil {
		log.Fatalf("Failed to create template: %v", err)
	}

	// Execute template
	vars := map[string]interface{}{
		"name": "World",
	}

	result, err := sandbox.ExecuteToString(template, vars)
	if err != nil {
		log.Printf("Template execution failed: %v", err)
	} else {
		fmt.Printf("  Template Result: %s\n", result)
	}

	// Test with blocked filter
	dangerousTemplate := "Hello {{ name|eval }}!"
	ast, err = parser.ParseTemplateWithEnv(parserEnv, dangerousTemplate, "dangerous", "dangerous")
	if err != nil {
		log.Printf("Failed to parse dangerous template: %v", err)
		return
	}

	dangerousTmpl, err := sandbox.NewTemplateFromAST(ast, "dangerous")
	if err != nil {
		log.Printf("Failed to create dangerous template: %v", err)
		return
	}

	_, err = sandbox.ExecuteToString(dangerousTmpl, vars)
	if err != nil {
		fmt.Printf("  Dangerous template blocked: %v\n", err)
	}
}

func demoInputValidation() {
	policy := runtime.NewSecurityPolicyBuilder("input-test", "Input validation test").
		SetMaxInputLength(100).
		AllowRestrictedContentPattern(`^[a-zA-Z0-9\s.,!?]+$`).
		AllowHTMLContent(false).
		Build()

	// Test cases
	testCases := []struct {
		input   string
		valid   bool
		reason  string
	}{
		{"Hello World 123!", true, "Valid text"},
		{strings.Repeat("a", 200), false, "Too long"},
		{"Hello <script>alert('xss')</script>", false, "Contains HTML/JavaScript"},
		{"User input: drop table users;", false, "Contains SQL injection pattern"},
		{"Normal blog post title here.", true, "Valid blog title"},
	}

	fmt.Println("  Input Validation Tests:")
	for i, tc := range testCases {
		allowed, violation := policy.ValidateInput(tc.input, "test")
		status := "✓"
		if !allowed {
			status = "✗"
		}

		fmt.Printf("    %d. %s '%s' - %s\n", i+1, status, tc.input[:min(len(tc.input), 30)], tc.reason)
		if violation != nil {
			fmt.Printf("       Violation: %s\n", violation.Description)
		}
	}
}

func demoAuditLogging() {
	// Create memory audit logger
	logger := runtime.NewMemoryAuditLogger(1000)

	// Configure audit logging
	runtime.ConfigureAuditLogging(logger, runtime.AuditLevelInfo)

	// Get audit manager
	auditManager := runtime.GetGlobalAuditManager()

	// Log some events
	auditManager.LogTemplateAccess("blog_post.html", "user123", "session456", true)
	auditManager.LogResourceAccess("filter", "upper", "blog_post.html", "template_rendering", true)
	auditManager.LogResourceAccess("filter", "eval", "blog_post.html", "template_rendering", false)

	violation := &runtime.SecurityViolation{
		Type:        runtime.ViolationTypeFilterAccess,
		Description: "Access to blocked filter 'eval'",
		Context:     "template_rendering",
		Timestamp:   time.Now(),
		Severity:    "high",
		Blocked:     true,
	}
	auditManager.LogSecurityViolation(violation, "blog_post.html", "template_rendering")

	// Show audit log
	events := logger.GetEvents()
	fmt.Printf("  Audit Log Entries: %d\n", len(events))
	for i, event := range events {
		fmt.Printf("    %d. [%s] %s: %s\n",
			i+1,
			event.Timestamp.Format("15:04:05"),
			event.Type.String(),
			event.Message)
		if !event.Success {
			fmt.Printf("       Status: FAILED\n")
		}
	}
}

func demoResourceLimits() {
	manager := runtime.NewSecurityManager()
	secCtx, err := manager.CreateSecurityContext("default", "limits_test")
	if err != nil {
		log.Fatalf("Failed to create security context: %v", err)
	}
	defer manager.CleanupSecurityContext("limits_test_123")

	fmt.Println("  Testing Resource Limits:")

	// Test recursion limit
	fmt.Print("    Testing recursion limit... ")
	recursionAllowed := true
	for i := 0; i < 150; i++ {
		if !secCtx.CheckRecursionLimit("test") {
			recursionAllowed = false
			break
		}
	}
	if !recursionAllowed {
		fmt.Println("✓ Recursion limit enforced")
	} else {
		fmt.Println("✗ Recursion limit not enforced")
	}

	// Test memory limit
	fmt.Print("    Testing memory limit... ")
	memoryAllowed := secCtx.UpdateMemoryUsage(100*1024*1024, "test") // 100MB
	if !memoryAllowed {
		fmt.Println("✓ Memory limit enforced")
	} else {
		fmt.Println("✗ Memory limit not enforced")
	}

	// Test output size limit
	fmt.Print("    Testing output size limit... ")
	outputAllowed := secCtx.UpdateOutputSize(20*1024*1024, "test") // 20MB
	if !outputAllowed {
		fmt.Println("✓ Output size limit enforced")
	} else {
		fmt.Println("✗ Output size limit not enforced")
	}

	// Show statistics
	stats := secCtx.GetExecutionStats()
	fmt.Printf("\n  Execution Statistics:\n")
	fmt.Printf("    - Violations: %d\n", stats["violations_count"])
	fmt.Printf("    - Memory Usage: %d bytes\n", stats["memory_usage"])
	fmt.Printf("    - Output Size: %d bytes\n", stats["output_size"])
}

func demoRealWorldSecurity() {
	// Create a realistic security policy for a blog platform
	blogPolicy := runtime.NewSecurityPolicyBuilder("blog-platform", "Security policy for blog platform").
		SetLevel(runtime.SecurityLevelProduction).
		// Allow safe text manipulation filters
		AllowFilters("upper", "lower", "title", "capitalize", "trim", "striptags", "escape", "safe", "length", "wordcount", "truncate", "default").
		// Allow formatting filters
		AllowFilters("date", "time", "datetime", "filesizeformat", "int", "float", "round").
		// Block dangerous filters
		BlockFilters("eval", "attr", "globals", "locals", "vars", "getitem", "getattribute").
		// Allow safe functions
		AllowFunctions("range", "dict", "cycler", "joiner", "lipsum").
		// Block system functions
		BlockFunctions("open", "file", "exec", "import", "reload", "super").
		// Allow user and post attributes
		AllowAttributes("user.name", "user.email", "user.avatar", "user.bio", "user.joined_date").
		AllowAttributePattern("^post\\.(title|content|excerpt|date|author|category|tags)$").
		AllowAttributePattern("^comment\\.(author|content|date)$").
		// Allow config attributes
		AllowAttributePattern("^config\\.(site_name|site_url|theme|language)$").
		// Block dangerous attribute access
		BlockAttributes("password", "secret", "token", "key", "private", "admin").
		BlockAttributes("password", "secret", "token", "key", "private").
		// Block all method calls
		BlockAllMethodCalls().
		// Set resource limits
		SetMaxExecutionTime(10 * time.Second).
		SetMaxRecursionDepth(50).
		SetMaxMemoryUsage(50 * 1024 * 1024). // 50MB
		SetMaxOutputSize(5 * 1024 * 1024).   // 5MB
		// Input validation
		SetMaxInputLength(10 * 1024). // 10KB
		ValidateAllInputs(true).
		// Content restrictions
		AllowHTMLContent(false).
		AllowJavaScriptContent(false).
		AllowCSSContent(false).
		// Security options
		EnableAuditLogging(true).
		BlockOnViolation(true).
		AutoEscapeOutput(true).
		Build()

	// Create security manager with blog policy
	manager := runtime.NewSecurityManager()
	manager.AddPolicy("blog", blogPolicy)

	// Test scenarios
	scenarios := []struct {
		name        string
		template    string
		vars        map[string]interface{}
		shouldSucceed bool
	}{
		{
			name:        "Valid blog post template",
			template:    "<h1>{{ post.title|title }}</h1><p>By {{ user.name|upper }} on {{ post.date|date('Y-m-d') }}</p><p>{{ post.content|striptags|truncate(200) }}</p>",
			vars:        map[string]interface{}{"user": map[string]interface{}{"name": "John Doe"}, "post": map[string]interface{}{"title": "My Blog Post", "content": "This is a great blog post with interesting content.", "date": "2023-01-01"}},
			shouldSucceed: true,
		},
		{
			name:        "Template with blocked filter",
			template:    "{{ user.name|eval }}",
			vars:        map[string]interface{}{"user": map[string]interface{}{"name": "__import__('os').system('ls')"}},
			shouldSucceed: false,
		},
		{
			name:        "Template with blocked attribute access",
			template:    "{{ user.password }}",
			vars:        map[string]interface{}{"user": map[string]interface{}{"name": "John", "password": "secret123"}},
			shouldSucceed: false,
		},
		{
			name:        "Template with dangerous content",
			template:    "<script>alert('xss')</script>{{ user.name }}",
			vars:        map[string]interface{}{"user": map[string]interface{}{"name": "John"}},
			shouldSucceed: false,
		},
	}

	fmt.Println("  Real-world Security Scenarios:")
	parserEnv := &parser.Environment{}

	for i, scenario := range scenarios {
		fmt.Printf("    %d. %s\n", i+1, scenario.name)
		fmt.Printf("       Template: %s\n", scenario.template)

		// Parse template
		_, err := parser.ParseTemplateWithEnv(parserEnv, scenario.template, "test", "test")
		if err != nil {
			fmt.Printf("       Parse error: %v\n", err)
			continue
		}

		// Create security context
		secCtx, err := manager.CreateSecurityContext("blog", "test")
		if err != nil {
			fmt.Printf("       Security context error: %v\n", err)
			continue
		}

		// Test template execution (simplified)
		var blocked bool
		if strings.Contains(scenario.template, "eval") {
			blocked = !secCtx.CheckFilterAccess("eval", "test", "scenario_test")
		} else if strings.Contains(scenario.template, "password") {
			blocked = !secCtx.CheckAttributeAccess("user.password", "test", "scenario_test")
		} else if strings.Contains(scenario.template, "<script>") {
			allowed, _ := blogPolicy.ValidateInput(scenario.template, "template")
			blocked = !allowed
		}

		if blocked == scenario.shouldSucceed {
			fmt.Printf("       Result: ✗ Unexpected behavior\n")
		} else {
			status := "BLOCKED"
			if !blocked {
				status = "ALLOWED"
			}
			fmt.Printf("       Result: ✓ %s\n", status)
		}

		// Show violations if any
		violations := secCtx.GetViolations()
		if len(violations) > 0 {
			fmt.Printf("       Violations: %d\n", len(violations))
			for _, violation := range violations {
				fmt.Printf("         - %s\n", violation.Description)
			}
		}

		manager.CleanupSecurityContext(fmt.Sprintf("test_%d", time.Now().UnixNano()))
	}
}

// Helper functions
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}