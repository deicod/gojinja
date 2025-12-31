package runtime

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SecurityLevel represents different security levels
type SecurityLevel int

const (
	SecurityLevelDevelopment SecurityLevel = iota
	SecurityLevelStaging
	SecurityLevelProduction
	SecurityLevelRestricted
)

// String returns the string representation of the security level
func (sl SecurityLevel) String() string {
	switch sl {
	case SecurityLevelDevelopment:
		return "development"
	case SecurityLevelStaging:
		return "staging"
	case SecurityLevelProduction:
		return "production"
	case SecurityLevelRestricted:
		return "restricted"
	default:
		return "unknown"
	}
}

// SecurityViolationType represents different types of security violations
type SecurityViolationType int

const (
	ViolationTypeFilterAccess SecurityViolationType = iota
	ViolationTypeFunctionAccess
	ViolationTypeTestAccess
	ViolationTypeAttributeAccess
	ViolationTypeMethodCall
	ViolationTypeTemplateAccess
	ViolationTypeRecursionLimit
	ViolationTypeExecutionTimeout
	ViolationTypeMemoryLimit
	ViolationTypeRestrictedContent
	ViolationTypeInputValidation
)

// String returns the string representation of the violation type
func (vt SecurityViolationType) String() string {
	switch vt {
	case ViolationTypeFilterAccess:
		return "filter_access"
	case ViolationTypeFunctionAccess:
		return "function_access"
	case ViolationTypeTestAccess:
		return "test_access"
	case ViolationTypeAttributeAccess:
		return "attribute_access"
	case ViolationTypeMethodCall:
		return "method_call"
	case ViolationTypeTemplateAccess:
		return "template_access"
	case ViolationTypeRecursionLimit:
		return "recursion_limit"
	case ViolationTypeExecutionTimeout:
		return "execution_timeout"
	case ViolationTypeMemoryLimit:
		return "memory_limit"
	case ViolationTypeRestrictedContent:
		return "restricted_content"
	case ViolationTypeInputValidation:
		return "input_validation"
	default:
		return "unknown"
	}
}

// SecurityViolation represents a security policy violation
type SecurityViolation struct {
	Type        SecurityViolationType `json:"type"`
	Description string                `json:"description"`
	Context     string                `json:"context"`
	Template    string                `json:"template"`
	Position    string                `json:"position"`
	Timestamp   time.Time             `json:"timestamp"`
	Severity    string                `json:"severity"`
	Blocked     bool                  `json:"blocked"`
}

// SecurityPolicy defines the security policy for template execution
type SecurityPolicy struct {
	// Policy metadata
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Level       SecurityLevel `json:"level"`
	Version     string        `json:"version"`
	Created     time.Time     `json:"created"`
	Updated     time.Time     `json:"updated"`

	// Filter restrictions
	AllowedFilters  map[string]bool `json:"allowed_filters"`
	BlockedFilters  map[string]bool `json:"blocked_filters"`
	FilterWhitelist bool            `json:"filter_whitelist"` // true = whitelist mode, false = blacklist mode

	// Function restrictions
	AllowedFunctions  map[string]bool `json:"allowed_functions"`
	BlockedFunctions  map[string]bool `json:"blocked_functions"`
	FunctionWhitelist bool            `json:"function_whitelist"` // true = whitelist mode, false = blacklist mode

	// Test restrictions
	AllowedTests  map[string]bool `json:"allowed_tests"`
	BlockedTests  map[string]bool `json:"blocked_tests"`
	TestWhitelist bool            `json:"test_whitelist"` // true = whitelist mode, false = blacklist mode

	// Attribute access restrictions
	AllowedAttributes  map[string]bool  `json:"allowed_attributes"`
	BlockedAttributes  map[string]bool  `json:"blocked_attributes"`
	AttributeWhitelist bool             `json:"attribute_whitelist"` // true = whitelist mode, false = blacklist mode
	AttributePatterns  []*regexp.Regexp `json:"-"`                   // Compiled regex patterns

	// Method call restrictions
	AllowedMethods  map[string]bool `json:"allowed_methods"`
	BlockedMethods  map[string]bool `json:"blocked_methods"`
	MethodWhitelist bool            `json:"method_whitelist"` // true = whitelist mode, false = blacklist mode
	BlockAllMethods bool            `json:"block_all_methods"`

	// Template access restrictions
	AllowedTemplates  map[string]bool  `json:"allowed_templates"`
	BlockedTemplates  map[string]bool  `json:"blocked_templates"`
	TemplateWhitelist bool             `json:"template_whitelist"` // true = whitelist mode, false = blacklist mode
	TemplatePatterns  []*regexp.Regexp `json:"-"`                  // Compiled regex patterns

	// Resource limits
	MaxExecutionTime  time.Duration `json:"max_execution_time"`
	MaxRecursionDepth int           `json:"max_recursion_depth"`
	MaxMemoryUsage    int64         `json:"max_memory_usage"` // in bytes
	MaxOutputSize     int64         `json:"max_output_size"`  // in bytes

	// Content restrictions
	RestrictedContentPatterns []*regexp.Regexp `json:"-"` // Compiled regex patterns
	AllowHTML                 bool             `json:"allow_html"`
	AllowJavaScript           bool             `json:"allow_javascript"`
	AllowCSS                  bool             `json:"allow_css"`

	// Input validation
	RequireInputSanitization bool     `json:"require_input_sanitization"`
	AllowedInputTypes        []string `json:"allowed_input_types"`
	MaxInputLength           int      `json:"max_input_length"`

	// Security options
	EnableAuditLogging   bool `json:"enable_audit_logging"`
	BlockOnViolation     bool `json:"block_on_violation"`
	LogAllowedOperations bool `json:"log_allowed_operations"`
	EscapeOutput         bool `json:"escape_output"`
	ValidateAllInputs    bool `json:"validate_all_inputs"`

	// Thread safety
	mu sync.RWMutex `json:"-"`
}

// SecurityPolicyBuilder provides a fluent interface for building security policies
type SecurityPolicyBuilder struct {
	policy *SecurityPolicy
}

// NewSecurityPolicyBuilder creates a new security policy builder
func NewSecurityPolicyBuilder(name, description string) *SecurityPolicyBuilder {
	now := time.Now()
	return &SecurityPolicyBuilder{
		policy: &SecurityPolicy{
			Name:        name,
			Description: description,
			Level:       SecurityLevelProduction,
			Version:     "1.0.0",
			Created:     now,
			Updated:     now,

			AllowedFilters:    make(map[string]bool),
			BlockedFilters:    make(map[string]bool),
			AllowedFunctions:  make(map[string]bool),
			BlockedFunctions:  make(map[string]bool),
			AllowedTests:      make(map[string]bool),
			BlockedTests:      make(map[string]bool),
			AllowedAttributes: make(map[string]bool),
			BlockedAttributes: make(map[string]bool),
			AllowedMethods:    make(map[string]bool),
			BlockedMethods:    make(map[string]bool),
			AllowedTemplates:  make(map[string]bool),
			BlockedTemplates:  make(map[string]bool),

			// Default to whitelist mode for maximum security
			FilterWhitelist:    true,
			FunctionWhitelist:  true,
			TestWhitelist:      true,
			AttributeWhitelist: true,
			MethodWhitelist:    true,
			TemplateWhitelist:  true,

			// Default resource limits
			MaxExecutionTime:  30 * time.Second,
			MaxRecursionDepth: 100,
			MaxMemoryUsage:    50 * 1024 * 1024, // 50MB
			MaxOutputSize:     10 * 1024 * 1024, // 10MB

			// Default security options
			EnableAuditLogging:   true,
			BlockOnViolation:     true,
			LogAllowedOperations: false,
			EscapeOutput:         true,
			ValidateAllInputs:    true,
			AllowHTML:            false,
			AllowJavaScript:      false,
			AllowCSS:             false,

			MaxInputLength: 1024 * 1024, // 1MB
		},
	}
}

// SetLevel sets the security level
func (spb *SecurityPolicyBuilder) SetLevel(level SecurityLevel) *SecurityPolicyBuilder {
	spb.policy.Level = level
	spb.policy.Updated = time.Now()
	return spb
}

// AllowFilters adds filters to the allowed list
func (spb *SecurityPolicyBuilder) AllowFilters(filters ...string) *SecurityPolicyBuilder {
	for _, filter := range filters {
		spb.policy.AllowedFilters[filter] = true
		delete(spb.policy.BlockedFilters, filter)
	}
	return spb
}

// BlockFilters adds filters to the blocked list
func (spb *SecurityPolicyBuilder) BlockFilters(filters ...string) *SecurityPolicyBuilder {
	for _, filter := range filters {
		spb.policy.BlockedFilters[filter] = true
		delete(spb.policy.AllowedFilters, filter)
	}
	return spb
}

// AllowFunctions adds functions to the allowed list
func (spb *SecurityPolicyBuilder) AllowFunctions(functions ...string) *SecurityPolicyBuilder {
	for _, function := range functions {
		spb.policy.AllowedFunctions[function] = true
		delete(spb.policy.BlockedFunctions, function)
	}
	return spb
}

// BlockFunctions adds functions to the blocked list
func (spb *SecurityPolicyBuilder) BlockFunctions(functions ...string) *SecurityPolicyBuilder {
	for _, function := range functions {
		spb.policy.BlockedFunctions[function] = true
		delete(spb.policy.AllowedFunctions, function)
	}
	return spb
}

// AllowTests adds tests to the allowed list
func (spb *SecurityPolicyBuilder) AllowTests(tests ...string) *SecurityPolicyBuilder {
	for _, test := range tests {
		spb.policy.AllowedTests[test] = true
		delete(spb.policy.BlockedTests, test)
	}
	return spb
}

// BlockTests adds tests to the blocked list
func (spb *SecurityPolicyBuilder) BlockTests(tests ...string) *SecurityPolicyBuilder {
	for _, test := range tests {
		spb.policy.BlockedTests[test] = true
		delete(spb.policy.AllowedTests, test)
	}
	return spb
}

// AllowAttributes adds attributes to the allowed list
func (spb *SecurityPolicyBuilder) AllowAttributes(attributes ...string) *SecurityPolicyBuilder {
	for _, attr := range attributes {
		spb.policy.AllowedAttributes[attr] = true
		delete(spb.policy.BlockedAttributes, attr)
	}
	return spb
}

// BlockAttributes adds attributes to the blocked list
func (spb *SecurityPolicyBuilder) BlockAttributes(attributes ...string) *SecurityPolicyBuilder {
	for _, attr := range attributes {
		spb.policy.BlockedAttributes[attr] = true
		delete(spb.policy.AllowedAttributes, attr)
	}
	return spb
}

// AllowAttributePattern adds a regex pattern for allowed attributes
func (spb *SecurityPolicyBuilder) AllowAttributePattern(pattern string) *SecurityPolicyBuilder {
	regex, err := regexp.Compile(pattern)
	if err == nil {
		spb.policy.AttributePatterns = append(spb.policy.AttributePatterns, regex)
	}
	return spb
}

// AllowMethods adds methods to the allowed list
func (spb *SecurityPolicyBuilder) AllowMethods(methods ...string) *SecurityPolicyBuilder {
	for _, method := range methods {
		spb.policy.AllowedMethods[method] = true
		delete(spb.policy.BlockedMethods, method)
	}
	return spb
}

// BlockMethods adds methods to the blocked list
func (spb *SecurityPolicyBuilder) BlockMethods(methods ...string) *SecurityPolicyBuilder {
	for _, method := range methods {
		spb.policy.BlockedMethods[method] = true
		delete(spb.policy.AllowedMethods, method)
	}
	return spb
}

// BlockAllMethodCalls blocks all method calls
func (spb *SecurityPolicyBuilder) BlockAllMethodCalls() *SecurityPolicyBuilder {
	spb.policy.BlockAllMethods = true
	return spb
}

// AllowTemplates adds templates to the allowed list
func (spb *SecurityPolicyBuilder) AllowTemplates(templates ...string) *SecurityPolicyBuilder {
	for _, template := range templates {
		spb.policy.AllowedTemplates[template] = true
		delete(spb.policy.BlockedTemplates, template)
	}
	return spb
}

// BlockTemplates adds templates to the blocked list
func (spb *SecurityPolicyBuilder) BlockTemplates(templates ...string) *SecurityPolicyBuilder {
	for _, template := range templates {
		spb.policy.BlockedTemplates[template] = true
		delete(spb.policy.AllowedTemplates, template)
	}
	return spb
}

// AllowTemplatePattern adds a regex pattern for allowed templates
func (spb *SecurityPolicyBuilder) AllowTemplatePattern(pattern string) *SecurityPolicyBuilder {
	regex, err := regexp.Compile(pattern)
	if err == nil {
		spb.policy.TemplatePatterns = append(spb.policy.TemplatePatterns, regex)
	}
	return spb
}

// SetMaxExecutionTime sets the maximum execution time
func (spb *SecurityPolicyBuilder) SetMaxExecutionTime(duration time.Duration) *SecurityPolicyBuilder {
	spb.policy.MaxExecutionTime = duration
	return spb
}

// SetMaxRecursionDepth sets the maximum recursion depth
func (spb *SecurityPolicyBuilder) SetMaxRecursionDepth(depth int) *SecurityPolicyBuilder {
	spb.policy.MaxRecursionDepth = depth
	return spb
}

// SetMaxMemoryUsage sets the maximum memory usage in bytes
func (spb *SecurityPolicyBuilder) SetMaxMemoryUsage(bytes int64) *SecurityPolicyBuilder {
	spb.policy.MaxMemoryUsage = bytes
	return spb
}

// SetMaxOutputSize sets the maximum output size in bytes
func (spb *SecurityPolicyBuilder) SetMaxOutputSize(bytes int64) *SecurityPolicyBuilder {
	spb.policy.MaxOutputSize = bytes
	return spb
}

// AllowRestrictedContentPattern adds a regex pattern for restricted content
func (spb *SecurityPolicyBuilder) AllowRestrictedContentPattern(pattern string) *SecurityPolicyBuilder {
	regex, err := regexp.Compile(pattern)
	if err == nil {
		spb.policy.RestrictedContentPatterns = append(spb.policy.RestrictedContentPatterns, regex)
	}
	return spb
}

// AllowHTMLContent allows HTML content in templates
func (spb *SecurityPolicyBuilder) AllowHTMLContent(allow bool) *SecurityPolicyBuilder {
	spb.policy.AllowHTML = allow
	return spb
}

// AllowJavaScriptContent allows JavaScript content in templates
func (spb *SecurityPolicyBuilder) AllowJavaScriptContent(allow bool) *SecurityPolicyBuilder {
	spb.policy.AllowJavaScript = allow
	return spb
}

// AllowCSSContent allows CSS content in templates
func (spb *SecurityPolicyBuilder) AllowCSSContent(allow bool) *SecurityPolicyBuilder {
	spb.policy.AllowCSS = allow
	return spb
}

// EnableAuditLogging enables audit logging
func (spb *SecurityPolicyBuilder) EnableAuditLogging(enable bool) *SecurityPolicyBuilder {
	spb.policy.EnableAuditLogging = enable
	return spb
}

// BlockOnViolation blocks execution on security violations
func (spb *SecurityPolicyBuilder) BlockOnViolation(block bool) *SecurityPolicyBuilder {
	spb.policy.BlockOnViolation = block
	return spb
}

// LogAllowedOperations logs allowed operations
func (spb *SecurityPolicyBuilder) LogAllowedOperations(log bool) *SecurityPolicyBuilder {
	spb.policy.LogAllowedOperations = log
	return spb
}

// AutoEscapeOutput enables automatic output escaping
func (spb *SecurityPolicyBuilder) AutoEscapeOutput(escape bool) *SecurityPolicyBuilder {
	spb.policy.EscapeOutput = escape
	return spb
}

// ValidateAllInputs enables validation of all inputs
func (spb *SecurityPolicyBuilder) ValidateAllInputs(validate bool) *SecurityPolicyBuilder {
	spb.policy.ValidateAllInputs = validate
	return spb
}

// SetMaxInputLength sets the maximum input length
func (spb *SecurityPolicyBuilder) SetMaxInputLength(length int) *SecurityPolicyBuilder {
	spb.policy.MaxInputLength = length
	return spb
}

// SetFilterWhitelistMode sets whether to use whitelist mode for filters
func (spb *SecurityPolicyBuilder) SetFilterWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.FilterWhitelist = whitelist
	return spb
}

// SetFunctionWhitelistMode sets whether to use whitelist mode for functions
func (spb *SecurityPolicyBuilder) SetFunctionWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.FunctionWhitelist = whitelist
	return spb
}

// SetTestWhitelistMode sets whether to use whitelist mode for tests
func (spb *SecurityPolicyBuilder) SetTestWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.TestWhitelist = whitelist
	return spb
}

// SetAttributeWhitelistMode sets whether to use whitelist mode for attributes
func (spb *SecurityPolicyBuilder) SetAttributeWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.AttributeWhitelist = whitelist
	return spb
}

// SetMethodWhitelistMode sets whether to use whitelist mode for methods
func (spb *SecurityPolicyBuilder) SetMethodWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.MethodWhitelist = whitelist
	return spb
}

// SetTemplateWhitelistMode sets whether to use whitelist mode for templates
func (spb *SecurityPolicyBuilder) SetTemplateWhitelistMode(whitelist bool) *SecurityPolicyBuilder {
	spb.policy.TemplateWhitelist = whitelist
	return spb
}

// Build creates the final security policy
func (spb *SecurityPolicyBuilder) Build() *SecurityPolicy {
	spb.policy.Updated = time.Now()
	return spb.policy
}

// IsFilterAllowed checks if a filter is allowed by the policy
func (sp *SecurityPolicy) IsFilterAllowed(filterName string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check blocked filters first
	if sp.BlockedFilters[filterName] {
		violation := &SecurityViolation{
			Type:        ViolationTypeFilterAccess,
			Description: fmt.Sprintf("Filter '%s' is blocked by security policy", filterName),
			Context:     filterName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if filter is allowed
	if sp.FilterWhitelist {
		if !sp.AllowedFilters[filterName] {
			violation := &SecurityViolation{
				Type:        ViolationTypeFilterAccess,
				Description: fmt.Sprintf("Filter '%s' is not in the allowed list", filterName),
				Context:     filterName,
				Timestamp:   time.Now(),
				Severity:    "medium",
				Blocked:     sp.BlockOnViolation,
			}
			return false, violation
		}
	}

	return true, nil
}

// IsFunctionAllowed checks if a function is allowed by the policy
func (sp *SecurityPolicy) IsFunctionAllowed(functionName string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check blocked functions first
	if sp.BlockedFunctions[functionName] {
		violation := &SecurityViolation{
			Type:        ViolationTypeFunctionAccess,
			Description: fmt.Sprintf("Function '%s' is blocked by security policy", functionName),
			Context:     functionName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if function is allowed
	if sp.FunctionWhitelist {
		if !sp.AllowedFunctions[functionName] {
			violation := &SecurityViolation{
				Type:        ViolationTypeFunctionAccess,
				Description: fmt.Sprintf("Function '%s' is not in the allowed list", functionName),
				Context:     functionName,
				Timestamp:   time.Now(),
				Severity:    "medium",
				Blocked:     sp.BlockOnViolation,
			}
			return false, violation
		}
	}

	return true, nil
}

// IsTestAllowed checks if a test is allowed by the policy
func (sp *SecurityPolicy) IsTestAllowed(testName string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check blocked tests first
	if sp.BlockedTests[testName] {
		violation := &SecurityViolation{
			Type:        ViolationTypeTestAccess,
			Description: fmt.Sprintf("Test '%s' is blocked by security policy", testName),
			Context:     testName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if test is allowed
	if sp.TestWhitelist {
		if !sp.AllowedTests[testName] {
			violation := &SecurityViolation{
				Type:        ViolationTypeTestAccess,
				Description: fmt.Sprintf("Test '%s' is not in the allowed list", testName),
				Context:     testName,
				Timestamp:   time.Now(),
				Severity:    "medium",
				Blocked:     sp.BlockOnViolation,
			}
			return false, violation
		}
	}

	return true, nil
}

// IsAttributeAllowed checks if attribute access is allowed by the policy
func (sp *SecurityPolicy) IsAttributeAllowed(attributePath string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check blocked attributes first
	if sp.BlockedAttributes[attributePath] {
		violation := &SecurityViolation{
			Type:        ViolationTypeAttributeAccess,
			Description: fmt.Sprintf("Attribute '%s' is blocked by security policy", attributePath),
			Context:     attributePath,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if attribute is allowed
	if sp.AttributeWhitelist {
		// Check exact match
		if !sp.AllowedAttributes[attributePath] {
			// Check regex patterns
			patternMatched := false
			for _, pattern := range sp.AttributePatterns {
				if pattern.MatchString(attributePath) {
					patternMatched = true
					break
				}
			}

			if !patternMatched {
				violation := &SecurityViolation{
					Type:        ViolationTypeAttributeAccess,
					Description: fmt.Sprintf("Attribute '%s' is not in the allowed list", attributePath),
					Context:     attributePath,
					Timestamp:   time.Now(),
					Severity:    "medium",
					Blocked:     sp.BlockOnViolation,
				}
				return false, violation
			}
		}
	}

	return true, nil
}

// IsMethodCallAllowed checks if a method call is allowed by the policy
func (sp *SecurityPolicy) IsMethodCallAllowed(methodName string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check if all methods are blocked
	if sp.BlockAllMethods {
		violation := &SecurityViolation{
			Type:        ViolationTypeMethodCall,
			Description: fmt.Sprintf("Method calls are blocked by security policy: '%s'", methodName),
			Context:     methodName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// Check blocked methods first
	if sp.BlockedMethods[methodName] {
		violation := &SecurityViolation{
			Type:        ViolationTypeMethodCall,
			Description: fmt.Sprintf("Method '%s' is blocked by security policy", methodName),
			Context:     methodName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if method is allowed
	if sp.MethodWhitelist {
		if !sp.AllowedMethods[methodName] {
			violation := &SecurityViolation{
				Type:        ViolationTypeMethodCall,
				Description: fmt.Sprintf("Method '%s' is not in the allowed list", methodName),
				Context:     methodName,
				Timestamp:   time.Now(),
				Severity:    "medium",
				Blocked:     sp.BlockOnViolation,
			}
			return false, violation
		}
	}

	return true, nil
}

// IsTemplateAllowed checks if template access is allowed by the policy
func (sp *SecurityPolicy) IsTemplateAllowed(templateName string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check blocked templates first
	if sp.BlockedTemplates[templateName] {
		violation := &SecurityViolation{
			Type:        ViolationTypeTemplateAccess,
			Description: fmt.Sprintf("Template '%s' is blocked by security policy", templateName),
			Context:     templateName,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// If whitelist mode, check if template is allowed
	if sp.TemplateWhitelist {
		// Check exact match
		if !sp.AllowedTemplates[templateName] {
			// Check regex patterns
			patternMatched := false
			for _, pattern := range sp.TemplatePatterns {
				if pattern.MatchString(templateName) {
					patternMatched = true
					break
				}
			}

			if !patternMatched {
				violation := &SecurityViolation{
					Type:        ViolationTypeTemplateAccess,
					Description: fmt.Sprintf("Template '%s' is not in the allowed list", templateName),
					Context:     templateName,
					Timestamp:   time.Now(),
					Severity:    "medium",
					Blocked:     sp.BlockOnViolation,
				}
				return false, violation
			}
		}
	}

	return true, nil
}

// ValidateInput validates input against the policy
func (sp *SecurityPolicy) ValidateInput(input string, inputType string) (bool, *SecurityViolation) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	// Check input length
	if len(input) > sp.MaxInputLength {
		violation := &SecurityViolation{
			Type:        ViolationTypeInputValidation,
			Description: fmt.Sprintf("Input length %d exceeds maximum allowed %d", len(input), sp.MaxInputLength),
			Context:     inputType,
			Timestamp:   time.Now(),
			Severity:    "medium",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	// Check restricted content patterns (these are ALLOWLIST patterns - input must match at least one)
	// The method name "AllowRestrictedContentPattern" adds patterns that define what content is ALLOWED
	if len(sp.RestrictedContentPatterns) > 0 {
		matched := false
		for _, pattern := range sp.RestrictedContentPatterns {
			if pattern.MatchString(input) {
				matched = true
				break
			}
		}
		if !matched {
			violation := &SecurityViolation{
				Type:        ViolationTypeInputValidation,
				Description: fmt.Sprintf("Input does not match any allowed content patterns"),
				Context:     inputType,
				Timestamp:   time.Now(),
				Severity:    "high",
				Blocked:     sp.BlockOnViolation,
			}
			return false, violation
		}
	}

	// Check for dangerous content if not explicitly allowed
	if !sp.AllowHTML && strings.Contains(strings.ToLower(input), "<script") {
		violation := &SecurityViolation{
			Type:        ViolationTypeRestrictedContent,
			Description: "Input contains potentially dangerous HTML/JavaScript content",
			Context:     inputType,
			Timestamp:   time.Now(),
			Severity:    "high",
			Blocked:     sp.BlockOnViolation,
		}
		return false, violation
	}

	return true, nil
}

// Clone creates a deep copy of the security policy
func (sp *SecurityPolicy) Clone() *SecurityPolicy {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	clone := &SecurityPolicy{
		Name:        sp.Name,
		Description: sp.Description,
		Level:       sp.Level,
		Version:     sp.Version,
		Created:     sp.Created,
		Updated:     time.Now(),

		AllowedFilters:    make(map[string]bool),
		BlockedFilters:    make(map[string]bool),
		AllowedFunctions:  make(map[string]bool),
		BlockedFunctions:  make(map[string]bool),
		AllowedTests:      make(map[string]bool),
		BlockedTests:      make(map[string]bool),
		AllowedAttributes: make(map[string]bool),
		BlockedAttributes: make(map[string]bool),
		AllowedMethods:    make(map[string]bool),
		BlockedMethods:    make(map[string]bool),
		AllowedTemplates:  make(map[string]bool),
		BlockedTemplates:  make(map[string]bool),

		FilterWhitelist:    sp.FilterWhitelist,
		FunctionWhitelist:  sp.FunctionWhitelist,
		TestWhitelist:      sp.TestWhitelist,
		AttributeWhitelist: sp.AttributeWhitelist,
		MethodWhitelist:    sp.MethodWhitelist,
		TemplateWhitelist:  sp.TemplateWhitelist,
		BlockAllMethods:    sp.BlockAllMethods,

		MaxExecutionTime:  sp.MaxExecutionTime,
		MaxRecursionDepth: sp.MaxRecursionDepth,
		MaxMemoryUsage:    sp.MaxMemoryUsage,
		MaxOutputSize:     sp.MaxOutputSize,

		AllowHTML:       sp.AllowHTML,
		AllowJavaScript: sp.AllowJavaScript,
		AllowCSS:        sp.AllowCSS,

		EnableAuditLogging:   sp.EnableAuditLogging,
		BlockOnViolation:     sp.BlockOnViolation,
		LogAllowedOperations: sp.LogAllowedOperations,
		EscapeOutput:         sp.EscapeOutput,
		ValidateAllInputs:    sp.ValidateAllInputs,

		MaxInputLength: sp.MaxInputLength,
	}

	// Copy maps
	for k, v := range sp.AllowedFilters {
		clone.AllowedFilters[k] = v
	}
	for k, v := range sp.BlockedFilters {
		clone.BlockedFilters[k] = v
	}
	for k, v := range sp.AllowedFunctions {
		clone.AllowedFunctions[k] = v
	}
	for k, v := range sp.BlockedFunctions {
		clone.BlockedFunctions[k] = v
	}
	for k, v := range sp.AllowedTests {
		clone.AllowedTests[k] = v
	}
	for k, v := range sp.BlockedTests {
		clone.BlockedTests[k] = v
	}
	for k, v := range sp.AllowedAttributes {
		clone.AllowedAttributes[k] = v
	}
	for k, v := range sp.BlockedAttributes {
		clone.BlockedAttributes[k] = v
	}
	for k, v := range sp.AllowedMethods {
		clone.AllowedMethods[k] = v
	}
	for k, v := range sp.BlockedMethods {
		clone.BlockedMethods[k] = v
	}
	for k, v := range sp.AllowedTemplates {
		clone.AllowedTemplates[k] = v
	}
	for k, v := range sp.BlockedTemplates {
		clone.BlockedTemplates[k] = v
	}

	// Copy regex patterns
	clone.AttributePatterns = make([]*regexp.Regexp, len(sp.AttributePatterns))
	copy(clone.AttributePatterns, sp.AttributePatterns)
	clone.TemplatePatterns = make([]*regexp.Regexp, len(sp.TemplatePatterns))
	copy(clone.TemplatePatterns, sp.TemplatePatterns)
	clone.RestrictedContentPatterns = make([]*regexp.Regexp, len(sp.RestrictedContentPatterns))
	copy(clone.RestrictedContentPatterns, sp.RestrictedContentPatterns)

	return clone
}

// SetMaxExecutionTime sets the maximum execution time for the security policy
func (sp *SecurityPolicy) SetMaxExecutionTime(duration time.Duration) {
	sp.mu.Lock()
	defer sp.mu.Unlock()
	sp.MaxExecutionTime = duration
	sp.Updated = time.Now()
}
