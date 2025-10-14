package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/deicod/gojinja/nodes"
)

// ImportManager handles template imports and macro imports
type ImportManager struct {
	environment *Environment
	registry    *MacroRegistry

	// Import cache to avoid circular dependencies
	importStack []string
	importCache map[string]*Template

	// Thread safety
	mu sync.RWMutex
}

// NewImportManager creates a new import manager
func NewImportManager(env *Environment) *ImportManager {
	return &ImportManager{
		environment: env,
		registry:    env.GetMacroRegistry(),
		importStack: make([]string, 0),
		importCache: make(map[string]*Template),
	}
}

// ImportTemplate imports a template and returns a macro namespace
func (im *ImportManager) ImportTemplate(ctx *Context, templateName string, withContext bool) (*MacroNamespace, error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Check for circular imports
	for _, name := range im.importStack {
		if name == templateName {
			return nil, NewImportError(templateName, "circular import detected", nodes.Position{}, nil)
		}
	}

	// Check cache first
	if template, exists := im.importCache[templateName]; exists {
		return im.createNamespaceFromTemplate(ctx, templateName, template, withContext)
	}

	// Load the template
	template, err := im.environment.LoadTemplate(templateName)
	if err != nil {
		return nil, NewImportError(templateName, err.Error(), nodes.Position{}, nil)
	}

	// Cache the template
	im.importCache[templateName] = template

	// Create namespace
	return im.createNamespaceFromTemplate(ctx, templateName, template, withContext)
}

// ImportMacros imports specific macros from a template
func (im *ImportManager) ImportMacros(ctx *Context, templateName string, macroNames []string, withContext bool) (map[string]*Macro, error) {
	namespace, err := im.ImportTemplate(ctx, templateName, withContext)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*Macro)
	for _, macroName := range macroNames {
		macro, err := namespace.GetMacro(macroName)
		if err != nil {
			return nil, NewImportError(templateName,
				fmt.Sprintf("macro '%s' not found in template", macroName),
				nodes.Position{}, nil)
		}

		// Handle aliasing
		alias := macroName
		if strings.Contains(macroName, " as ") {
			parts := strings.Split(macroName, " as ")
			if len(parts) == 2 {
				alias = strings.TrimSpace(parts[1])
				macroName = strings.TrimSpace(parts[0])
				macro, err = namespace.GetMacro(macroName)
				if err != nil {
					return nil, NewImportError(templateName,
						fmt.Sprintf("macro '%s' not found in template", macroName),
						nodes.Position{}, nil)
				}
			}
		}

		result[alias] = macro
	}

	return result, nil
}

// createNamespaceFromTemplate creates a macro namespace from a template
func (im *ImportManager) createNamespaceFromTemplate(ctx *Context, templateName string, template *Template, withContext bool) (*MacroNamespace, error) {
	// Create import context
	var importContext *Context
	if withContext {
		// Create a copy of the current context
		importContext = NewContextWithEnvironment(im.environment, ctx.scope.All())
	} else {
		importContext = NewContextWithEnvironment(im.environment, nil)
	}

	// Set template reference
	importContext.current = template

	// Create namespace
	namespace := NewMacroNamespace(templateName, template)
	namespace.Context = importContext

	// Process the template to collect macros
	im.importStack = append(im.importStack, templateName)
	defer func() {
		im.importStack = im.importStack[:len(im.importStack)-1]
	}()

	// Execute template in a special mode to collect macros
	evaluator := NewEvaluator(importContext)

	// Create a buffer to capture output (we discard it)
	var buf strings.Builder
	oldWriter := importContext.writer
	importContext.writer = &buf
	defer func() { importContext.writer = oldWriter }()

	// Process the template to collect macro definitions
	im.collectMacrosFromAST(evaluator, template.AST(), namespace)

	return namespace, nil
}

// collectMacrosFromAST walks the AST and collects macro definitions
func (im *ImportManager) collectMacrosFromAST(evaluator *Evaluator, node nodes.Node, namespace *MacroNamespace) {
	switch n := node.(type) {
	case *nodes.Template:
		for _, child := range n.Body {
			im.collectMacrosFromAST(evaluator, child, namespace)
		}
	case *nodes.Macro:
		// Create macro from AST node
		macro := NewMacro(n, namespace.Template)
		namespace.AddMacro(macro.Name, macro)

		// Also register in the global registry
		im.registry.RegisterNamespace(fmt.Sprintf("%s.%s", namespace.Name, macro.Name), namespace)
	case *nodes.Import:
		// Handle nested imports
		templateNameValue := evaluator.Evaluate(n.Template)
		if templateNameStr, ok := templateNameValue.(string); ok {
			nestedNamespace, err := im.ImportTemplate(evaluator.ctx, templateNameStr, n.WithContext)
			if err == nil {
				// Merge nested namespace macros
				for name, macro := range nestedNamespace.Macros {
					namespace.AddMacro(fmt.Sprintf("%s.%s", n.Target, name), macro)
				}
			}
		}
	case *nodes.FromImport:
		// Handle from imports
		templateNameValue := evaluator.Evaluate(n.Template)
		if templateNameStr, ok := templateNameValue.(string); ok {
			importedMacros := make([]string, len(n.Names))
			for i, importName := range n.Names {
				if importName.Alias != "" {
					importedMacros[i] = fmt.Sprintf("%s as %s", importName.Name, importName.Alias)
				} else {
					importedMacros[i] = importName.Name
				}
			}

			macros, err := im.ImportMacros(evaluator.ctx, templateNameStr, importedMacros, n.WithContext)
			if err == nil {
				for alias, macro := range macros {
					namespace.AddMacro(alias, macro)
				}
			}
		}
	default:
		// Recursively process child nodes
		for _, child := range node.GetChildren() {
			im.collectMacrosFromAST(evaluator, child, namespace)
		}
	}
}

// ResolveMacroPath resolves a macro path (e.g., "namespace.macro" or just "macro")
func (im *ImportManager) ResolveMacroPath(ctx *Context, path string) (*Macro, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	// Check if it's a namespaced path
	if strings.Contains(path, ".") {
		parts := strings.Split(path, ".")
		if len(parts) == 2 {
			namespaceName, macroName := parts[0], parts[1]

			// Try to find in registered namespaces
			macro, err := im.registry.FindNamespaceMacro(namespaceName, macroName)
			if err == nil {
				return macro, nil
			}

			// Try to find in current context as a variable
			if ctx != nil {
				if namespaceValue, ok := ctx.Get(namespaceName); ok {
					if namespace, ok := namespaceValue.(*MacroNamespace); ok {
						return namespace.GetMacro(macroName)
					}
				}
			}
		}
	}

	// Try to find as a regular macro
	return im.registry.FindMacro(ctx, path)
}

// GetImportedNamespaces returns all imported namespaces in the current context
func (im *ImportManager) GetImportedNamespaces(ctx *Context) map[string]*MacroNamespace {
	im.mu.RLock()
	defer im.mu.RUnlock()

	result := make(map[string]*MacroNamespace)

	// Get all namespaces from the registry
	for name, namespace := range im.registry.namespaces {
		result[name] = namespace
	}

	return result
}

// ClearCache clears the import cache
func (im *ImportManager) ClearCache() {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.importCache = make(map[string]*Template)
	im.importStack = make([]string, 0)
}

// GetImportStats returns import statistics
func (im *ImportManager) GetImportStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"cached_templates": len(im.importCache),
		"current_stack":    len(im.importStack),
		"namespaces":       len(im.registry.namespaces),
	}
}

// RelativeImportResolver handles relative template imports
type RelativeImportResolver struct {
	basePath string
}

// NewRelativeImportResolver creates a new relative import resolver
func NewRelativeImportResolver(basePath string) *RelativeImportResolver {
	return &RelativeImportResolver{basePath: basePath}
}

// Resolve resolves a relative template path
func (rir *RelativeImportResolver) Resolve(templatePath string) string {
	if strings.HasPrefix(templatePath, "./") || strings.HasPrefix(templatePath, "../") {
		// Join with base path
		resolved := filepath.Join(rir.basePath, templatePath)
		// Clean the path
		return filepath.Clean(resolved)
	}
	return templatePath
}

// MacroImport represents a macro import configuration
type MacroImport struct {
	Template     string
	Macros       []string
	Alias        string
	WithContext  bool
	Position     nodes.Position
}

// ImportContext maintains import context during template processing
type ImportContext struct {
	CurrentTemplate string
	ImportedFiles   map[string]bool
	VisitedMacros   map[string]bool

	// Thread safety
	mu sync.RWMutex
}

// NewImportContext creates a new import context
func NewImportContext(templateName string) *ImportContext {
	return &ImportContext{
		CurrentTemplate: templateName,
		ImportedFiles:   make(map[string]bool),
		VisitedMacros:   make(map[string]bool),
	}
}

// AddImport adds a file to the import context
func (ic *ImportContext) AddImport(templateName string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.ImportedFiles[templateName] = true
}

// HasImport checks if a file has been imported
func (ic *ImportContext) HasImport(templateName string) bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.ImportedFiles[templateName]
}

// AddMacro adds a macro to the visited set
func (ic *ImportContext) AddMacro(macroPath string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	ic.VisitedMacros[macroPath] = true
}

// HasMacro checks if a macro has been visited
func (ic *ImportContext) HasMacro(macroPath string) bool {
	ic.mu.RLock()
	defer ic.mu.RUnlock()
	return ic.VisitedMacros[macroPath]
}

// Clone creates a copy of the import context
func (ic *ImportContext) Clone() *ImportContext {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	clone := &ImportContext{
		CurrentTemplate: ic.CurrentTemplate,
		ImportedFiles:   make(map[string]bool),
		VisitedMacros:   make(map[string]bool),
	}

	for k, v := range ic.ImportedFiles {
		clone.ImportedFiles[k] = v
	}

	for k, v := range ic.VisitedMacros {
		clone.VisitedMacros[k] = v
	}

	return clone
}

// Merge merges another import context into this one
func (ic *ImportContext) Merge(other *ImportContext) {
	ic.mu.Lock()
	defer ic.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	for k, v := range other.ImportedFiles {
		ic.ImportedFiles[k] = v
	}

	for k, v := range other.VisitedMacros {
		ic.VisitedMacros[k] = v
	}
}
