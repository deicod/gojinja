package examples

import (
	"fmt"
	"log"
	"time"

	"github.com/deicod/gojinja/parser"
	"github.com/deicod/gojinja/runtime"
)

func RunSecurityQuickTest() {
	fmt.Println("=== Security System Quick Test ===")
	fmt.Println()

	// Test 1: Basic security functionality
	fmt.Println("1. Testing basic security functionality...")
	testBasicSecurity()

	// Test 2: Sandbox environment
	fmt.Println("\n2. Testing sandbox environment...")
	testSandboxEnvironment()

	// Test 3: Security policy builder
	fmt.Println("\n3. Testing security policy builder...")
	testSecurityPolicyBuilder()

	// Test 4: Input validation
	fmt.Println("\n4. Testing input validation...")
	testInputValidation()

	fmt.Println("\n✅ All security tests passed!")
}

func testBasicSecurity() {
	// Test security manager
	manager := runtime.NewSecurityManager()
	if manager == nil {
		log.Fatal("Failed to create security manager")
	}

	// Test default policy
	policy := runtime.DefaultSecurityPolicy()
	if policy == nil {
		log.Fatal("Failed to create default security policy")
	}

	// Test security context
	secCtx, err := manager.CreateSecurityContext("default", "test")
	if err != nil {
		log.Fatalf("Failed to create security context: %v", err)
	}
	defer manager.CleanupSecurityContext("test_123")

	// Test filter access
	allowed := secCtx.CheckFilterAccess("upper", "test", "test")
	if !allowed {
		log.Fatal("Expected 'upper' filter to be allowed")
	}

	// Test audit logging
	auditManager := runtime.GetGlobalAuditManager()
	auditManager.LogTemplateAccess("test.html", "user123", "session456", true)

	fmt.Println("   ✓ Basic security functionality works")
}

func testSandboxEnvironment() {
	// Create sandbox environment
	sandbox := runtime.NewSecureEnvironment()
	if sandbox == nil {
		log.Fatal("Failed to create sandbox environment")
	}

	// Test template parsing and execution
	templateContent := "Hello {{ name|upper }}!"
	parserEnv := &parser.Environment{}
	ast, err := parser.ParseTemplateWithEnv(parserEnv, templateContent, "test", "test")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

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
		log.Fatalf("Failed to execute template: %v", err)
	}

	expected := "HELLO WORLD!"
	if result != expected {
		log.Fatalf("Expected '%s', got '%s'", expected, result)
	}

	fmt.Println("   ✓ Sandbox environment works")
}

func testSecurityPolicyBuilder() {
	// Create custom policy
	policy := runtime.NewSecurityPolicyBuilder("test", "Test policy").
		SetLevel(runtime.SecurityLevelProduction).
		AllowFilters("upper", "lower").
		BlockFilters("eval").
		AllowFunctions("range", "dict").
		BlockFunctions("open").
		SetMaxExecutionTime(5 * time.Second).
		SetMaxRecursionDepth(10).
		Build()

	if policy.Name != "test" {
		log.Fatalf("Expected policy name 'test', got '%s'", policy.Name)
	}

	// Test filter access
	allowed, violation := policy.IsFilterAllowed("upper")
	if !allowed || violation != nil {
		log.Fatal("Expected 'upper' filter to be allowed")
	}

	allowed, violation = policy.IsFilterAllowed("eval")
	if allowed || violation == nil {
		log.Fatal("Expected 'eval' filter to be blocked")
	}

	// Test function access
	allowed, violation = policy.IsFunctionAllowed("range")
	if !allowed || violation != nil {
		log.Fatal("Expected 'range' function to be allowed")
	}

	allowed, violation = policy.IsFunctionAllowed("open")
	if allowed || violation == nil {
		log.Fatal("Expected 'open' function to be blocked")
	}

	fmt.Println("   ✓ Security policy builder works")
}

func testInputValidation() {
	// Create policy with input validation
	policy := runtime.NewSecurityPolicyBuilder("input-test", "Input validation test").
		SetMaxInputLength(100).
		AllowRestrictedContentPattern(`^[a-zA-Z0-9\s]+$`).
		Build()

	// Test valid input
	allowed, violation := policy.ValidateInput("Hello World 123", "test")
	if !allowed || violation != nil {
		log.Fatal("Expected valid input to be allowed")
	}

	// Test input too long
	longInput := "This is a very long input that exceeds the maximum allowed length and should be blocked by the security policy"
	allowed, violation = policy.ValidateInput(longInput, "test")
	if allowed || violation == nil {
		log.Fatal("Expected long input to be blocked")
	}

	// Test invalid content
	invalidInput := "Hello <script>alert('xss')</script>"
	allowed, violation = policy.ValidateInput(invalidInput, "test")
	if allowed || violation == nil {
		log.Fatal("Expected invalid content to be blocked")
	}

	fmt.Println("   ✓ Input validation works")
}