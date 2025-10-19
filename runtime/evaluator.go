package runtime

import (
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/deicod/gojinja/nodes"
)

type continueSignal struct{}
type breakSignal struct{}

func isControlSignal(value interface{}) (interface{}, bool) {
	switch value.(type) {
	case continueSignal, breakSignal:
		return value, true
	default:
		return nil, false
	}
}

func controlName(signal interface{}) string {
	switch signal.(type) {
	case continueSignal:
		return "continue"
	case breakSignal:
		return "break"
	default:
		return "control"
	}
}

// Markup represents a string that should not be HTML-escaped
type Markup string

// Evaluator implements the visitor pattern for evaluating AST nodes
type Evaluator struct {
	ctx            *Context
	securityCtx    *SecurityContext
	securityChecks bool
}

var (
	spacelessBetweenTags = regexp.MustCompile(`>\s+<`)
	transWhitespace      = regexp.MustCompile(`\s+`)
)

// NewEvaluator creates a new evaluator
func NewEvaluator(ctx *Context) *Evaluator {
	return &Evaluator{ctx: ctx}
}

// NewSecureEvaluator creates a new evaluator with security checks
func NewSecureEvaluator(ctx *Context, secCtx *SecurityContext) *Evaluator {
	return &Evaluator{
		ctx:            ctx,
		securityCtx:    secCtx,
		securityChecks: true,
	}
}

// Evaluate evaluates a node and returns the result
func (e *Evaluator) Evaluate(node nodes.Node) interface{} {
	if node == nil {
		return nil
	}

	// Perform security checks if enabled
	if e.securityChecks && e.securityCtx != nil {
		if !e.performSecurityChecks(node) {
			return fmt.Errorf("security violation during evaluation")
		}
	}

	return node.Accept(e)
}

// performSecurityChecks performs security checks on a node
func (e *Evaluator) performSecurityChecks(node nodes.Node) bool {
	// Check recursion limit
	templateName := "unknown"
	if e.ctx.current != nil {
		templateName = e.ctx.current.name
	}

	if !e.securityCtx.CheckRecursionLimit(templateName) {
		return false
	}

	// Check execution time
	if !e.securityCtx.CheckExecutionTime(templateName) {
		return false
	}

	// Node-specific security checks
	switch n := node.(type) {
	case *nodes.Filter:
		return e.securityCtx.CheckFilterAccess(n.Name, templateName, "filter_usage")
	case *nodes.Call:
		return e.checkCallSecurity(n)
	case *nodes.Getattr:
		return e.securityCtx.CheckAttributeAccess(n.Attr, templateName, "attribute_access")
	}

	return true
}

// checkCallSecurity performs security checks on function calls
func (e *Evaluator) checkCallSecurity(call *nodes.Call) bool {
	templateName := "unknown"
	if e.ctx.current != nil {
		templateName = e.ctx.current.name
	}

	// Check if this is a method call
	if _, ok := call.Node.(*nodes.Getattr); ok {
		// This is a method call
		methodName := e.extractMethodName(call)
		return e.securityCtx.CheckMethodCall(methodName, templateName, "method_call")
	}

	// Check function access
	functionName := e.extractFunctionName(call)
	return e.securityCtx.CheckFunctionAccess(functionName, templateName, "function_call")
}

// extractMethodName extracts method name from a call node
func (e *Evaluator) extractMethodName(call *nodes.Call) string {
	if getattr, ok := call.Node.(*nodes.Getattr); ok {
		return getattr.Attr
	}
	return "unknown"
}

// extractFunctionName extracts function name from a call node
func (e *Evaluator) extractFunctionName(call *nodes.Call) string {
	if name, ok := call.Node.(*nodes.Name); ok {
		return name.Name
	}
	return "unknown"
}

// Write writes content to the output writer
func (e *Evaluator) Write(content string) {
	if e.ctx.writer != nil {
		io.WriteString(e.ctx.writer, content)
	}
}

// Visit implements the Visitor interface
func (e *Evaluator) Visit(node nodes.Node) interface{} {
	switch n := node.(type) {
	case *nodes.Template:
		return e.visitTemplate(n)
	case *nodes.Output:
		return e.visitOutput(n)
	case *nodes.For:
		return e.visitFor(n)
	case *nodes.If:
		return e.visitIf(n)
	case *nodes.Block:
		return e.visitBlock(n)
	case *nodes.Extends:
		return e.visitExtends(n)
	case *nodes.Include:
		return e.visitInclude(n)
	case *nodes.Import:
		return e.visitImport(n)
	case *nodes.FromImport:
		return e.visitFromImport(n)
	case *nodes.Macro:
		return e.visitMacro(n)
	case *nodes.CallBlock:
		return e.visitCallBlock(n)
	case *nodes.FilterBlock:
		return e.visitFilterBlock(n)
	case *nodes.Spaceless:
		return e.visitSpaceless(n)
	case *nodes.With:
		return e.visitWith(n)
	case *nodes.Assign:
		return e.visitAssign(n)
	case *nodes.AssignBlock:
		return e.visitAssignBlock(n)
	case *nodes.Do:
		return e.visitDo(n)
	case *nodes.ExprStmt:
		return e.visitExprStmt(n)
	case *nodes.Continue:
		return e.visitContinue(n)
	case *nodes.Break:
		return e.visitBreak(n)
	case *nodes.Scope:
		return e.visitScope(n)
	case *nodes.Namespace:
		return e.visitNamespace(n)
	case *nodes.Trans:
		return e.visitTrans(n)

	// Expression nodes
	case *nodes.Name:
		return e.visitName(n)
	case *nodes.Const:
		return e.visitConst(n)
	case *nodes.TemplateData:
		return e.visitTemplateData(n)
	case *nodes.List:
		return e.visitList(n)
	case *nodes.Dict:
		return e.visitDict(n)
	case *nodes.Tuple:
		return e.visitTuple(n)
	case *nodes.BinExpr:
		return e.visitBinExpr(n)
	case *nodes.UnaryExpr:
		return e.visitUnaryExpr(n)
	case *nodes.Call:
		return e.visitCall(n)
	case *nodes.Getattr:
		return e.visitGetattr(n)
	case *nodes.Getitem:
		return e.visitGetitem(n)
	case *nodes.Slice:
		return e.visitSlice(n)
	case *nodes.Filter:
		return e.visitFilter(n)
	case *nodes.Test:
		return e.visitTest(n)
	case *nodes.Compare:
		return e.visitCompare(n)
	case *nodes.CondExpr:
		return e.visitCondExpr(n)
	case *nodes.Concat:
		return e.visitConcat(n)
	case *nodes.Pair:
		return e.visitPair(n)
	case *nodes.Keyword:
		return e.visitKeyword(n)
	case *nodes.FilterTestCommon:
		return e.visitFilterTestCommon(n)

	// Extension nodes
	case *nodes.MarkSafe:
		return e.visitMarkSafe(n)
	case *nodes.MarkSafeIfAutoescape:
		return e.visitMarkSafeIfAutoescape(n)
	case *nodes.ContextReference:
		return e.visitContextReference(n)
	case *nodes.DerivedContextReference:
		return e.visitDerivedContextReference(n)

	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("unknown node type: %T", node), node.GetPosition(), node)
	}
}

// Statement node visitors

func (e *Evaluator) visitTemplate(node *nodes.Template) interface{} {
	for _, child := range node.Body {
		if result := e.Evaluate(child); result != nil {
			if err, ok := result.(error); ok {
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				return NewError(ErrorTypeTemplate,
					fmt.Sprintf("%s statement not allowed outside of a loop", controlName(signal)),
					child.GetPosition(), child)
			}
		}
	}
	return nil
}

func (e *Evaluator) visitOutput(node *nodes.Output) interface{} {
	for _, expr := range node.Nodes {
		value := e.Evaluate(expr)
		if err, ok := value.(error); ok {
			return err
		}

		if _, isTemplateData := expr.(*nodes.TemplateData); !isTemplateData {
			finalized, err := e.finalizeValue(value)
			if err != nil {
				return err
			}
			value = finalized
		}

		// Check if this is a TemplateData node that should not be escaped
		if _, ok := expr.(*nodes.TemplateData); ok {
			// TemplateData should be written directly without escaping
			str := e.toString(value, node.GetPosition())
			e.Write(str)
		} else if markup, ok := value.(Markup); ok {
			// Markup strings are safe and should not be escaped
			e.Write(string(markup))
		} else {
			// Convert other values to string and apply autoescaping
			str := e.toString(value, node.GetPosition())
			if e.ctx.ShouldAutoescape() {
				str = e.escape(str)
			}
			e.Write(str)
		}
	}
	return nil
}

func (e *Evaluator) visitFor(node *nodes.For) interface{} {
	// Evaluate the iterable
	iterable := e.Evaluate(node.Iter)
	if err, ok := iterable.(error); ok {
		return err
	}

	// Convert to slice
	items, err := e.toSlice(iterable, node.GetPosition())
	if err != nil {
		return err
	}

	// Handle empty iteration
	if len(items) == 0 {
		if len(node.Else) > 0 {
			for _, stmt := range node.Else {
				if result := e.Evaluate(stmt); result != nil {
					if err, ok := result.(error); ok {
						return err
					}
					if signal, ok := isControlSignal(result); ok {
						return signal
					}
				}
			}
		}
		return nil
	}

	// Create new scope for the loop
	e.ctx.PushScope()
	defer e.ctx.PopScope()

	// Push loop context
	e.ctx.PushLoop(len(items), 1)
	defer e.ctx.PopLoop()

	// Iterate
	broken := false
outerLoop:
	for i, item := range items {
		// Update loop context
		var prevItem, nextItem interface{}
		if i > 0 {
			prevItem = items[i-1]
		}
		if i < len(items)-1 {
			nextItem = items[i+1]
		}
		e.ctx.UpdateLoop(i, item, prevItem, nextItem)

		// Assign loop variable
		if err := e.assignTarget(node.Target, item, node.GetPosition()); err != nil {
			return err
		}

		// Execute loop body
		for _, stmt := range node.Body {
			if result := e.Evaluate(stmt); result != nil {
				if err, ok := result.(error); ok {
					return err
				}
				if signal, ok := isControlSignal(result); ok {
					switch signal.(type) {
					case continueSignal:
						continue outerLoop
					case breakSignal:
						broken = true
						break outerLoop
					}
				}
			}
		}
	}

	if !broken && len(node.Else) > 0 {
		// Loop completed normally; execute else block
		for _, stmt := range node.Else {
			if result := e.Evaluate(stmt); result != nil {
				if err, ok := result.(error); ok {
					return err
				}
				if signal, ok := isControlSignal(result); ok {
					return signal
				}
			}
		}
	}

	return nil
}

func (e *Evaluator) visitIf(node *nodes.If) interface{} {
	// Evaluate test condition
	testValue := e.Evaluate(node.Test)
	if err, ok := testValue.(error); ok {
		return err
	}

	if e.isTruthy(testValue) {
		// Execute if body
		for _, stmt := range node.Body {
			if result := e.Evaluate(stmt); result != nil {
				if err, ok := result.(error); ok {
					return err
				}
				if signal, ok := isControlSignal(result); ok {
					return signal
				}
			}
		}
	} else {
		// Check elif conditions
		for _, elif := range node.Elif {
			elifTestValue := e.Evaluate(elif.Test)
			if err, ok := elifTestValue.(error); ok {
				return err
			}

			if e.isTruthy(elifTestValue) {
				for _, stmt := range elif.Body {
					if result := e.Evaluate(stmt); result != nil {
						if err, ok := result.(error); ok {
							return err
						}
						if signal, ok := isControlSignal(result); ok {
							return signal
						}
					}
				}
				return nil
			}
		}

		// Execute else body
		for _, stmt := range node.Else {
			if result := e.Evaluate(stmt); result != nil {
				if err, ok := result.(error); ok {
					return err
				}
				if signal, ok := isControlSignal(result); ok {
					return signal
				}
			}
		}
	}

	return nil
}

func (e *Evaluator) visitBlock(node *nodes.Block) interface{} {
	// Get inheritance context if available
	var inheritanceCtx *InheritanceContext
	if e.ctx.current != nil && e.ctx.current.inheritanceCtx != nil {
		inheritanceCtx = e.ctx.current.inheritanceCtx
	}

	// Push block context for inheritance
	if inheritanceCtx != nil {
		inheritanceCtx.PushBlock(node.Name)
		defer inheritanceCtx.PopBlock()
	}

	// Create new scope if block is scoped
	if node.Scoped {
		e.ctx.PushScope()
		defer e.ctx.PopScope()
	}

	// Execute block body
	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				return signal
			}
		}
	}

	return nil
}

func (e *Evaluator) visitExtends(node *nodes.Extends) interface{} {
	// Extends statements are handled during template compilation
	// They should not be executed during normal rendering
	return nil
}

func (e *Evaluator) visitInclude(node *nodes.Include) interface{} {
	if e.ctx == nil || e.ctx.environment == nil {
		return NewError(ErrorTypeTemplate, "no environment available for includes", node.GetPosition(), node)
	}

	templateNames, err := e.evaluateIncludeTargets(node.Template)
	if err != nil {
		return err
	}

	if len(templateNames) == 0 {
		if node.IgnoreMissing {
			return nil
		}
		return NewError(ErrorTypeTemplate, "include requires at least one template name", node.GetPosition(), node)
	}

	var lastErr error
	for _, name := range templateNames {
		tmpl, loadErr := e.ctx.environment.LoadTemplate(name)
		if loadErr != nil {
			if isTemplateNotFoundError(loadErr) {
				lastErr = loadErr
				continue
			}
			return loadErr
		}

		if renderErr := e.renderIncludedTemplate(tmpl, node.WithContext); renderErr != nil {
			return renderErr
		}

		return nil
	}

	if node.IgnoreMissing {
		return nil
	}

	if lastErr != nil {
		if len(templateNames) > 1 {
			return NewTemplatesNotFound(templateNames, templateNames, lastErr)
		}
		return lastErr
	}

	return NewError(ErrorTypeTemplate, "no templates found for include", node.GetPosition(), node)
}

func (e *Evaluator) evaluateIncludeTargets(expr nodes.Expr) ([]string, interface{}) {
	value := e.Evaluate(expr)
	if err, ok := value.(error); ok {
		return nil, err
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return []string{}, nil
		}
		return []string{v}, nil
	case []interface{}:
		names := make([]string, 0, len(v))
		for _, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, NewError(ErrorTypeTemplate, "include template list must contain only strings", expr.GetPosition(), expr)
			}
			if str != "" {
				names = append(names, str)
			}
		}
		return names, nil
	default:
		return nil, NewError(ErrorTypeTemplate, "include template must be a string or list of strings", expr.GetPosition(), expr)
	}
}

func (e *Evaluator) renderIncludedTemplate(tmpl *Template, withContext bool) error {
	if withContext {
		oldCurrent := e.ctx.current
		oldAutoescape := e.ctx.ShouldAutoescape()
		e.ctx.current = tmpl
		e.ctx.SetAutoescape(tmpl.Autoescape())
		e.ctx.PushScope()
		defer func() {
			e.ctx.PopScope()
			e.ctx.SetAutoescape(oldAutoescape)
			e.ctx.current = oldCurrent
		}()
		return tmpl.ExecuteWithContext(e.ctx)
	}

	includeCtx := NewContextWithEnvironment(e.ctx.environment, nil)
	includeCtx.SetAutoescape(tmpl.Autoescape())
	includeCtx.writer = e.ctx.writer
	includeCtx.current = tmpl
	return tmpl.ExecuteWithContext(includeCtx)
}

func isTemplateNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	var notFound *TemplateNotFoundError
	if errors.As(err, &notFound) {
		return true
	}

	var multiNotFound *TemplatesNotFoundError
	if errors.As(err, &multiNotFound) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "not found")
}

func (e *Evaluator) visitImport(node *nodes.Import) interface{} {
	// Evaluate template name
	templateNameValue := e.Evaluate(node.Template)
	if err, ok := templateNameValue.(error); ok {
		return err
	}

	templateName, ok := templateNameValue.(string)
	if !ok {
		return NewError(ErrorTypeTemplate, "import template name must be a string", node.GetPosition(), node)
	}

	// Import the template
	if e.ctx.environment == nil {
		return NewError(ErrorTypeTemplate, "no environment available for imports", node.GetPosition(), node)
	}

	importManager := NewImportManager(e.ctx.environment)
	namespace, err := importManager.ImportTemplate(e.ctx, templateName, node.WithContext)
	if err != nil {
		return NewImportError(templateName, err.Error(), node.GetPosition(), node)
	}

	// Store the namespace in the context
	e.ctx.Set(node.Target, namespace)

	return nil
}

func (e *Evaluator) visitFromImport(node *nodes.FromImport) interface{} {
	// Evaluate template name
	templateNameValue := e.Evaluate(node.Template)
	if err, ok := templateNameValue.(error); ok {
		return err
	}

	templateName, ok := templateNameValue.(string)
	if !ok {
		return NewError(ErrorTypeTemplate, "from import template name must be a string", node.GetPosition(), node)
	}

	// Import the template
	if e.ctx.environment == nil {
		return NewError(ErrorTypeTemplate, "no environment available for imports", node.GetPosition(), node)
	}

	importManager := NewImportManager(e.ctx.environment)

	// Prepare macro names to import
	macroNames := make([]string, len(node.Names))
	for i, importName := range node.Names {
		if importName.Alias != "" {
			macroNames[i] = fmt.Sprintf("%s as %s", importName.Name, importName.Alias)
		} else {
			macroNames[i] = importName.Name
		}
	}

	macros, err := importManager.ImportMacros(e.ctx, templateName, macroNames, node.WithContext)
	if err != nil {
		return NewImportError(templateName, err.Error(), node.GetPosition(), node)
	}

	// Store each macro in the current context
	for alias, macro := range macros {
		e.ctx.Set(alias, macro)
	}

	return nil
}

func (e *Evaluator) visitMacro(node *nodes.Macro) interface{} {
	// Create macro object
	macro := NewMacro(node, e.ctx.current)

	// Store macro in current scope and registry
	e.ctx.Set(node.Name, macro)

	// Register in template's macro registry
	if e.ctx.current != nil && e.ctx.environment != nil {
		registry := e.ctx.environment.GetMacroRegistry()
		if registry != nil {
			registry.RegisterTemplate(e.ctx.current.name, node.Name, macro)
		}
	}

	return nil
}

func (e *Evaluator) visitCallBlock(node *nodes.CallBlock) interface{} {
	callNode := node.Call
	callable := e.Evaluate(callNode.Node)
	if err, ok := callable.(error); ok {
		return err
	}
	baseCtx := e.ctx
	baseScope := e.ctx.scope

	args := make([]interface{}, len(callNode.Args))
	for i, argExpr := range callNode.Args {
		value := e.Evaluate(argExpr)
		if err, ok := value.(error); ok {
			return err
		}
		args[i] = value
	}

	kwargs := make(map[string]interface{})
	for _, kw := range callNode.Kwargs {
		value := e.Evaluate(kw.Value)
		if err, ok := value.(error); ok {
			return err
		}
		kwargs[kw.Key] = value
	}

	if callNode.DynArgs != nil {
		dynArgs := e.Evaluate(callNode.DynArgs)
		if err, ok := dynArgs.(error); ok {
			return err
		}
		if slice, ok := dynArgs.([]interface{}); ok {
			args = append(args, slice...)
		}
	}

	if callNode.DynKwargs != nil {
		dynKwargs := e.Evaluate(callNode.DynKwargs)
		if err, ok := dynKwargs.(error); ok {
			return err
		}
		if dict, ok := dynKwargs.(map[interface{}]interface{}); ok {
			for k, v := range dict {
				if key, ok := k.(string); ok {
					kwargs[key] = v
				}
			}
		}
	}

	callerFunc := GlobalFunc(func(callCtx *Context, args ...interface{}) (interface{}, error) {
		ctxForBlock := baseCtx
		if ctxForBlock == nil {
			ctxForBlock = e.ctx
		}

		vars := make(map[string]interface{})
		if baseScope != nil {
			vars = baseScope.All()
		}

		blockCtx := NewContextWithEnvironment(ctxForBlock.environment, vars)
		blockCtx.current = ctxForBlock.current
		blockCtx.SetAutoescape(ctxForBlock.ShouldAutoescape())

		var buf strings.Builder
		blockCtx.writer = &buf

		blockCtx.PushScope()
		defer blockCtx.PopScope()

		for i, param := range node.Args {
			if i < len(args) {
				blockCtx.Set(param.Name, args[i])
			} else if i < len(node.Defaults) {
				value := NewEvaluator(blockCtx).Evaluate(node.Defaults[i])
				if err, ok := value.(error); ok {
					return "", err
				}
				blockCtx.Set(param.Name, value)
			} else {
				blockCtx.Set(param.Name, nil)
			}
		}

		evaluator := NewEvaluator(blockCtx)
		for _, stmt := range node.Body {
			if result := evaluator.Evaluate(stmt); result != nil {
				if err, ok := result.(error); ok {
					return nil, err
				}
				if signal, ok := isControlSignal(result); ok {
					return signal, nil
				}
			}
		}
		return Markup(buf.String()), nil
	})

	e.ctx.PushScope()
	e.ctx.Set("caller", callerFunc)
	defer e.ctx.PopScope()

	if m, ok := callable.(*Macro); ok {
		m.callerFunc = callerFunc
		defer func() {
			m.callerFunc = nil
		}()
	}

	result := e.callFunction(callable, args, kwargs, node.GetPosition())
	if err, ok := result.(error); ok {
		return err
	}
	if signal, ok := isControlSignal(result); ok {
		return signal
	}

	if result != nil {
		finalized, err := e.finalizeValue(result)
		if err != nil {
			return err
		}
		result = finalized

		switch v := result.(type) {
		case Markup:
			e.Write(string(v))
		default:
			e.Write(e.toString(result, node.GetPosition()))
		}
	}

	return nil
}

func (e *Evaluator) visitFilterBlock(node *nodes.FilterBlock) interface{} {
	var buf strings.Builder
	oldWriter := e.ctx.writer
	e.ctx.writer = &buf

	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				e.ctx.writer = oldWriter
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				e.ctx.writer = oldWriter
				return signal
			}
		}
	}

	e.ctx.writer = oldWriter

	result := interface{}(buf.String())
	if node.Filter != nil {
		baseValue := &nodes.Const{Value: buf.String()}
		cloned := cloneFilterChain(node.Filter, baseValue)
		filterResult := e.visitFilter(cloned)
		if err, ok := filterResult.(error); ok {
			return err
		}
		result = filterResult
	}

	if signal, ok := isControlSignal(result); ok {
		return signal
	}

	if result != nil {
		finalized, err := e.finalizeValue(result)
		if err != nil {
			return err
		}
		result = finalized

		switch v := result.(type) {
		case Markup:
			e.Write(string(v))
		default:
			e.Write(e.toString(result, node.GetPosition()))
		}
	}

	return nil
}

func (e *Evaluator) visitSpaceless(node *nodes.Spaceless) interface{} {
	var buf strings.Builder
	oldWriter := e.ctx.writer
	e.ctx.writer = &buf

	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			e.ctx.writer = oldWriter
			if err, ok := result.(error); ok {
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				return signal
			}
		}
	}

	e.ctx.writer = oldWriter

	if oldWriter != nil {
		collapsed := applySpacelessTransform(buf.String())
		if _, err := io.WriteString(oldWriter, collapsed); err != nil {
			return err
		}
	}

	return nil
}

func applySpacelessTransform(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	return spacelessBetweenTags.ReplaceAllString(trimmed, "><")
}

func (e *Evaluator) finalizeValue(value interface{}) (interface{}, error) {
	if e.ctx == nil || e.ctx.environment == nil {
		return value, nil
	}
	return e.ctx.environment.applyFinalize(value)
}

func (e *Evaluator) visitWith(node *nodes.With) interface{} {
	// Create new scope
	e.ctx.PushScope()
	defer e.ctx.PopScope()

	// Assign targets with values
	for i, target := range node.Targets {
		if i < len(node.Values) {
			value := e.Evaluate(node.Values[i])
			if err, ok := value.(error); ok {
				return err
			}
			if err := e.assignTarget(target, value, node.GetPosition()); err != nil {
				return err
			}
		}
	}

	// Execute body
	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				return signal
			}
		}
	}

	return nil
}

func (e *Evaluator) visitAssign(node *nodes.Assign) interface{} {
	// Evaluate the expression
	value := e.Evaluate(node.Node)
	if err, ok := value.(error); ok {
		return err
	}

	// Assign to target
	return e.assignTarget(node.Target, value, node.GetPosition())
}

func (e *Evaluator) visitAssignBlock(node *nodes.AssignBlock) interface{} {
	var buf strings.Builder
	oldWriter := e.ctx.writer
	e.ctx.writer = &buf

	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				e.ctx.writer = oldWriter
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				e.ctx.writer = oldWriter
				return signal
			}
		}
	}

	e.ctx.writer = oldWriter

	value := interface{}(buf.String())
	if node.Filter != nil {
		baseValue := &nodes.Const{Value: buf.String()}
		cloned := cloneFilterChain(node.Filter, baseValue)
		filtered := e.visitFilter(cloned)
		if err, ok := filtered.(error); ok {
			return err
		}
		if signal, ok := isControlSignal(filtered); ok {
			return signal
		}
		value = filtered
	}

	return e.assignTarget(node.Target, value, node.GetPosition())
}

func (e *Evaluator) visitDo(node *nodes.Do) interface{} {
	result := e.Evaluate(node.Expr)
	if err, ok := result.(error); ok {
		return err
	}
	if signal, ok := isControlSignal(result); ok {
		return signal
	}
	return nil
}

func (e *Evaluator) visitExprStmt(node *nodes.ExprStmt) interface{} {
	// Evaluate expression and discard result
	return e.Evaluate(node.Node)
}

func (e *Evaluator) visitContinue(node *nodes.Continue) interface{} {
	return continueSignal{}
}

func (e *Evaluator) visitBreak(node *nodes.Break) interface{} {
	return breakSignal{}
}

func (e *Evaluator) visitScope(node *nodes.Scope) interface{} {
	// Create new scope
	e.ctx.PushScope()
	defer e.ctx.PopScope()

	// Execute body
	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				return err
			}
			if signal, ok := isControlSignal(result); ok {
				return signal
			}
		}
	}

	return nil
}

func (e *Evaluator) visitNamespace(node *nodes.Namespace) interface{} {
	var initial interface{}
	if node.Value != nil {
		initial = e.Evaluate(node.Value)
		if err, ok := initial.(error); ok {
			return err
		}
	}

	namespace, err := ensureNamespaceObject(node.Name, initial, node)
	if err != nil {
		return err
	}

	oldValue, oldExists := e.ctx.Get(node.Name)
	restore := func() {
		if oldExists {
			e.ctx.Set(node.Name, oldValue)
		} else {
			e.ctx.Delete(node.Name)
		}
	}

	e.ctx.Set(node.Name, namespace)

	e.ctx.PushScope()
	namespaceScope := e.ctx.scope

	var control interface{}
	var controlIsError bool
	for _, stmt := range node.Body {
		if result := e.Evaluate(stmt); result != nil {
			if err, ok := result.(error); ok {
				control = err
				controlIsError = true
				break
			}
			if signal, ok := isControlSignal(result); ok {
				control = signal
				break
			}
		}
	}

	e.ctx.PopScope()

	if control != nil {
		if controlIsError {
			restore()
			return control
		}
		return control
	}

	for name, value := range namespaceScope.vars {
		if name == node.Name {
			continue
		}
		namespace.Set(name, value)
	}

	e.ctx.Set(node.Name, namespace)
	return nil
}

func (e *Evaluator) visitTrans(node *nodes.Trans) interface{} {
	state := newTransPlaceholderState()

	base := make(map[string]interface{})
	for name, expr := range node.Variables {
		value := e.Evaluate(expr)
		if err, ok := value.(error); ok {
			return err
		}
		finalized, err := e.finalizeValue(value)
		if err != nil {
			return err
		}
		base[name] = finalized
		state.reserve(name)
	}

	var countValue interface{}
	countName := node.CountName
	if countName == "" {
		countName = "count"
	}
	if node.CountExpr != nil {
		value := e.Evaluate(node.CountExpr)
		if err, ok := value.(error); ok {
			return err
		}
		finalized, err := e.finalizeValue(value)
		if err != nil {
			return err
		}
		countValue = finalized
		base[countName] = finalized
		state.reserve(countName)
		if countName != "count" {
			state.reserve("count")
		}
	}

	singularMsg, singularVars, err := e.renderTransBody(node.Singular, base, state)
	if err != nil {
		return err
	}

	trimmed := node.Trimmed
	if !node.TrimmedSet && e.ctx.environment != nil {
		e.ctx.environment.mu.RLock()
		if policy, ok := e.ctx.environment.policies["ext.i18n.trimmed"]; ok {
			if val, ok := policy.(bool); ok {
				trimmed = val
			}
		}
		e.ctx.environment.mu.RUnlock()
	}
	if trimmed {
		singularMsg = trimTransWhitespace(singularMsg)
	}

	allVars := singularVars

	if node.CountExpr != nil && len(node.Plural) > 0 {
		pluralMsg, pluralVars, err := e.renderTransBody(node.Plural, allVars, state)
		if err != nil {
			return err
		}
		for k, v := range pluralVars {
			allVars[k] = v
		}
		if trimmed {
			pluralMsg = trimTransWhitespace(pluralMsg)
		}
		if countValue != nil {
			allVars[countName] = countValue
			if countName != "count" {
				if _, exists := allVars["count"]; !exists {
					allVars["count"] = countValue
				}
			}
		}

		result, err := e.invokeNGettext(node, singularMsg, pluralMsg, countValue, allVars)
		if err != nil {
			return err
		}

		finalized, err := e.finalizeValue(result)
		if err != nil {
			return err
		}
		output := e.toString(finalized, node.GetPosition())
		if e.ctx.ShouldAutoescape() {
			output = e.escape(output)
		}
		e.Write(output)
		return nil
	}

	result, err := e.invokeGettext(node, singularMsg, allVars)
	if err != nil {
		return err
	}
	finalized, err := e.finalizeValue(result)
	if err != nil {
		return err
	}
	output := e.toString(finalized, node.GetPosition())
	if e.ctx.ShouldAutoescape() {
		output = e.escape(output)
	}
	e.Write(output)
	return nil
}

func (e *Evaluator) renderTransBody(body []nodes.Node, base map[string]interface{}, state *transPlaceholderState) (string, map[string]interface{}, error) {
	mapping := make(map[string]interface{}, len(base))
	for k, v := range base {
		mapping[k] = v
	}

	var builder strings.Builder

	for _, child := range body {
		switch n := child.(type) {
		case *nodes.Output:
			for _, expr := range n.Nodes {
				if data, ok := expr.(*nodes.TemplateData); ok {
					builder.WriteString(data.Data)
					continue
				}

				placeholder := state.nameFor(expr)
				if _, exists := mapping[placeholder]; !exists {
					value := e.Evaluate(expr)
					if err, ok := value.(error); ok {
						return "", nil, err
					}
					finalized, err := e.finalizeValue(value)
					if err != nil {
						return "", nil, err
					}
					mapping[placeholder] = finalized
				}
				builder.WriteString("%(" + placeholder + ")s")
			}
		case *nodes.TemplateData:
			builder.WriteString(n.Data)
		default:
			value := e.Evaluate(n)
			if err, ok := value.(error); ok {
				return "", nil, err
			}
			finalized, err := e.finalizeValue(value)
			if err != nil {
				return "", nil, err
			}
			builder.WriteString(e.toString(finalized, n.GetPosition()))
		}
	}

	return builder.String(), mapping, nil
}

func (e *Evaluator) invokeGettext(node *nodes.Trans, message string, mapping map[string]interface{}) (interface{}, error) {
	if node.HasContext {
		if result, handled, err := e.callTransFunction(node, "pgettext", []interface{}{node.Context, message}, mapping); handled {
			return result, err
		} else if err != nil {
			return nil, err
		}
	}

	callable, err := e.resolveTransCallable("_", "gettext")
	if err != nil {
		return nil, err
	}
	if callable == nil {
		if len(mapping) == 0 {
			return message, nil
		}
		return formatWithMap(message, mapping), nil
	}

	args := []interface{}{message}
	if len(mapping) > 0 {
		args = append(args, mapping)
	}

	result := e.callFunction(callable, args, nil, node.GetPosition())
	if err, ok := result.(error); ok {
		return nil, err
	}
	return result, nil
}

func (e *Evaluator) invokeNGettext(node *nodes.Trans, singular, plural string, count interface{}, mapping map[string]interface{}) (interface{}, error) {
	if node.HasContext {
		if result, handled, err := e.callTransFunction(node, "npgettext", []interface{}{node.Context, singular, plural, count}, mapping); handled {
			return result, err
		} else if err != nil {
			return nil, err
		}
	}

	callable, err := e.resolveTransCallable("ngettext")
	if err != nil {
		return nil, err
	}

	if callable != nil {
		args := []interface{}{singular, plural, count}
		if len(mapping) > 0 {
			args = append(args, mapping)
		}
		result := e.callFunction(callable, args, nil, node.GetPosition())
		if err, ok := result.(error); ok {
			return nil, err
		}
		return result, nil
	}

	selected := plural
	if count != nil {
		if c, ok := toInt(count); ok && c == 1 {
			selected = singular
		}
	}

	if mapping == nil {
		mapping = make(map[string]interface{})
	}
	if _, exists := mapping["count"]; !exists {
		mapping["count"] = count
	}

	return formatWithMap(selected, mapping), nil
}

func (e *Evaluator) resolveTransCallable(names ...string) (interface{}, error) {
	for _, name := range names {
		value, err := e.ctx.Resolve(name)
		if err != nil {
			return nil, err
		}
		if value == nil || isUndefinedValue(value) {
			continue
		}
		return value, nil
	}
	return nil, nil
}

func (e *Evaluator) callTransFunction(node *nodes.Trans, name string, args []interface{}, mapping map[string]interface{}) (interface{}, bool, error) {
	callable, err := e.resolveTransCallable(name)
	if err != nil {
		return nil, true, err
	}
	if callable == nil {
		return nil, false, nil
	}

	callArgs := append([]interface{}{}, args...)
	if len(mapping) > 0 {
		callArgs = append(callArgs, mapping)
	}

	result := e.callFunction(callable, callArgs, nil, node.GetPosition())
	if err, ok := result.(error); ok {
		return nil, true, err
	}
	return result, true, nil
}

type transPlaceholderState struct {
	exprToName map[string]string
	nameToExpr map[string]string
	reserved   map[string]bool
	counter    int
}

func newTransPlaceholderState() *transPlaceholderState {
	return &transPlaceholderState{
		exprToName: make(map[string]string),
		nameToExpr: make(map[string]string),
		reserved:   make(map[string]bool),
	}
}

func (s *transPlaceholderState) reserve(name string) {
	if name == "" {
		return
	}
	s.reserved[name] = true
}

func (s *transPlaceholderState) nameFor(expr nodes.Expr) string {
	key := expr.String()
	if name, ok := s.exprToName[key]; ok {
		return name
	}

	candidate := sanitizePlaceholderName(deriveTransCandidate(expr))
	if candidate != "" {
		if existing, ok := s.nameToExpr[candidate]; ok {
			if existing == key {
				s.exprToName[key] = candidate
				return candidate
			}
		} else {
			s.nameToExpr[candidate] = key
			s.exprToName[key] = candidate
			return candidate
		}
	}

	for {
		s.counter++
		generated := fmt.Sprintf("value%d", s.counter)
		if s.reserved[generated] {
			continue
		}
		if _, exists := s.nameToExpr[generated]; exists {
			continue
		}
		s.nameToExpr[generated] = key
		s.exprToName[key] = generated
		return generated
	}
}

func deriveTransCandidate(expr nodes.Expr) string {
	switch n := expr.(type) {
	case *nodes.Name:
		return n.Name
	case *nodes.Getattr:
		return n.Attr
	case *nodes.Getitem:
		return "item"
	case *nodes.Const:
		return "value"
	default:
		return ""
	}
}

func sanitizePlaceholderName(name string) string {
	if name == "" {
		return ""
	}

	var builder strings.Builder
	for i, r := range name {
		if unicode.IsLetter(r) || r == '_' || (i > 0 && unicode.IsDigit(r)) {
			builder.WriteRune(r)
			continue
		}
		if i == 0 {
			builder.WriteRune('_')
		} else {
			builder.WriteRune('_')
		}
	}

	return builder.String()
}

func trimTransWhitespace(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	return transWhitespace.ReplaceAllString(trimmed, " ")
}

func ensureNamespaceObject(name string, value interface{}, node *nodes.Namespace) (*Namespace, error) {
	if value == nil {
		return NewNamespace(nil), nil
	}

	if _, ok := value.(undefinedType); ok {
		return NewNamespace(nil), nil
	}

	if ns, ok := value.(*Namespace); ok {
		return ns, nil
	}

	if mapping, ok := toStringInterfaceMap(value); ok {
		return NewNamespace(mapping), nil
	}

	message := fmt.Sprintf("namespace '%s' expects a namespace or mapping, got %T", name, value)
	return nil, NewError(ErrorTypeTemplate, message, node.GetPosition(), node)
}

// Expression node visitors

func (e *Evaluator) visitName(node *nodes.Name) interface{} {
	// First try to resolve normally
	value, err := e.ctx.Resolve(node.Name)
	if err != nil {
		return err
	}

	// If not found, try to resolve as namespaced macro
	if value == nil {
		if node.Name == "caller" {
			if macro := e.ctx.CurrentMacro(); macro != nil && macro.callerFunc != nil {
				return macro.callerFunc
			}
			if caller := e.ctx.CurrentCaller(); caller != nil {
				return caller
			}
		}
		if e.ctx.environment != nil {
			registry := e.ctx.environment.GetMacroRegistry()
			if registry != nil {
				macro, err := registry.ResolveMacroPath(e.ctx, node.Name)
				if err == nil {
					return macro
				}
			}
		}
		// Return undefined error
		return NewUndefinedError(node.Name, node.GetPosition(), node)
	}

	return value
}

func (e *Evaluator) visitConst(node *nodes.Const) interface{} {
	return node.Value
}

func (e *Evaluator) visitTemplateData(node *nodes.TemplateData) interface{} {
	return node.Data
}

func (e *Evaluator) visitList(node *nodes.List) interface{} {
	items := make([]interface{}, len(node.Items))
	for i, item := range node.Items {
		value := e.Evaluate(item)
		if err, ok := value.(error); ok {
			return err
		}
		items[i] = value
	}
	return items
}

func (e *Evaluator) visitDict(node *nodes.Dict) interface{} {
	result := make(map[interface{}]interface{})
	for _, pair := range node.Items {
		key := e.Evaluate(pair.Key)
		if err, ok := key.(error); ok {
			return err
		}

		value := e.Evaluate(pair.Value)
		if err, ok := value.(error); ok {
			return err
		}

		result[key] = value
	}
	return result
}

func (e *Evaluator) visitTuple(node *nodes.Tuple) interface{} {
	items := make([]interface{}, len(node.Items))
	for i, item := range node.Items {
		value := e.Evaluate(item)
		if err, ok := value.(error); ok {
			return err
		}
		items[i] = value
	}
	return items
}

func (e *Evaluator) visitBinExpr(node *nodes.BinExpr) interface{} {
	left := e.Evaluate(node.Left)
	if err, ok := left.(error); ok {
		return err
	}

	right := e.Evaluate(node.Right)
	if err, ok := right.(error); ok {
		return err
	}

	switch node.Operator {
	case "+":
		return e.add(left, right, node.GetPosition())
	case "-":
		return e.subtract(left, right, node.GetPosition())
	case "*":
		return e.multiply(left, right, node.GetPosition())
	case "/":
		return e.divide(left, right, node.GetPosition())
	case "//":
		return e.floorDivide(left, right, node.GetPosition())
	case "%":
		return e.modulo(left, right, node.GetPosition())
	case "**":
		return e.power(left, right, node.GetPosition())
	case "and":
		return e.logicalAnd(left, right)
	case "or":
		return e.logicalOr(left, right)
	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("unknown binary operator: %s", node.Operator), node.GetPosition(), node)
	}
}

func (e *Evaluator) visitUnaryExpr(node *nodes.UnaryExpr) interface{} {
	operand := e.Evaluate(node.Node)
	if err, ok := operand.(error); ok {
		return err
	}

	switch node.Operator {
	case "-":
		return e.negate(operand, node.GetPosition())
	case "+":
		return operand
	case "not":
		return e.logicalNot(operand)
	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("unknown unary operator: %s", node.Operator), node.GetPosition(), node)
	}
}

func (e *Evaluator) visitCall(node *nodes.Call) interface{} {
	// Evaluate the callable
	callable := e.Evaluate(node.Node)
	if err, ok := callable.(error); ok {
		return err
	}

	// Evaluate arguments
	args := make([]interface{}, len(node.Args))
	for i, arg := range node.Args {
		value := e.Evaluate(arg)
		if err, ok := value.(error); ok {
			return err
		}
		args[i] = value
	}

	// Evaluate keyword arguments
	kwargs := make(map[string]interface{})
	for _, kwarg := range node.Kwargs {
		value := e.Evaluate(kwarg.Value)
		if err, ok := value.(error); ok {
			return err
		}
		kwargs[kwarg.Key] = value
	}

	// Handle dynamic arguments
	if node.DynArgs != nil {
		dynArgs := e.Evaluate(node.DynArgs)
		if err, ok := dynArgs.(error); ok {
			return err
		}
		if slice, ok := dynArgs.([]interface{}); ok {
			args = append(args, slice...)
		}
	}

	if node.DynKwargs != nil {
		dynKwargs := e.Evaluate(node.DynKwargs)
		if err, ok := dynKwargs.(error); ok {
			return err
		}
		if dict, ok := dynKwargs.(map[interface{}]interface{}); ok {
			for k, v := range dict {
				if key, ok := k.(string); ok {
					kwargs[key] = v
				}
			}
		}
	}

	return e.callFunction(callable, args, kwargs, node.GetPosition())
}

func (e *Evaluator) visitGetattr(node *nodes.Getattr) interface{} {
	obj := e.Evaluate(node.Node)
	if err, ok := obj.(error); ok {
		return err
	}

	if node.Ctx == nodes.CtxStore {
		// Assignment to attribute
		return NewError(ErrorTypeTemplate, "attribute assignment not yet implemented", node.GetPosition(), node)
	}

	value, err := e.ctx.ResolveAttribute(obj, node.Attr)
	if err != nil {
		return err
	}

	return value
}

func (e *Evaluator) visitGetitem(node *nodes.Getitem) interface{} {
	obj := e.Evaluate(node.Node)
	if err, ok := obj.(error); ok {
		return err
	}

	index := e.Evaluate(node.Arg)
	if err, ok := index.(error); ok {
		return err
	}

	if node.Ctx == nodes.CtxStore {
		// Assignment to index
		return NewError(ErrorTypeTemplate, "index assignment not yet implemented", node.GetPosition(), node)
	}

	value, err := e.ctx.ResolveIndex(obj, index)
	if err != nil {
		return err
	}

	return value
}

func (e *Evaluator) visitSlice(node *nodes.Slice) interface{} {
	start := e.evaluateOptionalExpr(node.Start)
	stop := e.evaluateOptionalExpr(node.Stop)
	step := e.evaluateOptionalExpr(node.Step)

	return e.createSlice(start, stop, step, node.GetPosition())
}

func (e *Evaluator) visitFilter(node *nodes.Filter) interface{} {
	// Evaluate the input value
	input := e.Evaluate(node.Node)
	if err, ok := input.(error); ok {
		if IsUndefinedError(err) {
			if strings.EqualFold(node.Name, "default") {
				input = undefinedSentinel
			} else {
				return err
			}
		} else {
			return err
		}
	}

	// Evaluate filter arguments
	args := make([]interface{}, len(node.Args))
	for i, arg := range node.Args {
		value := e.Evaluate(arg)
		if err, ok := value.(error); ok {
			return err
		}
		args[i] = value
	}

	if len(node.Kwargs) > 0 || node.DynKwargs != nil {
		kwargs := make(map[string]interface{})
		for _, kwarg := range node.Kwargs {
			value := e.Evaluate(kwarg.Value)
			if err, ok := value.(error); ok {
				return err
			}
			keyValue := e.Evaluate(kwarg.Key)
			if err, ok := keyValue.(error); ok {
				return err
			}
			keyStr := toString(keyValue)
			if keyStr == "" {
				return NewFilterError(node.Name, "invalid keyword argument name", node.GetPosition(), node, nil)
			}
			kwargs[keyStr] = value
		}
		if node.DynKwargs != nil {
			value := e.Evaluate(node.DynKwargs)
			if err, ok := value.(error); ok {
				return err
			}
			if dict, ok := value.(map[interface{}]interface{}); ok {
				for k, v := range dict {
					if key, ok := k.(string); ok {
						kwargs[key] = v
					}
				}
			}
		}
		if len(kwargs) > 0 {
			args = append(args, kwargs)
		}
	}

	// Get filter function
	filterFunc, ok := e.ctx.environment.GetFilter(node.Name)
	if !ok {
		return NewFilterError(node.Name, "unknown filter", node.GetPosition(), node, nil)
	}

	// Apply filter
	result, err := filterFunc(e.ctx, input, args...)
	if err != nil {
		return NewFilterError(node.Name, err.Error(), node.GetPosition(), node, err)
	}

	return result
}

func (e *Evaluator) visitTest(node *nodes.Test) interface{} {
	// Evaluate the input value
	input := e.Evaluate(node.Node)
	if err, ok := input.(error); ok {
		if IsUndefinedError(err) {
			switch strings.ToLower(node.Name) {
			case "defined":
				return false
			case "undefined":
				return true
			default:
				return err
			}
		}
		return err
	}

	// Evaluate test arguments
	args := make([]interface{}, len(node.Args))
	for i, arg := range node.Args {
		value := e.Evaluate(arg)
		if err, ok := value.(error); ok {
			return err
		}
		args[i] = value
	}

	// Get test function
	testFunc, ok := e.ctx.environment.GetTest(node.Name)
	if !ok {
		return NewTestError(node.Name, "unknown test", node.GetPosition(), node, nil)
	}

	// Apply test
	result, err := testFunc(e.ctx, input, args...)
	if err != nil {
		return NewTestError(node.Name, err.Error(), node.GetPosition(), node, err)
	}

	return result
}

func (e *Evaluator) visitCompare(node *nodes.Compare) interface{} {
	left := e.Evaluate(node.Expr)
	if err, ok := left.(error); ok {
		return err
	}

	for _, op := range node.Ops {
		if op.Op == "is" || op.Op == "isnot" {
			result := e.evaluateTestOperator(op, left)
			if err, ok := result.(error); ok {
				return err
			}
			if !result.(bool) {
				return false
			}
			continue
		}

		right := e.Evaluate(op.Expr)
		if err, ok := right.(error); ok {
			return err
		}

		result := e.compare(op.Op, left, right, op.GetPosition())
		if err, ok := result.(error); ok {
			return err
		}

		if !result.(bool) {
			return false
		}

		left = right
	}

	return true
}

func (e *Evaluator) visitCondExpr(node *nodes.CondExpr) interface{} {
	test := e.Evaluate(node.Test)
	if err, ok := test.(error); ok {
		return err
	}

	if e.isTruthy(test) {
		return e.Evaluate(node.Expr1)
	}

	return e.Evaluate(node.Expr2)
}

func (e *Evaluator) visitConcat(node *nodes.Concat) interface{} {
	var result strings.Builder
	for _, n := range node.Nodes {
		value := e.Evaluate(n)
		if err, ok := value.(error); ok {
			return err
		}
		result.WriteString(e.toString(value, node.GetPosition()))
	}
	return result.String()
}

func (e *Evaluator) visitPair(node *nodes.Pair) interface{} {
	key := e.Evaluate(node.Key)
	if err, ok := key.(error); ok {
		return err
	}

	value := e.Evaluate(node.Value)
	if err, ok := value.(error); ok {
		return err
	}

	return []interface{}{key, value}
}

func (e *Evaluator) visitKeyword(node *nodes.Keyword) interface{} {
	value := e.Evaluate(node.Value)
	if err, ok := value.(error); ok {
		return err
	}

	return []interface{}{node.Key, value}
}

func (e *Evaluator) visitMarkSafe(node *nodes.MarkSafe) interface{} {
	value := e.Evaluate(node.Expr)
	if err, ok := value.(error); ok {
		return err
	}

	// Mark as safe by wrapping in Markup
	return Markup(e.toString(value, node.GetPosition()))
}

func (e *Evaluator) visitMarkSafeIfAutoescape(node *nodes.MarkSafeIfAutoescape) interface{} {
	value := e.Evaluate(node.Expr)
	if err, ok := value.(error); ok {
		return err
	}

	// Mark as safe only if autoescaping is active
	if e.ctx.ShouldAutoescape() {
		// Wrap in Markup to prevent escaping
		return Markup(e.toString(value, node.GetPosition()))
	}

	return value
}

func (e *Evaluator) visitContextReference(node *nodes.ContextReference) interface{} {
	// Return the current context
	return e.ctx.scope.All()
}

func (e *Evaluator) visitDerivedContextReference(node *nodes.DerivedContextReference) interface{} {
	// Return the derived context including locals
	return e.ctx.scope.All()
}

func (e *Evaluator) visitFilterTestCommon(node *nodes.FilterTestCommon) interface{} {
	if node.IsFilter {
		return e.visitFilter(&nodes.Filter{FilterTestCommon: *node})
	} else {
		return e.visitTest(&nodes.Test{FilterTestCommon: *node})
	}
}

// Helper methods

func (e *Evaluator) toString(value interface{}, pos nodes.Position) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case undefinedType:
		str, err := v.ToString()
		if err != nil {
			return e.handleUndefinedStringError(err, pos)
		}
		return str
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (e *Evaluator) handleUndefinedStringError(err error, pos nodes.Position) string {
	if err == nil {
		return ""
	}
	if undefErr, ok := err.(*UndefinedError); ok {
		err = NewUndefinedError(undefErr.Name, pos, nil)
	}
	if e.ctx != nil {
		e.ctx.AddError(err)
	}
	return ""
}

func (e *Evaluator) escape(value string) string {
	if e.ctx.environment != nil {
		return e.ctx.environment.escape(value)
	}
	// Fallback HTML escaping
	var buf strings.Builder
	for _, r := range value {
		switch r {
		case '<':
			buf.WriteString("&lt;")
		case '>':
			buf.WriteString("&gt;")
		case '&':
			buf.WriteString("&amp;")
		case '"':
			buf.WriteString("&quot;")
		case '\'':
			buf.WriteString("&#39;")
		default:
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

func (e *Evaluator) isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0
	case float32:
		return v != 0
	case string:
		return v != ""
	case []interface{}:
		return len(v) > 0
	case map[interface{}]interface{}:
		return len(v) > 0
	default:
		return true
	}
}

func (e *Evaluator) toSlice(value interface{}, pos nodes.Position) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case []string:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = item
		}
		return result, nil
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			result = append(result, item)
		}
		return result, nil
	case map[interface{}]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, item := range v {
			result = append(result, item)
		}
		return result, nil
	case string:
		// Convert string to slice of characters
		result := make([]interface{}, len(v))
		for i, r := range v {
			result[i] = string(r)
		}
		return result, nil
	default:
		// Use reflection to handle other types
		val := reflect.ValueOf(value)
		switch val.Kind() {
		case reflect.Slice, reflect.Array:
			result := make([]interface{}, val.Len())
			for i := 0; i < val.Len(); i++ {
				result[i] = val.Index(i).Interface()
			}
			return result, nil
		case reflect.Map:
			result := make([]interface{}, 0, val.Len())
			for _, key := range val.MapKeys() {
				result = append(result, val.MapIndex(key).Interface())
			}
			return result, nil
		}

		return nil, NewError(ErrorTypeTemplate, fmt.Sprintf("cannot convert %T to slice", value), pos, nil)
	}
}

func (e *Evaluator) assignTarget(target nodes.Expr, value interface{}, pos nodes.Position) error {
	switch t := target.(type) {
	case *nodes.Name:
		e.ctx.Set(t.Name, value)
		return nil
	case *nodes.Tuple:
		// Handle tuple assignment
		valueSlice, err := e.toSlice(value, pos)
		if err != nil {
			return err
		}

		if len(t.Items) != len(valueSlice) {
			return NewError(ErrorTypeAssignment, "tuple assignment size mismatch", pos, target)
		}

		for i, item := range t.Items {
			if err := e.assignTarget(item, valueSlice[i], pos); err != nil {
				return err
			}
		}
		return nil
	case *nodes.Getattr:
		container := e.Evaluate(t.Node)
		if err, ok := container.(error); ok {
			return err
		}
		return assignAttributeValue(container, t.Attr, value, pos, target)
	case *nodes.Getitem:
		container := e.Evaluate(t.Node)
		if err, ok := container.(error); ok {
			return err
		}
		idx := e.Evaluate(t.Arg)
		if err, ok := idx.(error); ok {
			return err
		}
		return assignIndexValue(container, idx, value, pos, target)
	case *nodes.NSRef:
		namespaceValue, exists := e.ctx.Get(t.Name)
		if !exists {
			return NewAssignmentError(target.String(), fmt.Sprintf("namespace '%s' is undefined", t.Name), pos, target)
		}

		if setter, ok := namespaceValue.(interface {
			Set(string, interface{}) interface{}
		}); ok {
			setter.Set(t.Attr, value)
			return nil
		}

		if ns, ok := namespaceValue.(*Namespace); ok {
			ns.Set(t.Attr, value)
			return nil
		}

		return NewAssignmentError(target.String(), fmt.Sprintf("'%s' is not a namespace", t.Name), pos, target)
	default:
		return NewAssignmentError(target.String(), "invalid assignment target", pos, target)
	}
}

func (e *Evaluator) callFunction(callable interface{}, args []interface{}, kwargs map[string]interface{}, pos nodes.Position) interface{} {
	switch fn := callable.(type) {
	case *Macro:
		callerValue, hasCaller := kwargs["__caller"]
		if hasCaller {
			delete(kwargs, "__caller")
			if gf, ok := callerValue.(GlobalFunc); ok {
				fn.callerFunc = gf
			}
		}
		result, err := fn.Execute(e.ctx, args, kwargs)
		fn.callerFunc = nil
		if err != nil {
			return NewMacroError(fn.Name, err.Error(), pos, fn)
		}
		return result
	case *MacroNamespace:
		// This shouldn't happen directly, but handle gracefully
		return NewError(ErrorTypeTemplate, "macro namespace is not callable", pos, nil)
	case GlobalFunc:
		callArgs := appendCallArgs(args, kwargs)
		result, err := fn(e.ctx, callArgs...)
		if err != nil {
			return NewError(ErrorTypeTemplate, err.Error(), pos, nil)
		}
		return result
	case func(*Context, ...interface{}) (interface{}, error):
		callArgs := appendCallArgs(args, kwargs)
		result, err := fn(e.ctx, callArgs...)
		if err != nil {
			return NewError(ErrorTypeTemplate, err.Error(), pos, nil)
		}
		return result
	case func(*Context, ...interface{}) interface{}:
		callArgs := appendCallArgs(args, kwargs)
		return fn(e.ctx, callArgs...)
	case func(...interface{}) (interface{}, error):
		callArgs := appendCallArgs(args, kwargs)
		result, err := fn(callArgs...)
		if err != nil {
			return NewError(ErrorTypeTemplate, err.Error(), pos, nil)
		}
		return result
	case func(...interface{}) interface{}:
		callArgs := appendCallArgs(args, kwargs)
		return fn(callArgs...)
	default:
		// Try using reflection for method calls
		argsWithKw := appendCallArgs(args, kwargs)
		val := reflect.ValueOf(callable)
		if val.Kind() == reflect.Func {
			funcType := val.Type()

			// Check if function is variadic
			isVariadic := funcType.IsVariadic()

			// Build function arguments
			callArgs := make([]reflect.Value, 0)

			// Check if it needs a context parameter
			hasContext := false
			if funcType.NumIn() > 0 {
				firstParamType := funcType.In(0)
				if firstParamType == reflect.TypeOf((*Context)(nil)) {
					// Add context as first parameter
					callArgs = append(callArgs, reflect.ValueOf(e.ctx))
					hasContext = true
				}
			}

			// For variadic functions, we need to check expected params
			if isVariadic {
				// Get the expected number of non-variadic params
				numFixed := funcType.NumIn()
				if hasContext {
					numFixed-- // Context is already added
				}
				numFixed-- // One param is variadic

				// Add fixed args
				for i := 0; i < numFixed && i < len(argsWithKw); i++ {
					callArgs = append(callArgs, reflect.ValueOf(argsWithKw[i]))
				}

				// Add remaining args as variadic
				for i := numFixed; i < len(argsWithKw); i++ {
					callArgs = append(callArgs, reflect.ValueOf(argsWithKw[i]))
				}
			} else {
				// Non-variadic: add all args
				for _, arg := range argsWithKw {
					callArgs = append(callArgs, reflect.ValueOf(arg))
				}
			}

			// Call the function
			var results []reflect.Value
			if isVariadic {
				results = val.CallSlice(callArgs)
			} else {
				results = val.Call(callArgs)
			}

			// Handle return values
			if len(results) == 0 {
				return nil
			} else if len(results) == 1 {
				return results[0].Interface()
			} else if len(results) == 2 {
				// Assume (result, error) pattern
				if !results[1].IsNil() {
					err, ok := results[1].Interface().(error)
					if ok {
						return NewError(ErrorTypeTemplate, err.Error(), pos, nil)
					}
				}
				return results[0].Interface()
			}

			// Return first result if multiple
			return results[0].Interface()
		}

		// Check if it's a macro-like callable
		if isMacroCallable(callable) {
			result, err := callMacroCallable(e.ctx, callable, args, kwargs)
			if err != nil {
				return NewError(ErrorTypeTemplate, err.Error(), pos, nil)
			}
			return result
		}
		return NewError(ErrorTypeTemplate, fmt.Sprintf("'%T' object is not callable", callable), pos, nil)
	}
}

func appendCallArgs(args []interface{}, kwargs map[string]interface{}) []interface{} {
	if len(kwargs) == 0 {
		return args
	}
	argsWithKw := make([]interface{}, 0, len(args)+1)
	argsWithKw = append(argsWithKw, args...)
	argsWithKw = append(argsWithKw, kwargs)
	return argsWithKw
}

func (e *Evaluator) evaluateOptionalExpr(expr nodes.Expr) interface{} {
	if expr == nil {
		return nil
	}
	return e.Evaluate(expr)
}

func (e *Evaluator) createSlice(start, stop, step interface{}, pos nodes.Position) interface{} {
	// Convert to integers with defaults
	startInt := 0
	stopInt := 0
	stepInt := 1

	if start != nil {
		if s, ok := toInt(start); ok {
			startInt = s
		}
	}

	if stop != nil {
		if s, ok := toInt(stop); ok {
			stopInt = s
		}
	}

	if step != nil {
		if s, ok := toInt(step); ok {
			stepInt = s
		}
	}

	return map[string]interface{}{
		"start": startInt,
		"stop":  stopInt,
		"step":  stepInt,
	}
}

// Arithmetic operations

func (e *Evaluator) add(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			return l + r
		case int64:
			return int64(l) + r
		case float64:
			return float64(l) + r
		}
	case int64:
		switch r := right.(type) {
		case int:
			return l + int64(r)
		case int64:
			return l + r
		case float64:
			return float64(l) + r
		}
	case float64:
		switch r := right.(type) {
		case int:
			return l + float64(r)
		case int64:
			return l + float64(r)
		case float64:
			return l + r
		}
	case string:
		// String concatenation only works with strings
		if r, ok := right.(string); ok {
			return l + r
		}
		// Return error for string + non-string
		return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for +: %T and %T", left, right), pos, nil)
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for +: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) subtract(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			return l - r
		case int64:
			return int64(l) - r
		case float64:
			return float64(l) - r
		}
	case int64:
		switch r := right.(type) {
		case int:
			return l - int64(r)
		case int64:
			return l - r
		case float64:
			return float64(l) - r
		}
	case float64:
		switch r := right.(type) {
		case int:
			return l - float64(r)
		case int64:
			return l - float64(r)
		case float64:
			return l - r
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for -: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) multiply(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			return l * r
		case int64:
			return int64(l) * r
		case float64:
			return float64(l) * r
		}
	case int64:
		switch r := right.(type) {
		case int:
			return l * int64(r)
		case int64:
			return l * r
		case float64:
			return float64(l) * r
		}
	case float64:
		switch r := right.(type) {
		case int:
			return l * float64(r)
		case int64:
			return l * float64(r)
		case float64:
			return l * r
		}
	case string:
		// String repetition
		if count, ok := toInt(right); ok && count >= 0 {
			var result strings.Builder
			for i := 0; i < count; i++ {
				result.WriteString(l)
			}
			return result.String()
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for *: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) divide(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / float64(r)
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / float64(r)
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / r
		}
	case int64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / float64(r)
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / float64(r)
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return float64(l) / r
		}
	case float64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / float64(r)
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / float64(r)
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / r
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for /: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) floorDivide(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / r
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return int64(l) / r
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return math.Floor(float64(l) / r)
		}
	case int64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / int64(r)
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return l / r
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return math.Floor(float64(l) / r)
		}
	case float64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return math.Floor(l / float64(r))
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return math.Floor(l / float64(r))
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "division by zero", pos, nil)
			}
			return math.Floor(l / r)
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for //: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) modulo(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return l % r
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return int64(l) % r
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return math.Mod(float64(l), r)
		}
	case int64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return l % int64(r)
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return l % r
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return math.Mod(float64(l), r)
		}
	case float64:
		switch r := right.(type) {
		case int:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return math.Mod(l, float64(r))
		case int64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return math.Mod(l, float64(r))
		case float64:
			if r == 0 {
				return NewError(ErrorTypeTemplate, "modulo by zero", pos, nil)
			}
			return math.Mod(l, r)
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for %%: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) power(left, right interface{}, pos nodes.Position) interface{} {
	switch l := left.(type) {
	case int:
		switch r := right.(type) {
		case int:
			return int(math.Pow(float64(l), float64(r)))
		case int64:
			return int64(math.Pow(float64(l), float64(r)))
		case float64:
			return math.Pow(float64(l), r)
		}
	case int64:
		switch r := right.(type) {
		case int:
			return int64(math.Pow(float64(l), float64(r)))
		case int64:
			return int64(math.Pow(float64(l), float64(r)))
		case float64:
			return math.Pow(float64(l), r)
		}
	case float64:
		switch r := right.(type) {
		case int:
			return math.Pow(l, float64(r))
		case int64:
			return math.Pow(l, float64(r))
		case float64:
			return math.Pow(l, r)
		}
	}

	return NewError(ErrorTypeTemplate, fmt.Sprintf("unsupported operand types for **: %T and %T", left, right), pos, nil)
}

func (e *Evaluator) logicalAnd(left, right interface{}) interface{} {
	if !e.isTruthy(left) {
		return left
	}
	return right
}

func (e *Evaluator) logicalOr(left, right interface{}) interface{} {
	if e.isTruthy(left) {
		return left
	}
	return right
}

func (e *Evaluator) logicalNot(operand interface{}) interface{} {
	return !e.isTruthy(operand)
}

func (e *Evaluator) negate(operand interface{}, pos nodes.Position) interface{} {
	switch v := operand.(type) {
	case int:
		return -v
	case int64:
		return -v
	case float64:
		return -v
	case float32:
		return -v
	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("bad operand type for unary -: %T", operand), pos, nil)
	}
}

func (e *Evaluator) compare(op string, left, right interface{}, pos nodes.Position) interface{} {
	switch op {
	case "eq", "==":
		if eq, ok := numericEqual(left, right); ok {
			return eq
		}
		return left == right
	case "ne", "!=":
		if eq, ok := numericEqual(left, right); ok {
			return !eq
		}
		return left != right
	case "lt", "<":
		return e.compareValues(left, right) < 0
	case "lteq", "<=":
		return e.compareValues(left, right) <= 0
	case "gt", ">":
		return e.compareValues(left, right) > 0
	case "gteq", ">=":
		return e.compareValues(left, right) >= 0
	case "in":
		return e.isInCollection(left, right)
	case "notin":
		return !e.isInCollection(left, right)
	default:
		return NewError(ErrorTypeTemplate, fmt.Sprintf("unknown comparison operator: %s", op), pos, nil)
	}
}

func (e *Evaluator) compareValues(left, right interface{}) int {
	leftVal, leftOk := toFloat64(left)
	rightVal, rightOk := toFloat64(right)

	if leftOk && rightOk {
		if leftVal < rightVal {
			return -1
		} else if leftVal > rightVal {
			return 1
		}
		return 0
	}

	leftStr := e.toString(left, nodes.Position{})
	rightStr := e.toString(right, nodes.Position{})

	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	}
	return 0
}

func cloneFilterChain(filter *nodes.Filter, base nodes.Expr) *nodes.Filter {
	if filter == nil {
		return nil
	}

	cloned := *filter
	switch node := filter.Node.(type) {
	case *nodes.Filter:
		cloned.Node = cloneFilterChain(node, base)
	case nil:
		cloned.Node = base
	default:
		cloned.Node = node
	}

	return &cloned
}

func assignAttributeValue(container interface{}, attr string, value interface{}, pos nodes.Position, node nodes.Node) error {
	if container == nil {
		return NewAssignmentError(node.String(), "cannot assign attribute on nil", pos, node)
	}

	if setter, ok := container.(interface {
		Set(string, interface{}) interface{}
	}); ok {
		setter.Set(attr, value)
		return nil
	}

	val := reflect.ValueOf(container)
	if !val.IsValid() {
		return NewAssignmentError(node.String(), "invalid value for attribute assignment", pos, node)
	}

	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return NewAssignmentError(node.String(), "cannot assign attribute on nil pointer", pos, node)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		key, err := convertToType(attr, val.Type().Key(), pos, node)
		if err != nil {
			return err
		}
		converted, err := convertToType(value, val.Type().Elem(), pos, node)
		if err != nil {
			return err
		}
		val.SetMapIndex(key, converted)
		return nil
	case reflect.Struct:
		field := val.FieldByName(attr)
		if !field.IsValid() {
			field = val.FieldByName(strings.Title(attr))
		}
		if !field.IsValid() || !field.CanSet() {
			return NewAssignmentError(node.String(), fmt.Sprintf("cannot assign attribute '%s' on %T", attr, container), pos, node)
		}
		converted, err := convertToType(value, field.Type(), pos, node)
		if err != nil {
			return err
		}
		field.Set(converted)
		return nil
	default:
		return NewAssignmentError(node.String(), fmt.Sprintf("cannot assign attribute on %T", container), pos, node)
	}
}

func assignIndexValue(container interface{}, index interface{}, value interface{}, pos nodes.Position, node nodes.Node) error {
	if container == nil {
		return NewAssignmentError(node.String(), "cannot assign index on nil", pos, node)
	}

	val := reflect.ValueOf(container)
	if !val.IsValid() {
		return NewAssignmentError(node.String(), "invalid value for index assignment", pos, node)
	}

	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return NewAssignmentError(node.String(), "cannot assign index on nil pointer", pos, node)
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Map:
		key, err := convertToType(index, val.Type().Key(), pos, node)
		if err != nil {
			return err
		}
		converted, err := convertToType(value, val.Type().Elem(), pos, node)
		if err != nil {
			return err
		}
		val.SetMapIndex(key, converted)
		return nil
	case reflect.Slice, reflect.Array:
		idx, err := normalizeIndex(index, val.Len(), pos, node)
		if err != nil {
			return err
		}
		elem := val.Index(idx)
		if !elem.CanSet() {
			return NewAssignmentError(node.String(), "index not assignable", pos, node)
		}
		converted, err := convertToType(value, elem.Type(), pos, node)
		if err != nil {
			return err
		}
		elem.Set(converted)
		return nil
	default:
		return NewAssignmentError(node.String(), fmt.Sprintf("cannot assign index on %T", container), pos, node)
	}
}

func convertToType(src interface{}, targetType reflect.Type, pos nodes.Position, node nodes.Node) (reflect.Value, error) {
	if targetType == nil {
		return reflect.Value{}, NewAssignmentError(node.String(), "invalid assignment target type", pos, node)
	}

	if targetType.Kind() == reflect.Interface {
		if targetType.NumMethod() == 0 {
			if src == nil {
				return reflect.Zero(targetType), nil
			}
			return reflect.ValueOf(src), nil
		}
		if src == nil {
			return reflect.Zero(targetType), nil
		}
		val := reflect.ValueOf(src)
		if val.Type().Implements(targetType) {
			return val, nil
		}
		if val.Type().AssignableTo(targetType) {
			return val.Convert(targetType), nil
		}
		if val.Type().ConvertibleTo(targetType) {
			return val.Convert(targetType), nil
		}
		return reflect.Value{}, NewAssignmentError(node.String(), fmt.Sprintf("cannot convert %T to %s", src, targetType), pos, node)
	}

	if src == nil {
		switch targetType.Kind() {
		case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Interface:
			return reflect.Zero(targetType), nil
		default:
			return reflect.Value{}, NewAssignmentError(node.String(), fmt.Sprintf("cannot assign nil to %s", targetType), pos, node)
		}
	}

	val := reflect.ValueOf(src)
	if val.Type() == targetType {
		return val, nil
	}
	if val.Type().AssignableTo(targetType) {
		return val.Convert(targetType), nil
	}
	if val.Type().ConvertibleTo(targetType) {
		return val.Convert(targetType), nil
	}

	return reflect.Value{}, NewAssignmentError(node.String(), fmt.Sprintf("cannot convert %T to %s", src, targetType), pos, node)
}

func normalizeIndex(index interface{}, length int, pos nodes.Position, node nodes.Node) (int, error) {
	var idx int
	switch v := index.(type) {
	case int:
		idx = v
	case int64:
		idx = int(v)
	case float64:
		if math.Trunc(v) != v {
			return 0, NewAssignmentError(node.String(), fmt.Sprintf("non-integer index %v", v), pos, node)
		}
		idx = int(v)
	default:
		return 0, NewAssignmentError(node.String(), fmt.Sprintf("unsupported index type %T", index), pos, node)
	}

	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 || idx >= length {
		return 0, NewAssignmentError(node.String(), fmt.Sprintf("index %d out of range", idx), pos, node)
	}

	return idx, nil
}

func numericEqual(left, right interface{}) (bool, bool) {
	leftVal, leftOk := toFloat64(left)
	rightVal, rightOk := toFloat64(right)
	if leftOk && rightOk {
		return leftVal == rightVal, true
	}
	return false, false
}

func (e *Evaluator) evaluateTestOperator(op *nodes.Operand, value interface{}) interface{} {
	var testName string
	var argExprs []nodes.Expr
	var kwargExprs []*nodes.Keyword

	switch expr := op.Expr.(type) {
	case *nodes.Name:
		testName = expr.Name
	case *nodes.Call:
		nameNode, ok := expr.Node.(*nodes.Name)
		if !ok {
			return NewTestError("", "invalid test expression", op.GetPosition(), expr, nil)
		}
		testName = nameNode.Name
		argExprs = expr.Args
		kwargExprs = expr.Kwargs
		if expr.DynArgs != nil || expr.DynKwargs != nil {
			return NewTestError(testName, "dynamic test arguments not yet supported", op.GetPosition(), expr, nil)
		}
	case *nodes.Const:
		switch v := expr.Value.(type) {
		case bool:
			if v {
				testName = "true"
			} else {
				testName = "false"
			}
		case nil:
			testName = "none"
		default:
			return NewTestError("", "invalid test expression", op.GetPosition(), expr, nil)
		}
	default:
		return NewTestError("", "invalid test expression", op.GetPosition(), op.Expr, nil)
	}

	testFunc, ok := e.ctx.environment.GetTest(testName)
	if !ok {
		return NewTestError(testName, "unknown test", op.GetPosition(), op.Expr, nil)
	}

	args := make([]interface{}, len(argExprs))
	for i, argExpr := range argExprs {
		val := e.Evaluate(argExpr)
		if err, ok := val.(error); ok {
			return err
		}
		args[i] = val
	}

	if len(kwargExprs) > 0 {
		kwargs := make(map[string]interface{}, len(kwargExprs))
		for _, kwarg := range kwargExprs {
			value := e.Evaluate(kwarg.Value)
			if err, ok := value.(error); ok {
				return err
			}
			kwargs[kwarg.Key] = value
		}
		args = append(args, kwargs)
	}

	passed, err := testFunc(e.ctx, value, args...)
	if err != nil {
		return NewTestError(testName, err.Error(), op.GetPosition(), op.Expr, err)
	}

	if op.Op == "isnot" {
		passed = !passed
	}

	return passed
}

func (e *Evaluator) isInCollection(item, collection interface{}) bool {
	switch coll := collection.(type) {
	case []interface{}:
		for _, v := range coll {
			if v == item {
				return true
			}
		}
	case map[interface{}]interface{}:
		_, exists := coll[item]
		return exists
	case map[string]interface{}:
		if str, ok := item.(string); ok {
			_, exists := coll[str]
			return exists
		}
	case string:
		if str, ok := item.(string); ok {
			return strings.Contains(coll, str)
		}
	}
	return false
}
