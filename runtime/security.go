package runtime

import (
	"fmt"
	"html"
	"regexp"
	"sync"
	"time"
)

// SecurityContext manages security context during template execution
type SecurityContext struct {
	policy         *SecurityPolicy
	violations     []*SecurityViolation
	auditLog       []*SecurityAuditEntry
	executionStart time.Time
	recursionDepth int
	memoryUsage    int64
	outputSize     int64
	templatesUsed  map[string]bool
	sessionID      string
	mu             sync.RWMutex
}

// SecurityAuditEntry represents an audit log entry
type SecurityAuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Operation   string    `json:"operation"`
	Resource    string    `json:"resource"`
	Allowed     bool      `json:"allowed"`
	Context     string    `json:"context"`
	Template    string    `json:"template"`
	Description string    `json:"description"`
}

// SecurityManager manages security policies and enforcement
type SecurityManager struct {
	policies       map[string]*SecurityPolicy
	defaultPolicy  *SecurityPolicy
	activeSessions map[string]*SecurityContext
	mu             sync.RWMutex
}

// NewSecurityManager creates a new security manager
func NewSecurityManager() *SecurityManager {
	sm := &SecurityManager{
		policies:       make(map[string]*SecurityPolicy),
		activeSessions: make(map[string]*SecurityContext),
	}

	// Create and set default policy
	sm.defaultPolicy = DefaultSecurityPolicy()
	sm.policies["default"] = sm.defaultPolicy

	return sm
}

// DefaultSecurityPolicy returns a secure default policy for production use
func DefaultSecurityPolicy() *SecurityPolicy {
	return NewSecurityPolicyBuilder("default", "Secure default policy for production use").
		SetLevel(SecurityLevelProduction).
		// Allow safe filters only
		AllowFilters("upper", "lower", "title", "capitalize", "trim", "length", "first", "last", "join", "replace", "escape", "safe").
		// Allow safe functions only
		AllowFunctions("range", "dict", "cycler", "joiner").
		AllowTests(
			"divisibleby", "defined", "undefined", "none", "null", "boolean", "true", "false", "number", "integer", "float",
			"string", "sequence", "mapping", "iterable", "callable", "sameas", "escaped", "module", "list", "tuple", "dict",
			"lower", "upper", "even", "odd", "in", "filter", "test", "equalto", "==", "!=", "eq", "ne", "lt", "le", "gt", "ge",
			">", "<", ">=", "<=", "greaterthan", "lessthan", "matching", "search", "startingwith", "endingwith", "containing",
			"infinite", "nan", "finite",
		).
		// Allow safe attribute patterns
		AllowAttributePattern("^user\\.(name|email|avatar)$").
		AllowAttributePattern("^config\\.(theme|language)$").
		AllowAttributePattern("^[a-z_]+$"). // Only simple attributes
		// Block dangerous method calls
		BlockAllMethodCalls().
		// Set resource limits
		SetMaxExecutionTime(10 * time.Second).
		SetMaxRecursionDepth(50).
		SetMaxMemoryUsage(10 * 1024 * 1024). // 10MB
		SetMaxOutputSize(1024 * 1024).       // 1MB
		// Content restrictions
		AllowHTMLContent(false).
		AllowJavaScriptContent(false).
		AllowCSSContent(false).
		// Security options
		EnableAuditLogging(true).
		BlockOnViolation(true).
		AutoEscapeOutput(true).
		ValidateAllInputs(true).
		SetMaxInputLength(100 * 1024). // 100KB
		Build()
}

// DevelopmentSecurityPolicy returns a permissive policy for development
func DevelopmentSecurityPolicy() *SecurityPolicy {
	return NewSecurityPolicyBuilder("development", "Permissive policy for development use").
		SetLevel(SecurityLevelDevelopment).
		// Use blacklist mode for filters (allow all except blocked)
		SetFilterWhitelistMode(false).
		SetFunctionWhitelistMode(false).
		SetTestWhitelistMode(false).
		SetAttributeWhitelistMode(false).
		SetMethodWhitelistMode(false).
		SetTemplateWhitelistMode(false).
		// Block dangerous filters explicitly
		BlockFilters("eval", "exec").
		// Block dangerous functions explicitly
		BlockFunctions("open", "exec", "eval").
		// Allow broader attribute access in development
		AllowAttributePattern(".*").
		// Allow some method calls in development
		AllowMethods("String", "Len", "Int", "Float").
		// Higher limits for development
		SetMaxExecutionTime(60 * time.Second).
		SetMaxRecursionDepth(200).
		SetMaxMemoryUsage(100 * 1024 * 1024). // 100MB
		SetMaxOutputSize(10 * 1024 * 1024).   // 10MB
		// More permissive content
		AllowHTMLContent(true).
		AllowJavaScriptContent(false).
		AllowCSSContent(true).
		// Development options
		EnableAuditLogging(true).
		BlockOnViolation(false).
		AutoEscapeOutput(false).
		ValidateAllInputs(true).
		SetMaxInputLength(1024 * 1024). // 1MB
		Build()
}

// RestrictedSecurityPolicy returns a highly restrictive policy for untrusted templates
func RestrictedSecurityPolicy() *SecurityPolicy {
	return NewSecurityPolicyBuilder("restricted", "Highly restrictive policy for untrusted templates").
		SetLevel(SecurityLevelRestricted).
		// Only allow basic text manipulation filters
		AllowFilters("upper", "lower", "trim", "escape").
		AllowFunctions("range").
		AllowTests(
			"divisibleby", "defined", "undefined", "none", "null", "boolean", "true", "false", "number", "integer", "float",
			"string", "sequence", "mapping", "iterable", "callable", "sameas", "escaped", "module", "list", "tuple", "dict",
			"lower", "upper", "even", "odd", "in", "filter", "test", "equalto", "==", "!=", "eq", "ne", "lt", "le", "gt", "ge",
			">", "<", ">=", "<=", "greaterthan", "lessthan", "matching", "search", "startingwith", "endingwith", "containing",
			"infinite", "nan", "finite",
		).
		// Very restrictive attribute access
		AllowAttributes("value", "text", "content").
		// Block all method calls
		BlockAllMethodCalls().
		// Very strict limits
		SetMaxExecutionTime(2 * time.Second).
		SetMaxRecursionDepth(10).
		SetMaxMemoryUsage(1024 * 1024). // 1MB
		SetMaxOutputSize(10 * 1024).    // 10KB
		// No content allowed
		AllowHTMLContent(false).
		AllowJavaScriptContent(false).
		AllowCSSContent(false).
		// Maximum security
		EnableAuditLogging(true).
		BlockOnViolation(true).
		AutoEscapeOutput(true).
		ValidateAllInputs(true).
		SetMaxInputLength(1024). // 1KB
		Build()
}

// AddPolicy adds a security policy to the manager
func (sm *SecurityManager) AddPolicy(name string, policy *SecurityPolicy) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if policy == nil {
		return fmt.Errorf("policy cannot be nil")
	}

	sm.policies[name] = policy
	return nil
}

// GetPolicy retrieves a security policy by name
func (sm *SecurityManager) GetPolicy(name string) (*SecurityPolicy, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	policy, ok := sm.policies[name]
	if !ok {
		return nil, fmt.Errorf("security policy '%s' not found", name)
	}

	return policy, nil
}

// RemovePolicy removes a security policy
func (sm *SecurityManager) RemovePolicy(name string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if name == "default" {
		return fmt.Errorf("cannot remove default security policy")
	}

	delete(sm.policies, name)
	return nil
}

// ListPolicies returns a list of all available policies
func (sm *SecurityManager) ListPolicies() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	policies := make([]string, 0, len(sm.policies))
	for name := range sm.policies {
		policies = append(policies, name)
	}

	return policies
}

// CreateSecurityContext creates a new security context for template execution
func (sm *SecurityManager) CreateSecurityContext(policyName, templateName string) (*SecurityContext, error) {
	policy, err := sm.GetPolicy(policyName)
	if err != nil {
		return nil, err
	}

	sessionID := fmt.Sprintf("%s_%d", templateName, time.Now().UnixNano())
	ctx := &SecurityContext{
		policy:         policy,
		violations:     make([]*SecurityViolation, 0),
		auditLog:       make([]*SecurityAuditEntry, 0),
		executionStart: time.Now(),
		templatesUsed:  make(map[string]bool),
		sessionID:      sessionID,
	}

	// Add to active sessions
	sm.mu.Lock()
	sm.activeSessions[sessionID] = ctx
	sm.mu.Unlock()

	return ctx, nil
}

// CleanupSecurityContext removes a security context from active sessions
func (sm *SecurityManager) CleanupSecurityContext(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.activeSessions, sessionID)
}

// GetActiveSessions returns the number of active security sessions
func (sm *SecurityManager) GetActiveSessions() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.activeSessions)
}

// SecurityContext methods

// GetPolicy returns the security policy for this context
func (sc *SecurityContext) GetPolicy() *SecurityPolicy {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.policy
}

// CheckFilterAccess checks if filter access is allowed
func (sc *SecurityContext) CheckFilterAccess(filterName, templateName, context string) bool {
	allowed, violation := sc.policy.IsFilterAllowed(filterName)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "filter_access",
		Resource:    filterName,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Filter access: %s", filterName),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckFunctionAccess checks if function access is allowed
func (sc *SecurityContext) CheckFunctionAccess(functionName, templateName, context string) bool {
	allowed, violation := sc.policy.IsFunctionAllowed(functionName)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "function_access",
		Resource:    functionName,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Function access: %s", functionName),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckTestAccess checks if test access is allowed
func (sc *SecurityContext) CheckTestAccess(testName, templateName, context string) bool {
	allowed, violation := sc.policy.IsTestAllowed(testName)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "test_access",
		Resource:    testName,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Test access: %s", testName),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckAttributeAccess checks if attribute access is allowed
func (sc *SecurityContext) CheckAttributeAccess(attributePath, templateName, context string) bool {
	allowed, violation := sc.policy.IsAttributeAllowed(attributePath)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "attribute_access",
		Resource:    attributePath,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Attribute access: %s", attributePath),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckMethodCall checks if method call is allowed
func (sc *SecurityContext) CheckMethodCall(methodName, templateName, context string) bool {
	allowed, violation := sc.policy.IsMethodCallAllowed(methodName)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "method_call",
		Resource:    methodName,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Method call: %s", methodName),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckTemplateAccess checks if template access is allowed
func (sc *SecurityContext) CheckTemplateAccess(templateName, context string) bool {
	allowed, violation := sc.policy.IsTemplateAllowed(templateName)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "template_access",
		Resource:    templateName,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Template access: %s", templateName),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	if allowed {
		sc.mu.Lock()
		sc.templatesUsed[templateName] = true
		sc.mu.Unlock()
	}

	return allowed || !sc.policy.BlockOnViolation
}

// ValidateInput validates input against the security policy
func (sc *SecurityContext) ValidateInput(input, inputType, templateName, context string) bool {
	allowed, violation := sc.policy.ValidateInput(input, inputType)

	entry := &SecurityAuditEntry{
		Timestamp:   time.Now(),
		Operation:   "input_validation",
		Resource:    inputType,
		Allowed:     allowed,
		Context:     context,
		Template:    templateName,
		Description: fmt.Sprintf("Input validation: %s", inputType),
	}

	if violation != nil {
		entry.Description = violation.Description
		violation.Template = templateName
		sc.addViolation(violation)
	}

	sc.addAuditEntry(entry)

	return allowed || !sc.policy.BlockOnViolation
}

// CheckRecursionLimit checks if recursion limit is exceeded
func (sc *SecurityContext) CheckRecursionLimit(templateName string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.recursionDepth++

	if sc.recursionDepth > sc.policy.MaxRecursionDepth {
		violation := &SecurityViolation{
			Type:        ViolationTypeRecursionLimit,
			Description: fmt.Sprintf("Recursion depth %d exceeds limit %d", sc.recursionDepth, sc.policy.MaxRecursionDepth),
			Context:     fmt.Sprintf("depth: %d", sc.recursionDepth),
			Template:    templateName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sc.policy.BlockOnViolation,
		}

		sc.addViolationLocked(violation)

		entry := &SecurityAuditEntry{
			Timestamp:   time.Now(),
			Operation:   "recursion_check",
			Resource:    "recursion_depth",
			Allowed:     false,
			Context:     fmt.Sprintf("depth: %d", sc.recursionDepth),
			Template:    templateName,
			Description: violation.Description,
		}
		sc.addAuditEntryLocked(entry)

		return !sc.policy.BlockOnViolation
	}

	return true
}

// CheckExecutionTime checks if execution time limit is exceeded
func (sc *SecurityContext) CheckExecutionTime(templateName string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	executionTime := time.Since(sc.executionStart)

	if executionTime > sc.policy.MaxExecutionTime {
		violation := &SecurityViolation{
			Type:        ViolationTypeExecutionTimeout,
			Description: fmt.Sprintf("Execution time %s exceeds limit %s", executionTime, sc.policy.MaxExecutionTime),
			Context:     fmt.Sprintf("time: %s", executionTime),
			Template:    templateName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sc.policy.BlockOnViolation,
		}

		sc.addViolation(violation)

		entry := &SecurityAuditEntry{
			Timestamp:   time.Now(),
			Operation:   "execution_time_check",
			Resource:    "execution_time",
			Allowed:     false,
			Context:     fmt.Sprintf("time: %s", executionTime),
			Template:    templateName,
			Description: violation.Description,
		}
		sc.addAuditEntry(entry)

		return !sc.policy.BlockOnViolation
	}

	return true
}

// UpdateMemoryUsage updates the memory usage tracking
func (sc *SecurityContext) UpdateMemoryUsage(bytes int64, templateName string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.memoryUsage += bytes

	if sc.memoryUsage > sc.policy.MaxMemoryUsage {
		violation := &SecurityViolation{
			Type:        ViolationTypeMemoryLimit,
			Description: fmt.Sprintf("Memory usage %d bytes exceeds limit %d bytes", sc.memoryUsage, sc.policy.MaxMemoryUsage),
			Context:     fmt.Sprintf("memory: %d bytes", sc.memoryUsage),
			Template:    templateName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sc.policy.BlockOnViolation,
		}

		sc.addViolationLocked(violation)

		entry := &SecurityAuditEntry{
			Timestamp:   time.Now(),
			Operation:   "memory_usage_check",
			Resource:    "memory_usage",
			Allowed:     false,
			Context:     fmt.Sprintf("memory: %d bytes", sc.memoryUsage),
			Template:    templateName,
			Description: violation.Description,
		}
		sc.addAuditEntryLocked(entry)

		return !sc.policy.BlockOnViolation
	}

	return true
}

// UpdateOutputSize updates the output size tracking
func (sc *SecurityContext) UpdateOutputSize(bytes int64, templateName string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.outputSize += bytes

	if sc.outputSize > sc.policy.MaxOutputSize {
		violation := &SecurityViolation{
			Type:        ViolationTypeMemoryLimit,
			Description: fmt.Sprintf("Output size %d bytes exceeds limit %d bytes", sc.outputSize, sc.policy.MaxOutputSize),
			Context:     fmt.Sprintf("output: %d bytes", sc.outputSize),
			Template:    templateName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sc.policy.BlockOnViolation,
		}

		sc.addViolationLocked(violation)

		entry := &SecurityAuditEntry{
			Timestamp:   time.Now(),
			Operation:   "output_size_check",
			Resource:    "output_size",
			Allowed:     false,
			Context:     fmt.Sprintf("output: %d bytes", sc.outputSize),
			Template:    templateName,
			Description: violation.Description,
		}
		sc.addAuditEntryLocked(entry)

		return !sc.policy.BlockOnViolation
	}

	return true
}

// SanitizeOutput sanitizes output according to the security policy
func (sc *SecurityContext) SanitizeOutput(output string, templateName string) string {
	if sc.policy.EscapeOutput {
		// HTML escape the output
		output = html.EscapeString(output)
	}

	// Check for restricted content patterns
	for _, pattern := range sc.policy.RestrictedContentPatterns {
		if pattern.MatchString(output) {
			violation := &SecurityViolation{
				Type:        ViolationTypeRestrictedContent,
				Description: fmt.Sprintf("Output contains restricted content matching pattern: %s", pattern.String()),
				Context:     "output_sanitization",
				Template:    templateName,
				Timestamp:   time.Now(),
				Severity:    "high",
				Blocked:     sc.policy.BlockOnViolation,
			}

			sc.addViolation(violation)

			entry := &SecurityAuditEntry{
				Timestamp:   time.Now(),
				Operation:   "output_sanitization",
				Resource:    "restricted_content",
				Allowed:     false,
				Context:     "output_sanitization",
				Template:    templateName,
				Description: violation.Description,
			}
			sc.addAuditEntry(entry)

			if sc.policy.BlockOnViolation {
				return "" // Block the output
			}
		}
	}

	return output
}

// GetViolations returns all security violations
func (sc *SecurityContext) GetViolations() []*SecurityViolation {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	violations := make([]*SecurityViolation, len(sc.violations))
	copy(violations, sc.violations)

	return violations
}

// GetAuditLog returns the audit log
func (sc *SecurityContext) GetAuditLog() []*SecurityAuditEntry {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	log := make([]*SecurityAuditEntry, len(sc.auditLog))
	copy(log, sc.auditLog)

	return log
}

// GetExecutionStats returns execution statistics
func (sc *SecurityContext) GetExecutionStats() map[string]interface{} {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return map[string]interface{}{
		"execution_time":   time.Since(sc.executionStart),
		"recursion_depth":  sc.recursionDepth,
		"memory_usage":     sc.memoryUsage,
		"output_size":      sc.outputSize,
		"templates_used":   len(sc.templatesUsed),
		"violations_count": len(sc.violations),
		"audit_entries":    len(sc.auditLog),
	}
}

// HasViolations returns true if there are any security violations
func (sc *SecurityContext) HasViolations() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return len(sc.violations) > 0
}

// HasBlockedViolations returns true if there are any blocked violations
func (sc *SecurityContext) HasBlockedViolations() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	for _, violation := range sc.violations {
		if violation.Blocked {
			return true
		}
	}

	return false
}

// Reset resets the security context for reuse
func (sc *SecurityContext) Reset() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.violations = make([]*SecurityViolation, 0)
	sc.auditLog = make([]*SecurityAuditEntry, 0)
	sc.executionStart = time.Now()
	sc.recursionDepth = 0
	sc.memoryUsage = 0
	sc.outputSize = 0
	sc.templatesUsed = make(map[string]bool)
}

// Internal methods

func (sc *SecurityContext) addViolation(violation *SecurityViolation) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.violations = append(sc.violations, violation)
}

func (sc *SecurityContext) addViolationLocked(violation *SecurityViolation) {
	// This method assumes the mutex is already locked
	sc.violations = append(sc.violations, violation)
}

func (sc *SecurityContext) addAuditEntry(entry *SecurityAuditEntry) {
	if !sc.policy.EnableAuditLogging {
		return
	}

	if !sc.policy.LogAllowedOperations && entry.Allowed {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.auditLog = append(sc.auditLog, entry)
}

func (sc *SecurityContext) addAuditEntryLocked(entry *SecurityAuditEntry) {
	// This method assumes the mutex is already locked
	if !sc.policy.EnableAuditLogging {
		return
	}

	if !sc.policy.LogAllowedOperations && entry.Allowed {
		return
	}

	sc.auditLog = append(sc.auditLog, entry)
}

// Global security manager instance
var globalSecurityManager = NewSecurityManager()

// GetGlobalSecurityManager returns the global security manager
func GetGlobalSecurityManager() *SecurityManager {
	return globalSecurityManager
}

// CreateSecureContext creates a secure context using the global security manager
func CreateSecureContext(policyName, templateName string) (*SecurityContext, error) {
	return globalSecurityManager.CreateSecurityContext(policyName, templateName)
}

// Predefined dangerous patterns for content filtering
var (
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)data:text/html`),
		regexp.MustCompile(`(?i)on\w+\s*=`),
		regexp.MustCompile(`(?i)<iframe[^>]*>`),
		regexp.MustCompile(`(?i)<object[^>]*>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)<form[^>]*action\s*=\s*["']?javascript:`),
		regexp.MustCompile(`(?i)<meta[^>]*http-equiv`),
	}
)

// AddDangerousPatterns adds dangerous content patterns to a policy
func AddDangerousPatterns(policy *SecurityPolicy) {
	for _, pattern := range dangerousPatterns {
		policy.RestrictedContentPatterns = append(policy.RestrictedContentPatterns, pattern)
	}
}
