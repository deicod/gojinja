package runtime

import (
	"strings"
	"testing"
	"time"
)

// TestSecurityPolicyBuilder tests the security policy builder
func TestSecurityPolicyBuilder(t *testing.T) {
	// Test basic policy creation
	policy := NewSecurityPolicyBuilder("test", "Test policy").
		SetLevel(SecurityLevelProduction).
		AllowFilters("upper", "lower").
		BlockFilters("eval").
		AllowFunctions("range", "dict").
		BlockFunctions("open").
		SetMaxExecutionTime(5 * time.Second).
		SetMaxRecursionDepth(10).
		Build()

	if policy.Name != "test" {
		t.Errorf("Expected policy name 'test', got '%s'", policy.Name)
	}

	if policy.Level != SecurityLevelProduction {
		t.Errorf("Expected security level %d, got %d", SecurityLevelProduction, policy.Level)
	}

	// Test filter access
	allowed, violation := policy.IsFilterAllowed("upper")
	if !allowed || violation != nil {
		t.Errorf("Expected 'upper' filter to be allowed")
	}

	allowed, violation = policy.IsFilterAllowed("eval")
	if allowed || violation == nil {
		t.Errorf("Expected 'eval' filter to be blocked")
	}

	// Test function access
	allowed, violation = policy.IsFunctionAllowed("range")
	if !allowed || violation != nil {
		t.Errorf("Expected 'range' function to be allowed")
	}

	allowed, violation = policy.IsFunctionAllowed("open")
	if allowed || violation == nil {
		t.Errorf("Expected 'open' function to be blocked")
	}

	// Test limits
	if policy.MaxExecutionTime != 5*time.Second {
		t.Errorf("Expected max execution time 5s, got %v", policy.MaxExecutionTime)
	}

	if policy.MaxRecursionDepth != 10 {
		t.Errorf("Expected max recursion depth 10, got %d", policy.MaxRecursionDepth)
	}
}

// TestDefaultSecurityPolicies tests the predefined security policies
func TestDefaultSecurityPolicies(t *testing.T) {
	// Test default policy
	defaultPolicy := DefaultSecurityPolicy()
	if defaultPolicy.FilterWhitelist != true {
		t.Errorf("Default policy should use whitelist mode for filters")
	}

	if defaultPolicy.BlockOnViolation != true {
		t.Errorf("Default policy should block on violations")
	}

	// Test development policy
	devPolicy := DevelopmentSecurityPolicy()
	if devPolicy.FilterWhitelist != false {
		t.Errorf("Development policy should not use whitelist mode for filters")
	}

	if devPolicy.BlockOnViolation != false {
		t.Errorf("Development policy should not block on violations")
	}

	// Test restricted policy
	restrictedPolicy := RestrictedSecurityPolicy()
	if restrictedPolicy.BlockAllMethods != true {
		t.Errorf("Restricted policy should block all methods")
	}

	if restrictedPolicy.MaxRecursionDepth != 10 {
		t.Errorf("Restricted policy should have max recursion depth of 10")
	}
}

// TestSecurityManager tests the security manager
func TestSecurityManager(t *testing.T) {
	manager := NewSecurityManager()

	// Test adding and getting policies
	customPolicy := NewSecurityPolicyBuilder("custom", "Custom policy").
		AllowFilters("test").
		Build()

	err := manager.AddPolicy("custom", customPolicy)
	if err != nil {
		t.Errorf("Failed to add custom policy: %v", err)
	}

	retrievedPolicy, err := manager.GetPolicy("custom")
	if err != nil {
		t.Errorf("Failed to get custom policy: %v", err)
	}

	if retrievedPolicy.Name != "custom" {
		t.Errorf("Expected policy name 'custom', got '%s'", retrievedPolicy.Name)
	}

	// Test security context creation
	secCtx, err := manager.CreateSecurityContext("custom", "test_template")
	if err != nil {
		t.Errorf("Failed to create security context: %v", err)
	}

	if secCtx == nil {
		t.Error("Security context should not be nil")
	}

	// Test filter access through security context
	allowed := secCtx.CheckFilterAccess("test", "test_template", "test_context")
	if !allowed {
		t.Errorf("Expected 'test' filter to be allowed")
	}

	blocked := secCtx.CheckFilterAccess("blocked", "test_template", "test_context")
	if blocked {
		t.Errorf("Expected 'blocked' filter to be blocked")
	}

	// Test violations
	violations := secCtx.GetViolations()
	if len(violations) == 0 {
		t.Errorf("Expected at least one security violation")
	}

	// Cleanup
	manager.CleanupSecurityContext("test_template_123")
}

// TestSandboxEnvironment tests the sandbox environment
func TestSandboxEnvironment(t *testing.T) {
	// Create sandbox environment
	sandbox := NewSecureEnvironment()

	if sandbox.GetSecurityPolicy() != "default" {
		t.Errorf("Expected default security policy")
	}

	// Test setting custom policy
	customPolicy := NewSecurityPolicyBuilder("test", "Test policy").
		AllowFilters("upper").
		Build()

	sandbox.securityManager.AddPolicy("test", customPolicy)
	err := sandbox.SetSecurityPolicy("test")
	if err != nil {
		t.Errorf("Failed to set security policy: %v", err)
	}

	if sandbox.GetSecurityPolicy() != "test" {
		t.Errorf("Expected security policy 'test'")
	}
}

// TestInputValidation tests input validation
func TestInputValidation(t *testing.T) {
	policy := NewSecurityPolicyBuilder("test", "Test policy").
		SetMaxInputLength(100).
		AllowRestrictedContentPattern(`^[a-zA-Z0-9\s]+$`).
		Build()

	// Test valid input
	allowed, violation := policy.ValidateInput("Hello World 123", "test")
	if !allowed || violation != nil {
		t.Errorf("Expected valid input to be allowed")
	}

	// Test input too long
	longInput := strings.Repeat("a", 200)
	allowed, violation = policy.ValidateInput(longInput, "test")
	if allowed || violation == nil {
		t.Errorf("Expected long input to be blocked")
	}

	if violation.Type != ViolationTypeInputValidation {
		t.Errorf("Expected input validation violation")
	}
}

// TestResourceLimits tests resource limits
func TestResourceLimits(t *testing.T) {
	manager := NewSecurityManager()
	secCtx, err := manager.CreateSecurityContext("default", "test")
	if err != nil {
		t.Fatalf("Failed to create security context: %v", err)
	}

	// Test recursion limit
	for i := 0; i < 150; i++ { // Exceed default limit of 100
		secCtx.CheckRecursionLimit("test")
	}

	violations := secCtx.GetViolations()
	foundRecursionViolation := false
	for _, violation := range violations {
		if violation.Type == ViolationTypeRecursionLimit {
			foundRecursionViolation = true
			break
		}
	}

	if !foundRecursionViolation {
		t.Errorf("Expected recursion limit violation")
	}

	// Test memory limit
	secCtx.UpdateMemoryUsage(200*1024*1024, "test") // Exceed default limit of 50MB

	violations = secCtx.GetViolations()
	foundMemoryViolation := false
	for _, violation := range violations {
		if violation.Type == ViolationTypeMemoryLimit {
			foundMemoryViolation = true
			break
		}
	}

	if !foundMemoryViolation {
		t.Errorf("Expected memory limit violation")
	}
}

// TestAuditLogging tests audit logging
func TestAuditLogging(t *testing.T) {
	// Create memory audit logger
	logger := NewMemoryAuditLogger(1000)

	// Create audit manager
	auditManager := NewAuditManager(logger)

	// Test logging events
	violation := &SecurityViolation{
		Type:        ViolationTypeFilterAccess,
		Description: "Test violation",
		Context:     "test",
		Timestamp:   time.Now(),
		Severity:    "high",
		Blocked:     true,
	}

	auditManager.LogSecurityViolation(violation, "test_template", "test_context")

	// Check logged events
	events := logger.GetEvents()
	if len(events) == 0 {
		t.Errorf("Expected at least one audit event")
		return
	}

	event := events[0]
	if event.Type != AuditEventSecurityViolation {
		t.Errorf("Expected security violation event type")
	}

	if event.Template != "test_template" {
		t.Errorf("Expected template name 'test_template'")
	}
}

// TestAttributeAccessPatterns tests attribute access patterns
func TestAttributeAccessPatterns(t *testing.T) {
	policy := NewSecurityPolicyBuilder("test", "Test policy").
		AllowAttributePattern("^user\\.(name|email)$").
		AllowAttributePattern("^config\\.(theme|language)$").
		Build()

	// Test allowed attributes
	testCases := []struct {
		attribute string
		allowed   bool
	}{
		{"user.name", true},
		{"user.email", true},
		{"config.theme", true},
		{"config.language", true},
		{"user.password", false},
		{"config.secret", false},
		{"system.admin", false},
	}

	for _, tc := range testCases {
		allowed, violation := policy.IsAttributeAllowed(tc.attribute)
		if allowed != tc.allowed {
			t.Errorf("Attribute '%s': expected allowed=%v, got allowed=%v", tc.attribute, tc.allowed, allowed)
		}

		if tc.allowed && violation != nil {
			t.Errorf("Attribute '%s': expected no violation, got %v", tc.attribute, violation)
		}

		if !tc.allowed && violation == nil {
			t.Errorf("Attribute '%s': expected violation, got none", tc.attribute)
		}
	}
}

// TestOutputSanitization tests output sanitization
func TestOutputSanitization(t *testing.T) {
	manager := NewSecurityManager()
	secCtx, err := manager.CreateSecurityContext("default", "test")
	if err != nil {
		t.Fatalf("Failed to create security context: %v", err)
	}

	// Test HTML escaping
	input := "<script>alert('xss')</script>"
	output := secCtx.SanitizeOutput(input, "test")

	if output == input {
		t.Errorf("Expected output to be escaped")
	}

	if strings.Contains(output, "<script>") {
		t.Errorf("Output should not contain script tags")
	}
}

// TestPolicyCloning tests policy cloning
func TestPolicyCloning(t *testing.T) {
	original := NewSecurityPolicyBuilder("original", "Original policy").
		AllowFilters("upper", "lower").
		SetMaxExecutionTime(10 * time.Second).
		Build()

	clone := original.Clone()

	// Verify clone has same properties
	if clone.Name != original.Name {
		t.Errorf("Clone should have same name")
	}

	if clone.MaxExecutionTime != original.MaxExecutionTime {
		t.Errorf("Clone should have same execution time limit")
	}

	// Modify clone
	clone.SetMaxExecutionTime(20 * time.Second)

	// Verify original is unchanged
	if original.MaxExecutionTime == clone.MaxExecutionTime {
		t.Errorf("Original should not be affected by clone modifications")
	}
}

// TestMultipleSecurityContexts tests multiple security contexts
func TestMultipleSecurityContexts(t *testing.T) {
	manager := NewSecurityManager()

	// Create multiple security contexts
	ctx1, _ := manager.CreateSecurityContext("default", "template1")
	ctx2, _ := manager.CreateSecurityContext("default", "template2")

	// Check they are independent
	ctx1.CheckFilterAccess("upper", "template1", "context1")
	ctx2.CheckFilterAccess("blocked", "template2", "context2")

	violations1 := ctx1.GetViolations()
	violations2 := ctx2.GetViolations()

	if len(violations1) == 0 && len(violations2) == 0 {
		t.Errorf("Expected violations in at least one context")
	}

	// Cleanup
	manager.CleanupSecurityContext("template1_123")
	manager.CleanupSecurityContext("template2_123")
}

// BenchmarkSecurityPolicyCheck benchmarks security policy checks
func BenchmarkSecurityPolicyCheck(b *testing.B) {
	policy := DefaultSecurityPolicy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		policy.IsFilterAllowed("upper")
		policy.IsFunctionAllowed("range")
		policy.IsAttributeAllowed("user.name")
	}
}

// BenchmarkSecurityContext benchmarks security context operations
func BenchmarkSecurityContext(b *testing.B) {
	manager := NewSecurityManager()
	secCtx, _ := manager.CreateSecurityContext("default", "benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		secCtx.CheckFilterAccess("upper", "template", "context")
		secCtx.CheckFunctionAccess("range", "template", "context")
		secCtx.CheckAttributeAccess("user.name", "template", "context")
	}
}