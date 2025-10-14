package runtime

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/parser"
)

// FilterFunc represents a filter function
type FilterFunc func(ctx *Context, value interface{}, args ...interface{}) (interface{}, error)

// TestFunc represents a test function
type TestFunc func(ctx *Context, value interface{}, args ...interface{}) (bool, error)

// GlobalFunc represents a global function
type GlobalFunc func(ctx *Context, args ...interface{}) (interface{}, error)

// FinalizeFunc represents a finalize callable invoked before output
type FinalizeFunc func(value interface{}) (interface{}, error)

// UndefinedFactory creates undefined values based on name
type UndefinedFactory func(name string) undefinedType

// Loader represents a template loader interface
type Loader interface {
	Load(name string) (string, error)
}

// FileSystemLoader loads templates from the file system
type FileSystemLoader struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileSystemLoader creates a new file system loader
func NewFileSystemLoader(basePath string) *FileSystemLoader {
	return &FileSystemLoader{
		basePath: basePath,
	}
}

// Load loads a template from the file system
func (l *FileSystemLoader) Load(name string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	fullPath := filepath.Join(l.basePath, name)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", name, err)
	}
	return string(data), nil
}

// MapLoader loads templates from a map
type MapLoader struct {
	templates map[string]string
	mu        sync.RWMutex
}

// NewMapLoader creates a new map loader
func NewMapLoader(templates map[string]string) *MapLoader {
	return &MapLoader{
		templates: templates,
	}
}

// Load loads a template from the map
func (l *MapLoader) Load(name string) (string, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	template, ok := l.templates[name]
	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}
	return template, nil
}

// Autoescape values
const (
	AutoescapeTrue    = true
	AutoescapeFalse   = false
	AutoescapeDefault = "default"
)

// Environment configuration
type Environment struct {
	// Template loading
	loader              Loader
	autoescape          interface{}
	cacheSize           int
	trimBlocks          bool
	lstripBlocks        bool
	keepTrailingNewline bool
	newlineSequence     string
	finalize            FinalizeFunc
	undefinedFactory    UndefinedFactory

	// Extensions
	extensions []string
	policies   map[string]interface{}

	// Security
	sandboxed       bool
	secureDefaults  bool
	securityPolicy  *SecurityPolicy
	securityManager *SecurityManager

	// Built-ins
	filters map[string]FilterFunc
	tests   map[string]TestFunc
	globals map[string]GlobalFunc

	// Runtime state
	compiledTemplates map[string]*Template
	cache             *TemplateCache
	macroRegistry     *MacroRegistry
	urlFor            GlobalFunc
	mu                sync.RWMutex
	loadingTemplates  map[string]bool // Guard against concurrent loading of the same template
}

// NewEnvironment creates a new Jinja2 environment
func NewEnvironment() *Environment {
	env := &Environment{
		loader:              nil,
		autoescape:          AutoescapeDefault,
		cacheSize:           400,
		trimBlocks:          false,
		lstripBlocks:        false,
		keepTrailingNewline: false,
		extensions:          []string{},
		policies:            make(map[string]interface{}),
		sandboxed:           false,
		secureDefaults:      true,
		securityPolicy:      DefaultSecurityPolicy(),
		securityManager:     GetGlobalSecurityManager(),
		filters:             make(map[string]FilterFunc),
		tests:               make(map[string]TestFunc),
		globals:             make(map[string]GlobalFunc),
		undefinedFactory:    func(name string) undefinedType { return DebugUndefined{name: name} },
		compiledTemplates:   make(map[string]*Template),
		cache:               NewTemplateCache(0, 400), // No TTL by default
		macroRegistry:       NewMacroRegistry(),
		loadingTemplates:    make(map[string]bool),
		newlineSequence:     "\n",
	}

	// Populate policy defaults to match Jinja2 behaviour
	env.policies["urlize.rel"] = "noopener"
	env.policies["urlize.target"] = nil
	env.policies["urlize.extra_schemes"] = nil

	// Register built-in filters
	env.registerBuiltinFilters()

	// Register built-in tests
	env.registerBuiltinTests()

	// Register built-in globals
	env.registerBuiltinGlobals()

	return env
}

// SetSecurityPolicy sets the security policy for the environment
func (env *Environment) SetSecurityPolicy(policy *SecurityPolicy) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.securityPolicy = policy
}

// GetSecurityPolicy returns the current security policy
func (env *Environment) GetSecurityPolicy() *SecurityPolicy {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.securityPolicy
}

// SetSecurityManager sets the security manager for the environment
func (env *Environment) SetSecurityManager(manager *SecurityManager) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.securityManager = manager
}

// GetSecurityManager returns the current security manager
func (env *Environment) GetSecurityManager() *SecurityManager {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.securityManager
}

// SetSandboxed enables or disables sandboxed mode
func (env *Environment) SetSandboxed(sandboxed bool) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.sandboxed = sandboxed
}

// IsSandboxed returns whether the environment is sandboxed
func (env *Environment) IsSandboxed() bool {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.sandboxed
}

// ExecuteTemplate executes a template with security controls
func (env *Environment) ExecuteTemplate(template *Template, vars map[string]interface{}, writer io.Writer) error {
	if env.sandboxed {
		sandbox := &SandboxEnvironment{
			Environment:     env,
			securityManager: env.securityManager,
			policyName:      "default",
		}

		if env.securityPolicy != nil {
			// Register the environment's policy if it has a custom name
			if env.securityPolicy.Name != "default" && env.securityPolicy.Name != "development" && env.securityPolicy.Name != "restricted" {
				env.securityManager.AddPolicy(env.securityPolicy.Name, env.securityPolicy)
				sandbox.policyName = env.securityPolicy.Name
			}
		}

		return sandbox.ExecuteTemplate(template, vars, writer)
	}

	// Create security context for monitoring
	secCtx, err := env.securityManager.CreateSecurityContext("default", template.name)
	if err != nil {
		return fmt.Errorf("failed to create security context: %w", err)
	}
	defer env.securityManager.CleanupSecurityContext(fmt.Sprintf("%s_%d", template.name, time.Now().UnixNano()))

	// Create context
	ctx := NewContextWithEnvironment(env, vars)
	if writer != nil {
		ctx.writer = writer
	}

	// Log execution start
	GetGlobalAuditManager().LogExecutionStart(template.name, "", "", secCtx.GetPolicy().Name, vars)

	// Execute with timing
	start := time.Now()
	err = template.ExecuteWithContext(ctx)
	duration := time.Since(start)

	// Log execution end
	if err != nil {
		GetGlobalAuditManager().LogExecutionEnd(template.name, "", "", duration, false, err.Error())
	} else {
		GetGlobalAuditManager().LogExecutionEnd(template.name, "", "", duration, true, "")
	}

	return err
}

// ExecuteToString executes a template and returns the result as a string
func (env *Environment) ExecuteToString(template *Template, vars map[string]interface{}) (string, error) {
	var buf strings.Builder
	err := env.ExecuteTemplate(template, vars, &buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SetLoader sets the template loader
func (env *Environment) SetLoader(loader Loader) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.loader = loader
}

// SetAutoescape sets the autoescape mode
func (env *Environment) SetAutoescape(value interface{}) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.autoescape = value
}

// SetTrimBlocks sets whether to trim the first newline after a block tag
func (env *Environment) SetTrimBlocks(trim bool) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.trimBlocks = trim
}

// SetLstripBlocks sets whether to strip whitespace before blocks
func (env *Environment) SetLstripBlocks(strip bool) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.lstripBlocks = strip
}

// SetKeepTrailingNewline sets whether to preserve trailing newlines
func (env *Environment) SetKeepTrailingNewline(keep bool) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.keepTrailingNewline = keep
}

// ShouldKeepTrailingNewline returns whether trailing newlines should be preserved.
func (env *Environment) ShouldKeepTrailingNewline() bool {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.keepTrailingNewline
}

// SetNewlineSequence configures the sequence used when generating newlines in filters
func (env *Environment) SetNewlineSequence(seq string) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.newlineSequence = seq
}

// NewlineSequence returns the configured newline sequence, defaulting to \n when unset
func (env *Environment) NewlineSequence() string {
	env.mu.RLock()
	defer env.mu.RUnlock()
	if env.newlineSequence == "" {
		return "\n"
	}
	return env.newlineSequence
}

// SetFinalize registers a finalize function executed on values before rendering
func (env *Environment) SetFinalize(f FinalizeFunc) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.finalize = f
}

// SetUndefinedFactory configures how undefined values are created
func (env *Environment) SetUndefinedFactory(factory UndefinedFactory) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.undefinedFactory = factory
}

// AddFilter adds a custom filter
func (env *Environment) AddFilter(name string, filter FilterFunc) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.filters[name] = filter
}

// AddTest adds a custom test
func (env *Environment) AddTest(name string, test TestFunc) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.tests[name] = test
}

// AddGlobal adds a global variable or function
func (env *Environment) AddGlobal(name string, value interface{}) {
	env.mu.Lock()
	defer env.mu.Unlock()

	switch fn := value.(type) {
	case GlobalFunc:
		env.globals[name] = fn
	case func(*Context, ...interface{}) (interface{}, error):
		env.globals[name] = GlobalFunc(fn)
	case func(*Context, ...interface{}) interface{}:
		env.globals[name] = func(ctx *Context, args ...interface{}) (interface{}, error) {
			return fn(ctx, args...), nil
		}
	case func(...interface{}) (interface{}, error):
		env.globals[name] = func(ctx *Context, args ...interface{}) (interface{}, error) {
			return fn(args...)
		}
	case func(...interface{}) interface{}:
		env.globals[name] = func(ctx *Context, args ...interface{}) (interface{}, error) {
			return fn(args...), nil
		}
	default:
		env.globals[name] = func(ctx *Context, args ...interface{}) (interface{}, error) {
			return value, nil
		}
	}
}

// GetFilter returns a filter function by name
func (env *Environment) GetFilter(name string) (FilterFunc, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	filter, ok := env.filters[name]
	return filter, ok
}

// GetTest returns a test function by name
func (env *Environment) GetTest(name string) (TestFunc, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	test, ok := env.tests[name]
	return test, ok
}

// GetGlobal returns a global function by name
func (env *Environment) GetGlobal(name string) (GlobalFunc, bool) {
	env.mu.RLock()
	defer env.mu.RUnlock()

	global, ok := env.globals[name]
	return global, ok
}

// shouldAutoescape determines if autoescaping should be enabled for a template
func (env *Environment) shouldAutoescape(templateName string) bool {
	env.mu.RLock()
	defer env.mu.RUnlock()

	switch v := env.autoescape.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(v) {
		case "default":
			return hasHTMLLikeExtension(templateName, []string{".html", ".htm", ".xml"})
		case "true", "on", "yes":
			return true
		case "false", "off", "no":
			return false
		default:
			return hasHTMLLikeExtension(templateName, []string{v})
		}
	case []string:
		return hasHTMLLikeExtension(templateName, v)
	case func(string) bool:
		return v(templateName)
	default:
		return hasHTMLLikeExtension(templateName, []string{".html", ".htm", ".xml"})
	}
}

func (env *Environment) applyFinalize(value interface{}) (interface{}, error) {
	env.mu.RLock()
	f := env.finalize
	env.mu.RUnlock()

	if f == nil {
		return value, nil
	}

	return f(value)
}

func (env *Environment) newUndefined(name string) undefinedType {
	env.mu.RLock()
	factory := env.undefinedFactory
	env.mu.RUnlock()
	if factory == nil {
		return DebugUndefined{name: name}
	}
	return factory(name)
}

func hasHTMLLikeExtension(templateName string, exts []string) bool {
	lowerName := strings.ToLower(templateName)
	for _, ext := range exts {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if strings.HasSuffix(lowerName, strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

// escape escapes HTML content
func (env *Environment) escape(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return template.HTMLEscapeString(v)
	case fmt.Stringer:
		return template.HTMLEscapeString(v.String())
	default:
		return template.HTMLEscapeString(fmt.Sprintf("%v", v))
	}
}

// resolveValue resolves a value using reflection
func (env *Environment) resolveValue(value interface{}, attr string) (interface{}, error) {
	if value == nil {
		return nil, NewUndefinedError(attr, nodes.Position{}, nil)
	}

	// Handle LoopContext specially
	if loopCtx, ok := value.(*LoopContext); ok {
		switch attr {
		case "index":
			return loopCtx.Index, nil
		case "index0":
			return loopCtx.Index0, nil
		case "revindex":
			return loopCtx.Revindex, nil
		case "revindex0":
			return loopCtx.Revindex0, nil
		case "first":
			return loopCtx.First, nil
		case "last":
			return loopCtx.Last, nil
		case "length":
			return loopCtx.Length, nil
		case "previtem":
			return loopCtx.Previtem, nil
		case "nextitem":
			return loopCtx.Nextitem, nil
		case "depth":
			return loopCtx.Depth, nil
		case "depth0":
			return loopCtx.Depth0, nil
		case "changed":
			return loopCtx.Changed, nil
		case "cycle":
			return loopCtx.Cycle, nil
		}
	}

	if ns, ok := value.(*Namespace); ok {
		if v, exists := ns.Get(attr); exists {
			return v, nil
		}
	}

	// Handle string methods
	if str, ok := value.(string); ok {
		switch attr {
		case "upper":
			return func() string { return strings.ToUpper(str) }, nil
		case "lower":
			return func() string { return strings.ToLower(str) }, nil
		case "title":
			return func() string { return strings.Title(str) }, nil
		case "capitalize":
			return func() string {
				if len(str) == 0 {
					return str
				}
				return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
			}, nil
		}
	}

	val := reflect.ValueOf(value)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, NewUndefinedError(attr, nodes.Position{}, nil)
		}
		capitalizedAttr := strings.Title(attr)
		if method := val.MethodByName(capitalizedAttr); method.IsValid() {
			return method.Interface(), nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		if val.Type().ConvertibleTo(reflect.TypeOf(map[string]interface{}{})) {
			mapVal := val.Convert(reflect.TypeOf(map[string]interface{}{}))
			if mapVal.IsValid() {
				if result := mapVal.MapIndex(reflect.ValueOf(attr)); result.IsValid() {
					return result.Interface(), nil
				}
			}
		} else {
			// Try to get the value directly from the map
			keyVal := reflect.ValueOf(attr)
			if result := val.MapIndex(keyVal); result.IsValid() {
				return result.Interface(), nil
			}
		}
	case reflect.Struct:
		// Try exported fields first (capitalized)
		capitalizedAttr := strings.Title(attr)
		field := val.FieldByName(capitalizedAttr)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}

		// Try exact field name
		field = val.FieldByName(attr)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}

		// Try methods
		method := val.MethodByName(capitalizedAttr)
		if method.IsValid() {
			return method.Interface(), nil
		}

		method = val.MethodByName(attr)
		if method.IsValid() {
			return method.Interface(), nil
		}
	case reflect.Slice, reflect.Array:
		// Try to convert to []interface{} first
		if val.Type().ConvertibleTo(reflect.TypeOf([]interface{}{})) {
			sliceVal := val.Convert(reflect.TypeOf([]interface{}{}))
			return sliceVal.Interface(), nil
		}
	case reflect.Interface:
		return env.resolveValue(val.Interface(), attr)
	}

	return nil, NewUndefinedError(attr, nodes.Position{}, nil)
}

// resolveIndex resolves a value by index
func (env *Environment) resolveIndex(value interface{}, index interface{}) (interface{}, error) {
	if value == nil {
		return nil, NewUndefinedError(fmt.Sprintf("%v", index), nodes.Position{}, nil)
	}

	val := reflect.ValueOf(value)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, NewUndefinedError(fmt.Sprintf("%v", index), nodes.Position{}, nil)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		keyVal := reflect.ValueOf(index)
		// Try to convert the key to the appropriate type
		if !keyVal.Type().ConvertibleTo(val.Type().Key()) {
			return nil, NewError(ErrorTypeTemplate,
				fmt.Sprintf("invalid map key type: %T", index),
				nodes.Position{}, nil)
		}
		convertedKey := keyVal.Convert(val.Type().Key())
		if result := val.MapIndex(convertedKey); result.IsValid() {
			return result.Interface(), nil
		}
	case reflect.Slice, reflect.Array:
		// Convert index to int
		var idx int
		switch i := index.(type) {
		case int:
			idx = i
		case int64:
			idx = int(i)
		case float64:
			idx = int(i)
		default:
			return nil, NewError(ErrorTypeTemplate,
				fmt.Sprintf("invalid index type: %T", index),
				nodes.Position{}, nil)
		}

		// Handle negative indices
		if idx < 0 {
			idx = val.Len() + idx
		}

		if idx < 0 || idx >= val.Len() {
			return nil, NewError(ErrorTypeRange,
				fmt.Sprintf("index %d out of range", idx),
				nodes.Position{}, nil)
		}

		return val.Index(idx).Interface(), nil
	case reflect.String:
		// Handle string indexing to return characters
		str := val.String()
		var idx int
		switch i := index.(type) {
		case int:
			idx = i
		case int64:
			idx = int(i)
		case float64:
			idx = int(i)
		default:
			return nil, NewError(ErrorTypeTemplate,
				fmt.Sprintf("invalid index type: %T", index),
				nodes.Position{}, nil)
		}

		// Handle negative indices
		if idx < 0 {
			idx = len(str) + idx
		}

		if idx < 0 || idx >= len(str) {
			return nil, NewError(ErrorTypeRange,
				fmt.Sprintf("index %d out of range", idx),
				nodes.Position{}, nil)
		}

		// Return character as string, not rune/byte
		return string(str[idx]), nil
	case reflect.Interface:
		return env.resolveIndex(val.Interface(), index)
	}

	return nil, NewError(ErrorTypeTemplate,
		fmt.Sprintf("cannot index %T", value),
		nodes.Position{}, nil)
}

// NewTemplate creates a new template from the given template string
func (env *Environment) NewTemplate(templateString string) (*Template, error) {
	return env.NewTemplateWithName(templateString, "template")
}

// NewTemplateWithName creates a new template with the given name
func (env *Environment) NewTemplateWithName(templateString, name string) (*Template, error) {
	if name == "" {
		name = "template"
	}

	return env.parseTemplateFromString(templateString, name)
}

// NewTemplateFromAST creates a template from an existing AST
func (env *Environment) NewTemplateFromAST(ast *nodes.Template, name string) (*Template, error) {
	if ast == nil {
		return nil, NewError(ErrorTypeTemplate, "AST cannot be nil", nodes.Position{}, nil)
	}

	template := &Template{
		name:        name,
		environment: env,
		ast:         ast,
		autoescape:  env.shouldAutoescape(name),
		blocks:      make(map[string]*nodes.Block),
		macros:      make(map[string]*nodes.Macro),
		imports:     make(map[string]*Template),
	}

	// Set the macro registry reference
	template.macroRegistry = env.macroRegistry

	// Pre-process the template to collect blocks and macros
	if err := template.preprocess(); err != nil {
		return nil, fmt.Errorf("failed to preprocess template: %w", err)
	}

	// Set up inheritance if this template extends another
	if err := template.setupInheritance(); err != nil {
		return nil, fmt.Errorf("failed to setup inheritance: %w", err)
	}

	return template, nil
}

// LoadTemplate loads and parses a template by name
func (env *Environment) LoadTemplate(name string) (*Template, error) {
	// Check cache first
	if tmpl, ok := env.cache.Get(name); ok {
		return tmpl, nil
	}

	env.mu.Lock()
	// Check if this template is currently being loaded (circular dependency detection)
	if env.loadingTemplates[name] {
		env.mu.Unlock()
		return nil, NewError(ErrorTypeTemplate, fmt.Sprintf("circular template inheritance detected: %s", name), nodes.Position{}, nil)
	}

	// Mark this template as being loaded
	env.loadingTemplates[name] = true
	env.mu.Unlock()

	// Ensure we clean up the loading flag even if there's an error
	defer func() {
		env.mu.Lock()
		delete(env.loadingTemplates, name)
		env.mu.Unlock()
	}()

	// Load from loader
	if env.loader == nil {
		return nil, NewError(ErrorTypeTemplate, "no loader configured", nodes.Position{}, nil)
	}

	source, err := env.loader.Load(name)
	if err != nil {
		return nil, WrapError(err, nodes.Position{}, nil)
	}

	// Parse template
	return env.parseTemplateFromString(source, name)
}

// ClearCache clears the template cache
func (env *Environment) ClearCache() {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.compiledTemplates = make(map[string]*Template)
	env.cache.Clear()
}

// CacheSize returns the current cache size
func (env *Environment) CacheSize() int {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.cache.Size()
}

// GetMacroRegistry returns the macro registry
func (env *Environment) GetMacroRegistry() *MacroRegistry {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.macroRegistry
}

// SetMacroRegistry sets the macro registry
func (env *Environment) SetMacroRegistry(registry *MacroRegistry) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.macroRegistry = registry
}

// AddGlobalMacro adds a global macro to the environment
func (env *Environment) AddGlobalMacro(name string, macro *Macro) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.macroRegistry.RegisterGlobal(name, macro)
}

// GetGlobalMacro gets a global macro by name
func (env *Environment) GetGlobalMacro(name string) (*Macro, error) {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.macroRegistry.FindMacro(nil, name)
}

// ClearMacroRegistry clears the macro registry
func (env *Environment) ClearMacroRegistry() {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.macroRegistry.Clear()
}

// GetMacroStats returns macro registry statistics
func (env *Environment) GetMacroStats() map[string]int {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.macroRegistry.Stats()
}

// parseTemplateFromString parses a template from a string
func (env *Environment) parseTemplateFromString(source, name string) (*Template, error) {
	// Create parser environment using the environment configuration
	parserEnv := &parser.Environment{
		TrimBlocks:          env.trimBlocks,
		LstripBlocks:        env.lstripBlocks,
		KeepTrailingNewline: env.keepTrailingNewline,
	}

	// Parse the template
	ast, err := parser.ParseTemplateWithEnv(parserEnv, source, name, name)
	if err != nil {
		return nil, WrapError(err, nodes.Position{}, nil)
	}

	// Collect parent blocks during inheritance processing
	parentBlocks := make(map[string]*nodes.Block)

	// Process inheritance
	processedAST, err := env.processInheritanceWithContext(ast, name, make(map[string]bool), parentBlocks)
	if err != nil {
		return nil, err
	}

	// Create template
	tmpl, err := env.NewTemplateFromAST(processedAST, name)
	if err != nil {
		return nil, err
	}

	// If this template has inheritance context, update it with the parent blocks
	if tmpl.inheritanceCtx != nil {
		for blockName, parentBlock := range parentBlocks {
			tmpl.inheritanceCtx.SetParentBlock(blockName, parentBlock)
		}
	}

	// Cache the template
	dependencies := make(map[string]time.Time)
	dependencies[name] = time.Now() // In a real implementation, get actual file mod time
	env.cache.Set(name, tmpl, dependencies)

	return tmpl, nil
}

// processInheritance resolves template inheritance chains
func (env *Environment) processInheritance(ast *nodes.Template, name string, visited map[string]bool) (*nodes.Template, error) {
	return env.processInheritanceWithContext(ast, name, visited, nil)
}

// processInheritanceWithContext resolves template inheritance chains with context for parent blocks
func (env *Environment) processInheritanceWithContext(ast *nodes.Template, name string, visited map[string]bool, parentBlocks map[string]*nodes.Block) (*nodes.Template, error) {
	// Check for circular dependencies
	if visited[name] {
		return nil, NewError(ErrorTypeTemplate, fmt.Sprintf("circular template inheritance detected: %s", name), nodes.Position{}, nil)
	}
	visited[name] = true

	// Find the extends statement
	var extendsNode *nodes.Extends
	var childBlocks []*nodes.Block
	var nonExtendsNodes []nodes.Node

	for _, node := range ast.Body {
		if ext, ok := node.(*nodes.Extends); ok {
			if extendsNode != nil {
				return nil, NewError(ErrorTypeTemplate, "multiple extends statements not allowed", node.GetPosition(), node)
			}
			extendsNode = ext
		} else if block, ok := node.(*nodes.Block); ok {
			childBlocks = append(childBlocks, block)
		} else {
			// Only keep non-whitespace template data and assignment nodes
			// Skip pure whitespace TemplateData nodes when template extends another
			if templateData, ok := node.(*nodes.TemplateData); ok {
				// Skip if it's only whitespace
				if strings.TrimSpace(templateData.Data) == "" {
					continue
				}
			}
			nonExtendsNodes = append(nonExtendsNodes, node)
		}
	}

	// If no extends, return the original AST and collect parent blocks
	if extendsNode == nil {
		// This is the top-level parent template, collect its blocks
		if parentBlocks != nil {
			for _, block := range childBlocks {
				parentBlocks[block.Name] = block
			}
		}
		return ast, nil
	}

	// Evaluate the parent template name
	parentNameValue := env.evaluateExpression(extendsNode.Template)
	if err, ok := parentNameValue.(error); ok {
		return nil, err
	}

	parentName, ok := parentNameValue.(string)
	if !ok {
		return nil, NewError(ErrorTypeTemplate, "extends template name must be a string", extendsNode.GetPosition(), extendsNode)
	}

	// Check for circular dependencies BEFORE loading the parent template
	if visited[parentName] {
		return nil, NewError(ErrorTypeTemplate, fmt.Sprintf("circular template inheritance detected: %s", parentName), nodes.Position{}, nil)
	}

	// Load the parent template
	parent, err := env.LoadTemplate(parentName)
	if err != nil {
		return nil, err
	}

	// Process parent inheritance recursively
	parentAST, err := env.processInheritanceWithContext(parent.AST(), parentName, visited, parentBlocks)
	if err != nil {
		return nil, err
	}

	// Apply child blocks to parent
	resultAST, err := env.applyBlocksToParent(parentAST, childBlocks)
	if err != nil {
		return nil, err
	}

	// Note: In Jinja2, when a template extends another, only blocks are inherited.
	// Content outside blocks (except variable assignments) is discarded.
	// For now, we skip nonExtendsNodes entirely.

	return resultAST, nil
}

// applyBlocksToParent applies child template blocks to parent template
func (env *Environment) applyBlocksToParent(parentAST *nodes.Template, childBlocks []*nodes.Block) (*nodes.Template, error) {
	// Convert child blocks slice to map
	childBlockMap := make(map[string]*nodes.Block)
	for _, block := range childBlocks {
		trimBlockEdges(block)
		childBlockMap[block.Name] = block
	}

	// Create a new template body
	newBody := make([]nodes.Node, len(parentAST.Body))
	for i, node := range parentAST.Body {
		newBody[i] = replaceBlocksInNode(node, childBlockMap)
	}

	// Create new template
	result := &nodes.Template{
		Body: newBody,
	}
	result.SetPosition(parentAST.GetPosition())

	return result, nil
}

// trimBlockEdges removes leading and trailing whitespace-only output nodes from a block.
func trimBlockEdges(block *nodes.Block) {
	if block == nil {
		return
	}

	trimBlockLeadingWhitespace(block)
	trimBlockTrailingWhitespace(block)
}

func trimBlockLeadingWhitespace(block *nodes.Block) {
	for len(block.Body) > 0 {
		node := block.Body[0]
		switch n := node.(type) {
		case *nodes.Output:
			if trimOutputLeadingWhitespace(n) {
				if len(n.Nodes) == 0 {
					block.Body = block.Body[1:]
					continue
				}
				continue
			}
			return
		case *nodes.TemplateData:
			if strings.TrimSpace(n.Data) == "" {
				block.Body = block.Body[1:]
				continue
			}
			return
		default:
			return
		}
	}
}

func trimBlockTrailingWhitespace(block *nodes.Block) {
	for len(block.Body) > 0 {
		idx := len(block.Body) - 1
		node := block.Body[idx]
		switch n := node.(type) {
		case *nodes.Output:
			if trimOutputTrailingWhitespace(n) {
				if len(n.Nodes) == 0 {
					block.Body = block.Body[:idx]
					continue
				}
				continue
			}
			return
		case *nodes.TemplateData:
			if strings.TrimSpace(n.Data) == "" {
				block.Body = block.Body[:idx]
				continue
			}
			return
		default:
			return
		}
	}
}

// trimOutputLeadingWhitespace removes whitespace-only template data nodes from
// the beginning of an output node.
func trimOutputLeadingWhitespace(output *nodes.Output) bool {
	changed := false
	for len(output.Nodes) > 0 {
		expr := output.Nodes[0]
		td, ok := expr.(*nodes.TemplateData)
		if !ok {
			break
		}
		if strings.TrimSpace(td.Data) == "" {
			output.Nodes = output.Nodes[1:]
			changed = true
			continue
		}
		trimmed := strings.TrimLeft(td.Data, " \t\r\n")
		if trimmed != td.Data {
			if trimmed == "" {
				output.Nodes = output.Nodes[1:]
			} else {
				td.Data = trimmed
			}
			changed = true
			continue
		}
		break
	}
	return changed
}

// trimOutputTrailingWhitespace removes whitespace-only template data nodes from
// the end of an output node.
func trimOutputTrailingWhitespace(output *nodes.Output) bool {
	changed := false
	for len(output.Nodes) > 0 {
		idx := len(output.Nodes) - 1
		expr := output.Nodes[idx]
		td, ok := expr.(*nodes.TemplateData)
		if !ok {
			break
		}
		if strings.TrimSpace(td.Data) == "" {
			output.Nodes = output.Nodes[:idx]
			changed = true
			continue
		}
		trimmed := strings.TrimRight(td.Data, " \t\r\n")
		if trimmed != td.Data {
			if trimmed == "" {
				output.Nodes = output.Nodes[:idx]
			} else {
				td.Data = trimmed
			}
			changed = true
			continue
		}
		break
	}
	return changed
}

// replaceBlocksInNode walks the node tree and replaces block nodes with overridden
// child blocks when available.
func replaceBlocksInNode(node nodes.Node, childBlockMap map[string]*nodes.Block) nodes.Node {
	switch n := node.(type) {
	case *nodes.Block:
		if child, ok := childBlockMap[n.Name]; ok {
			return child
		}
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.Template:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.For:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		n.Else = replaceBlocksInBody(n.Else, childBlockMap)
		return n
	case *nodes.If:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		for _, elif := range n.Elif {
			elif.Body = replaceBlocksInBody(elif.Body, childBlockMap)
		}
		n.Else = replaceBlocksInBody(n.Else, childBlockMap)
		return n
	case *nodes.Macro:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.CallBlock:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.FilterBlock:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.AssignBlock:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.Scope:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.With:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.ScopedEvalContextModifier:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	case *nodes.Trans:
		n.Body = replaceBlocksInBody(n.Body, childBlockMap)
		return n
	default:
		return node
	}
}

func replaceBlocksInBody(body []nodes.Node, childBlockMap map[string]*nodes.Block) []nodes.Node {
	for i, child := range body {
		body[i] = replaceBlocksInNode(child, childBlockMap)
	}
	return body
}

// evaluateExpression evaluates a simple expression (used for template names)
func (env *Environment) evaluateExpression(expr nodes.Expr) interface{} {
	// This is a simplified expression evaluator for template names
	// In a full implementation, this would use the proper evaluator
	switch e := expr.(type) {
	case *nodes.Const:
		return e.Value
	case *nodes.Name:
		if e.Name == "true" {
			return true
		} else if e.Name == "false" {
			return false
		} else if e.Name == "none" {
			return nil
		}
		return NewError(ErrorTypeTemplate, fmt.Sprintf("undefined variable: %s", e.Name), e.GetPosition(), e)
	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("complex expressions not supported in extends: %T", expr), e.GetPosition(), expr)
	}
}

// ParseFile parses a template file by name
func (env *Environment) ParseFile(name string) (*Template, error) {
	return env.LoadTemplate(name)
}

// SetCacheTTL sets the cache time-to-live
func (env *Environment) SetCacheTTL(ttl time.Duration) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.cache.ttl = ttl
}

// NewTemplateFromSource creates a new template from source without using the loader
func (env *Environment) NewTemplateFromSource(source, name string) (*Template, error) {
	return env.parseTemplateFromString(source, name)
}

// registerBuiltinGlobals registers built-in global functions
func (env *Environment) registerBuiltinGlobals() {
	// Add built-in global functions
	env.AddGlobal("range", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.rangeFunc(args...)
	}))
	env.AddGlobal("lipsum", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.lipsumFunc(args...)
	}))
	env.AddGlobal("dict", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.dictFunc(args...)
	}))
	env.AddGlobal("cycler", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.cyclerFunc(args...)
	}))
	env.AddGlobal("joiner", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.joinerFunc(args...)
	}))
	env.AddGlobal("namespace", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.namespaceFunc(args...)
	}))
	env.AddGlobal("class", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.classFunc(args...)
	}))
	env.AddGlobal("_", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.gettextFunc(args...)
	}))
	env.AddGlobal("gettext", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.gettextFunc(args...)
	}))
	env.AddGlobal("ngettext", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.ngettextFunc(args...)
	}))
	env.AddGlobal("debug", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.debugFunc(args...)
	}))
	env.AddGlobal("self", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.selfFunc(args...)
	}))
	env.AddGlobal("environment", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.environmentFunc(args...)
	}))
	env.AddGlobal("url_for", GlobalFunc(func(ctx *Context, args ...interface{}) (interface{}, error) {
		return ctx.urlForFunc(args...)
	}))
}

// SetURLFor sets the callback used by the `url_for` global.
func (env *Environment) SetURLFor(fn GlobalFunc) {
	env.mu.Lock()
	defer env.mu.Unlock()
	env.urlFor = fn
}
