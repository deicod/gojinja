package runtime

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"

	"github.com/deicod/gojinja/nodes"
)

// LoopContext represents the context of a for loop
type LoopContext struct {
	Index     int         `json:"index"`
	Index0    int         `json:"index0"`
	Revindex  int         `json:"revindex"`
	Revindex0 int         `json:"revindex0"`
	First     bool        `json:"first"`
	Last      bool        `json:"last"`
	Length    int         `json:"length"`
	Previtem  interface{} `json:"previtem,omitempty"`
	Nextitem  interface{} `json:"nextitem,omitempty"`
	Depth     int         `json:"depth"`
	Depth0    int         `json:"depth0"`
	Changed   bool        `json:"changed"`
	Cycle     interface{} `json:"cycle,omitempty"`
}

// Scope represents a variable scope
type Scope struct {
	parent    *Scope
	vars      map[string]interface{}
	exports   map[string]interface{}
	overrides map[string]interface{}
}

// NewScope creates a new scope
func NewScope() *Scope {
	return &Scope{
		vars:      make(map[string]interface{}),
		exports:   make(map[string]interface{}),
		overrides: make(map[string]interface{}),
	}
}

// NewChildScope creates a child scope
func (s *Scope) NewChildScope() *Scope {
	child := NewScope()
	child.parent = s
	return child
}

// Set sets a variable in the current scope
func (s *Scope) Set(name string, value interface{}) {
	s.vars[name] = value
}

// Get gets a variable, searching parent scopes if not found
func (s *Scope) Get(name string) (interface{}, bool) {
	// Check current scope first
	if value, ok := s.vars[name]; ok {
		return value, true
	}

	// Check exports
	if value, ok := s.exports[name]; ok {
		return value, true
	}

	// Check overrides
	if value, ok := s.overrides[name]; ok {
		return value, true
	}

	// Check parent scope
	if s.parent != nil {
		return s.parent.Get(name)
	}

	return nil, false
}

// SetExport sets an exported variable
func (s *Scope) SetExport(name string, value interface{}) {
	s.exports[name] = value
}

// SetOverride sets an override variable
func (s *Scope) SetOverride(name string, value interface{}) {
	s.overrides[name] = value
}

// Delete deletes a variable from the current scope
func (s *Scope) Delete(name string) {
	delete(s.vars, name)
	delete(s.exports, name)
	delete(s.overrides, name)
}

// Has checks if a variable exists in any scope
func (s *Scope) Has(name string) bool {
	_, ok := s.Get(name)
	return ok
}

// Keys returns all variable names in the current scope and parents
func (s *Scope) Keys() []string {
	keys := make(map[string]bool)

	// Collect from current scope
	for k := range s.vars {
		keys[k] = true
	}
	for k := range s.exports {
		keys[k] = true
	}
	for k := range s.overrides {
		keys[k] = true
	}

	// Collect from parent scopes
	if s.parent != nil {
		parentKeys := s.parent.Keys()
		for _, k := range parentKeys {
			keys[k] = true
		}
	}

	result := make([]string, 0, len(keys))
	for k := range keys {
		result = append(result, k)
	}
	return result
}

// Local returns a copy of variables in the current scope only
func (s *Scope) Local() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range s.vars {
		result[k] = v
	}
	return result
}

// All returns all variables from all scopes
func (s *Scope) All() map[string]interface{} {
	result := make(map[string]interface{})

	if s.parent != nil {
		parentVars := s.parent.All()
		for k, v := range parentVars {
			result[k] = v
		}
	}

	for k, v := range s.vars {
		result[k] = v
	}
	for k, v := range s.exports {
		result[k] = v
	}
	for k, v := range s.overrides {
		result[k] = v
	}

	return result
}

// Context represents the rendering context
type Context struct {
	environment *Environment
	scope       *Scope
	autoescape  bool
	writer      io.Writer

	// Loop handling
	loopStack   []*LoopContext
	currentLoop *LoopContext

	// Template inheritance
	blocks  map[string]*nodes.Block
	parent  *Template
	current *Template

	// Macro handling
	macroStack  []*Macro
	callerStack []*MacroCaller

	// Error handling
	errors []error

	// Concurrency safety
	mu sync.RWMutex
}

// NewContext creates a new context with the given variables
func NewContext(vars map[string]interface{}) *Context {
	ctx := &Context{
		scope:       NewScope(),
		loopStack:   make([]*LoopContext, 0),
		blocks:      make(map[string]*nodes.Block),
		macroStack:  make([]*Macro, 0),
		callerStack: make([]*MacroCaller, 0),
		errors:      make([]error, 0),
	}

	// Set initial variables
	if vars != nil {
		for k, v := range vars {
			ctx.scope.Set(k, v)
		}
	}

	return ctx
}

// NewContextWithEnvironment creates a new context with an environment
func NewContextWithEnvironment(env *Environment, vars map[string]interface{}) *Context {
	ctx := NewContext(vars)
	ctx.environment = env

	// Add global variables to the context
	if env != nil {
		ctx.addGlobals()
	}

	return ctx
}

// addGlobals adds global variables from the environment
func (ctx *Context) addGlobals() {
	if ctx.environment == nil {
		return
	}

	// Add standard globals - wrap methods to ensure correct type
	rangeWrapper := func(c *Context, args ...interface{}) (interface{}, error) {
		return c.rangeFunc(args...)
	}
	lipsumWrapper := func(c *Context, args ...interface{}) (interface{}, error) {
		return c.lipsumFunc(args...)
	}
	dictWrapper := func(c *Context, args ...interface{}) (interface{}, error) {
		return c.dictFunc(args...)
	}
	cyclerWrapper := func(c *Context, args ...interface{}) (interface{}, error) {
		return c.cyclerFunc(args...)
	}
	joinerWrapper := func(c *Context, args ...interface{}) (interface{}, error) {
		return c.joinerFunc(args...)
	}

	ctx.scope.Set("range", GlobalFunc(rangeWrapper))
	ctx.scope.Set("lipsum", GlobalFunc(lipsumWrapper))
	ctx.scope.Set("dict", GlobalFunc(dictWrapper))
	ctx.scope.Set("cycler", GlobalFunc(cyclerWrapper))
	ctx.scope.Set("joiner", GlobalFunc(joinerWrapper))

	// Add custom globals from environment
	for name, globalFunc := range ctx.environment.globals {
		ctx.scope.Set(name, globalFunc)
	}
}

// Set sets a variable in the current scope
func (ctx *Context) Set(name string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.scope.Set(name, value)
}

// Get gets a variable from the context
func (ctx *Context) Get(name string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.scope.Get(name)
}

// Delete deletes a variable from the current scope
func (ctx *Context) Delete(name string) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.scope.Delete(name)
}

// Has checks if a variable exists in the context
func (ctx *Context) Has(name string) bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.scope.Has(name)
}

// PushScope creates a new child scope
func (ctx *Context) PushScope() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.scope = ctx.scope.NewChildScope()
}

// PopScope returns to the parent scope
func (ctx *Context) PopScope() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.scope.parent != nil {
		ctx.scope = ctx.scope.parent
	}
}

// PushLoop pushes a new loop context
func (ctx *Context) PushLoop(length, depth int) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	loopCtx := &LoopContext{
		Length: length,
		Depth:  depth,
		Depth0: depth - 1,
	}

	ctx.loopStack = append(ctx.loopStack, loopCtx)
	ctx.currentLoop = loopCtx
}

// UpdateLoop updates the current loop context
func (ctx *Context) UpdateLoop(index int, currentItem, prevItem, nextItem interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if ctx.currentLoop != nil {
		ctx.currentLoop.Index = index + 1
		ctx.currentLoop.Index0 = index
		ctx.currentLoop.Revindex = ctx.currentLoop.Length - index
		ctx.currentLoop.Revindex0 = ctx.currentLoop.Length - index - 1
		ctx.currentLoop.First = index == 0
		ctx.currentLoop.Last = index == ctx.currentLoop.Length-1
		ctx.currentLoop.Previtem = prevItem
		ctx.currentLoop.Nextitem = nextItem
	}
}

// PopLoop pops the current loop context
func (ctx *Context) PopLoop() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	if len(ctx.loopStack) > 0 {
		ctx.loopStack = ctx.loopStack[:len(ctx.loopStack)-1]
		if len(ctx.loopStack) > 0 {
			ctx.currentLoop = ctx.loopStack[len(ctx.loopStack)-1]
		} else {
			ctx.currentLoop = nil
		}
	}
}

// CurrentLoop returns the current loop context
func (ctx *Context) CurrentLoop() *LoopContext {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.currentLoop
}

// SetAutoescape sets the autoescape mode
func (ctx *Context) SetAutoescape(autoescape bool) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.autoescape = autoescape
}

// ShouldAutoescape returns whether autoescaping is enabled
func (ctx *Context) ShouldAutoescape() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.autoescape
}

// Resolve resolves a variable name in the context
func (ctx *Context) Resolve(name string) (interface{}, error) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	// Check for special names
	if name == "loop" {
		if ctx.currentLoop == nil {
			return nil, NewUndefinedError("loop", nodes.Position{}, nil)
		}
		return ctx.currentLoop, nil
	}

	value, ok := ctx.scope.Get(name)
	if !ok {
		if ctx.environment != nil {
			return ctx.environment.newUndefined(name), nil
		}
		return DebugUndefined{name: name}, nil
	}

	return value, nil
}

// ResolveAttribute resolves an attribute access
func (ctx *Context) ResolveAttribute(obj interface{}, attr string) (interface{}, error) {
	if obj == nil {
		return nil, NewUndefinedError(attr, nodes.Position{}, nil)
	}

	if ctx.environment != nil {
		return ctx.environment.resolveValue(obj, attr)
	}

	// Fallback implementation using reflection
	return resolveAttributeFallback(obj, attr)
}

// ResolveIndex resolves an index access
func (ctx *Context) ResolveIndex(obj interface{}, index interface{}) (interface{}, error) {
	if obj == nil {
		return nil, NewUndefinedError(fmt.Sprintf("%v", index), nodes.Position{}, nil)
	}

	if ctx.environment != nil {
		return ctx.environment.resolveIndex(obj, index)
	}

	// Fallback implementation using reflection
	return resolveIndexFallback(obj, index)
}

// AddError adds an error to the context
func (ctx *Context) AddError(err error) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.errors = append(ctx.errors, err)
}

// GetErrors returns all errors that occurred during rendering
func (ctx *Context) GetErrors() []error {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	errors := make([]error, len(ctx.errors))
	copy(errors, ctx.errors)
	return errors
}

// HasErrors returns whether any errors occurred
func (ctx *Context) HasErrors() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return len(ctx.errors) > 0
}

// ClearErrors clears all errors
func (ctx *Context) ClearErrors() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.errors = make([]error, 0)
}

// PushMacro pushes a macro onto the macro stack
func (ctx *Context) PushMacro(macro *Macro) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.macroStack = append(ctx.macroStack, macro)
}

// PopMacro pops the current macro from the macro stack
func (ctx *Context) PopMacro() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if len(ctx.macroStack) > 0 {
		ctx.macroStack = ctx.macroStack[:len(ctx.macroStack)-1]
	}
}

// CurrentMacro returns the current macro being executed
func (ctx *Context) CurrentMacro() *Macro {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.macroStack) > 0 {
		return ctx.macroStack[len(ctx.macroStack)-1]
	}
	return nil
}

// InMacro checks if we're currently executing a macro
func (ctx *Context) InMacro() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return len(ctx.macroStack) > 0
}

// MacroDepth returns the current macro call depth
func (ctx *Context) MacroDepth() int {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return len(ctx.macroStack)
}

// PushCaller pushes a caller context onto the caller stack
func (ctx *Context) PushCaller(caller *MacroCaller) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	ctx.callerStack = append(ctx.callerStack, caller)
}

// PopCaller pops the current caller from the caller stack
func (ctx *Context) PopCaller() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if len(ctx.callerStack) > 0 {
		ctx.callerStack = ctx.callerStack[:len(ctx.callerStack)-1]
	}
}

// CurrentCaller returns the current caller context
func (ctx *Context) CurrentCaller() *MacroCaller {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if len(ctx.callerStack) > 0 {
		return ctx.callerStack[len(ctx.callerStack)-1]
	}
	return nil
}

// HasCaller checks if there's a caller context available
func (ctx *Context) HasCaller() bool {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return len(ctx.callerStack) > 0
}

// GetMacroVariable gets a variable from macro scope (checks caller first, then current scope)
func (ctx *Context) GetMacroVariable(name string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()

	// First check caller context variables
	if len(ctx.callerStack) > 0 {
		caller := ctx.callerStack[len(ctx.callerStack)-1]
		if value, ok := caller.Variables[name]; ok {
			return value, true
		}
	}

	// Then check current scope
	return ctx.scope.Get(name)
}

// SetMacroVariable sets a variable in macro scope
func (ctx *Context) SetMacroVariable(name string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// If we have a caller context, set it in the caller's variables
	if len(ctx.callerStack) > 0 {
		caller := ctx.callerStack[len(ctx.callerStack)-1]
		caller.Variables[name] = value
		return
	}

	// Otherwise set in current scope
	ctx.scope.Set(name, value)
}

// CreateMacroCaller creates a new macro caller context
func (ctx *Context) CreateMacroCaller(name string, args []interface{}, kwargs map[string]interface{}) *MacroCaller {
	return &MacroCaller{
		Name:      name,
		Args:      args,
		Kwargs:    kwargs,
		Variables: make(map[string]interface{}),
	}
}

// resolveAttributeFallback is a fallback for attribute resolution
func resolveAttributeFallback(obj interface{}, attr string) (interface{}, error) {
	val := reflect.ValueOf(obj)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, NewUndefinedError(attr, nodes.Position{}, nil)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		mapVal := val.Convert(reflect.TypeOf(map[string]interface{}{}))
		if mapVal.IsValid() {
			if result := mapVal.MapIndex(reflect.ValueOf(attr)); result.IsValid() {
				return result.Interface(), nil
			}
		}
	case reflect.Struct:
		field := val.FieldByName(attr)
		if field.IsValid() && field.CanInterface() {
			return field.Interface(), nil
		}

		// Try methods
		method := val.MethodByName(attr)
		if method.IsValid() && method.CanInterface() {
			return method.Interface(), nil
		}
	}

	return nil, NewUndefinedError(attr, nodes.Position{}, nil)
}

// resolveIndexFallback is a fallback for index resolution
func resolveIndexFallback(obj interface{}, index interface{}) (interface{}, error) {
	val := reflect.ValueOf(obj)

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
	}

	return nil, NewError(ErrorTypeTemplate,
		fmt.Sprintf("cannot index %T", obj),
		nodes.Position{}, nil)
}

// Global function implementations

func (ctx *Context) rangeFunc(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, NewError(ErrorTypeTemplate, "range() requires at least one argument", nodes.Position{}, nil)
	}

	var start, end, step int

	switch len(args) {
	case 1:
		// range(stop)
		stop, ok := toInt(args[0])
		if !ok {
			return nil, NewError(ErrorTypeTemplate, "range() argument must be an integer", nodes.Position{}, nil)
		}
		start, end, step = 0, stop, 1
	case 2:
		// range(start, stop)
		startVal, ok1 := toInt(args[0])
		endVal, ok2 := toInt(args[1])
		if !ok1 || !ok2 {
			return nil, NewError(ErrorTypeTemplate, "range() arguments must be integers", nodes.Position{}, nil)
		}
		start, end, step = startVal, endVal, 1
	case 3:
		// range(start, stop, step)
		startVal, ok1 := toInt(args[0])
		endVal, ok2 := toInt(args[1])
		stepVal, ok3 := toInt(args[2])
		if !ok1 || !ok2 || !ok3 {
			return nil, NewError(ErrorTypeTemplate, "range() arguments must be integers", nodes.Position{}, nil)
		}
		start, end, step = startVal, endVal, stepVal
	default:
		return nil, NewError(ErrorTypeTemplate, "range() accepts at most 3 arguments", nodes.Position{}, nil)
	}

	if step == 0 {
		return nil, NewError(ErrorTypeTemplate, "range() step argument must not be zero", nodes.Position{}, nil)
	}

	result := make([]interface{}, 0)
	for i := start; (step > 0 && i < end) || (step < 0 && i > end); i += step {
		result = append(result, i)
	}

	return result, nil
}

func (ctx *Context) lipsumFunc(args ...interface{}) (interface{}, error) {
	// Simple lorem ipsum implementation
	lorem := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	return lorem, nil
}

func (ctx *Context) dictFunc(args ...interface{}) (interface{}, error) {
	result := make(map[string]interface{})

	remaining := args
	if len(remaining) > 0 {
		if kwargMap, ok := toStringInterfaceMap(remaining[len(remaining)-1]); ok {
			for k, v := range kwargMap {
				result[k] = v
			}
			remaining = remaining[:len(remaining)-1]
		}
	}

	if len(remaining)%2 != 0 {
		return nil, NewError(ErrorTypeTemplate, "dict() requires name/value pairs", nodes.Position{}, nil)
	}

	for i := 0; i < len(remaining); i += 2 {
		key, ok := remaining[i].(string)
		if !ok {
			return nil, NewError(ErrorTypeTemplate, "dict() keys must be strings", nodes.Position{}, nil)
		}
		result[key] = remaining[i+1]
	}

	return result, nil
}

func (ctx *Context) cyclerFunc(args ...interface{}) (interface{}, error) {
	// Simple cycler implementation
	if len(args) == 0 {
		return nil, NewError(ErrorTypeTemplate, "cycler() requires at least one argument", nodes.Position{}, nil)
	}

	return &cycler{
		items: args,
		index: 0,
	}, nil
}

func (ctx *Context) joinerFunc(args ...interface{}) (interface{}, error) {
	sep := ""
	if len(args) > 0 {
		if s, ok := args[0].(string); ok {
			sep = s
		}
	}

	return &joiner{
		separator: sep,
		first:     true,
	}, nil
}

func (ctx *Context) namespaceFunc(args ...interface{}) (interface{}, error) {
	initial := make(map[string]interface{})
	remaining := args

	if len(remaining) > 0 {
		if kwargMap, ok := toStringInterfaceMap(remaining[len(remaining)-1]); ok {
			for k, v := range kwargMap {
				initial[k] = v
			}
			remaining = remaining[:len(remaining)-1]
		}
	}

	if len(remaining)%2 != 0 {
		return nil, NewError(ErrorTypeTemplate, "namespace() requires name/value pairs", nodes.Position{}, nil)
	}

	for i := 0; i < len(remaining); i += 2 {
		key, ok := remaining[i].(string)
		if !ok {
			return nil, NewError(ErrorTypeTemplate, "namespace() keys must be strings", nodes.Position{}, nil)
		}
		initial[key] = remaining[i+1]
	}

	return NewNamespace(initial), nil
}

func (ctx *Context) classFunc(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return nil, NewError(ErrorTypeTemplate, "class() requires a name", nodes.Position{}, nil)
	}

	name := toString(args[0])
	result := make(map[string]interface{})
	if name != "" {
		result["__name__"] = name
	}

	mergeAttributes := func(value interface{}) {
		if value == nil {
			return
		}
		if ns, ok := value.(*Namespace); ok {
			for k, v := range ns.Items() {
				result[k] = v
			}
			return
		}
		if mapping, ok := toStringInterfaceMap(value); ok {
			for k, v := range mapping {
				result[k] = v
			}
		}
	}

	remaining := args[1:]
	var kwargMap map[string]interface{}
	if len(remaining) > 0 {
		if kw, ok := toStringInterfaceMap(remaining[len(remaining)-1]); ok {
			kwargMap = kw
			remaining = remaining[:len(remaining)-1]
		}
	}

	for _, item := range remaining {
		mergeAttributes(item)
	}

	if kwargMap != nil {
		mergeAttributes(kwargMap)
	}

	return NewNamespace(result), nil
}

func (ctx *Context) debugFunc(args ...interface{}) (interface{}, error) {
	return fmt.Sprintf("%v", ctx.scope.All()), nil
}

func (ctx *Context) selfFunc(args ...interface{}) (interface{}, error) {
	return ctx.current, nil
}

func (ctx *Context) contextFunc(args ...interface{}) (interface{}, error) {
	return ctx.scope.All(), nil
}

func (ctx *Context) environmentFunc(args ...interface{}) (interface{}, error) {
	return ctx.environment, nil
}

func (ctx *Context) urlForFunc(args ...interface{}) (interface{}, error) {
	if ctx.environment != nil && ctx.environment.urlFor != nil {
		return ctx.environment.urlFor(ctx, args...)
	}
	return nil, NewError(ErrorTypeTemplate, "url_for is not configured for this environment", nodes.Position{}, nil)
}

func (ctx *Context) gettextFunc(args ...interface{}) (interface{}, error) {
	if len(args) == 0 {
		return "", nil
	}
	message := toString(args[0])
	if len(args) == 1 {
		return message, nil
	}

	replacements := args[1]
	if mapping, ok := toStringInterfaceMap(replacements); ok {
		return formatWithMap(message, mapping), nil
	}

	return fmt.Sprintf(message, args[1:]...), nil
}

func (ctx *Context) ngettextFunc(args ...interface{}) (interface{}, error) {
	if len(args) < 3 {
		return nil, NewError(ErrorTypeTemplate, "ngettext() requires singular, plural, count", nodes.Position{}, nil)
	}
	singular := toString(args[0])
	plural := toString(args[1])
	count, ok := toInt(args[2])
	if !ok {
		return nil, NewError(ErrorTypeTemplate, "ngettext() count must be numeric", nodes.Position{}, nil)
	}

	choice := plural
	if count == 1 {
		choice = singular
	}

	if len(args) > 3 {
		replacements := args[3]
		if mapping, ok := toStringInterfaceMap(replacements); ok {
			mapping["count"] = count
			return formatWithMap(choice, mapping), nil
		}
		vals := append([]interface{}{count}, args[4:]...)
		return fmt.Sprintf(choice, vals...), nil
	}

	return fmt.Sprintf(choice, count), nil
}

func (ctx *Context) pgettextFunc(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, NewError(ErrorTypeTemplate, "pgettext() requires context and message", nodes.Position{}, nil)
	}

	message := toString(args[1])
	if len(args) == 2 {
		return message, nil
	}

	replacements := args[2]
	if mapping, ok := toStringInterfaceMap(replacements); ok {
		return formatWithMap(message, mapping), nil
	}

	return fmt.Sprintf(message, args[2:]...), nil
}

func (ctx *Context) npgettextFunc(args ...interface{}) (interface{}, error) {
	if len(args) < 4 {
		return nil, NewError(ErrorTypeTemplate, "npgettext() requires context, singular, plural, count", nodes.Position{}, nil)
	}

	singular := toString(args[1])
	plural := toString(args[2])
	count, ok := toInt(args[3])
	if !ok {
		return nil, NewError(ErrorTypeTemplate, "npgettext() count must be numeric", nodes.Position{}, nil)
	}

	choice := plural
	if count == 1 {
		choice = singular
	}

	if len(args) > 4 {
		replacements := args[4]
		if mapping, ok := toStringInterfaceMap(replacements); ok {
			mapping["count"] = count
			return formatWithMap(choice, mapping), nil
		}
		vals := append([]interface{}{count}, args[5:]...)
		return fmt.Sprintf(choice, vals...), nil
	}

	return fmt.Sprintf(choice, count), nil
}

func formatWithMap(message string, mapping map[string]interface{}) string {
	result := message
	for key, value := range mapping {
		token := fmt.Sprintf("%%(%s)s", key)
		replacement := toString(value)
		result = strings.ReplaceAll(result, token, replacement)
		result = strings.ReplaceAll(result, fmt.Sprintf("%%(%s)d", key), replacement)
		result = strings.ReplaceAll(result, fmt.Sprintf("%%(%s)i", key), replacement)
		result = strings.ReplaceAll(result, fmt.Sprintf("%%(%s)f", key), replacement)
		result = strings.ReplaceAll(result, fmt.Sprintf("%%(%s)g", key), replacement)
	}
	return result
}

// cycler implements a simple cycling iterator
type cycler struct {
	items []interface{}
	index int
}

func (c *cycler) Next() interface{} {
	item := c.items[c.index]
	c.index = (c.index + 1) % len(c.items)
	return item
}

func (c *cycler) Current() interface{} {
	return c.items[c.index]
}

func (c *cycler) Reset() {
	c.index = 0
}

// joiner implements a simple string joiner
type joiner struct {
	separator string
	first     bool
}

func (j *joiner) Next() string {
	if j.first {
		j.first = false
		return ""
	}
	return j.separator
}
