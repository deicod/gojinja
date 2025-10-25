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

	// Check for circular imports
	for _, name := range im.importStack {
		if name == templateName {
			im.mu.Unlock()
			return nil, NewImportError(templateName, "circular import detected", nodes.Position{}, nil)
		}
	}

	// Check cache first
	if template, exists := im.importCache[templateName]; exists {
		im.mu.Unlock()
		return im.createNamespaceFromTemplate(ctx, templateName, template, withContext)
	}

	im.mu.Unlock()

	// Load the template
	template, err := im.environment.LoadTemplate(templateName)
	if err != nil {
		return nil, NewImportError(templateName, err.Error(), nodes.Position{}, nil)
	}

	im.mu.Lock()
	im.importCache[templateName] = template
	im.mu.Unlock()

	// Create namespace
	return im.createNamespaceFromTemplate(ctx, templateName, template, withContext)
}

// ImportMacros imports specific macros from a template
func (im *ImportManager) ImportMacros(ctx *Context, templateName string, macroNames []string, withContext bool) (map[string]interface{}, error) {
	namespace, err := im.ImportTemplate(ctx, templateName, withContext)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, macroName := range macroNames {
		name := macroName
		alias := macroName
		if strings.Contains(macroName, " as ") {
			parts := strings.SplitN(macroName, " as ", 2)
			if len(parts) == 2 {
				name = strings.TrimSpace(parts[0])
				alias = strings.TrimSpace(parts[1])
			}
		}

		value, ok := namespace.Resolve(name)
		if !ok {
			return nil, NewImportError(templateName,
				fmt.Sprintf("name '%s' not found in template", name),
				nodes.Position{}, nil)
		}

		result[alias] = value
	}

	return result, nil
}

// createNamespaceFromTemplate creates a macro namespace from a template
func (im *ImportManager) createNamespaceFromTemplate(ctx *Context, templateName string, template *Template, withContext bool) (*MacroNamespace, error) {
	var vars map[string]interface{}
	if withContext && ctx != nil {
		vars = ctx.scope.All()
	}

	im.mu.Lock()
	im.importStack = append(im.importStack, templateName)
	im.mu.Unlock()
	defer func() {
		im.mu.Lock()
		im.importStack = im.importStack[:len(im.importStack)-1]
		im.mu.Unlock()
	}()

	moduleCtx := template.newModuleContext(vars)
	moduleCtx.SetImportManager(im)

	module, err := template.makeModuleFromContext(moduleCtx)
	if err != nil {
		return nil, err
	}

	// Ensure the namespace is registered for later resolution
	im.registry.RegisterNamespace(templateName, module)

	return module, nil
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
	Template    string
	Macros      []string
	Alias       string
	WithContext bool
	Position    nodes.Position
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
