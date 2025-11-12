package runtime

import (
	"fmt"
	"io"

	"github.com/deicod/gojinja/nodes"
	"github.com/deicod/gojinja/parser"
)

// Simple API functions for ease of use

// ParseString parses a template string and returns a ready-to-use Template
func ParseString(templateString string) (*Template, error) {
	return ParseStringWithName(templateString, "template")
}

// ParseStringWithName parses a template string with a given name
func ParseStringWithName(templateString, name string) (*Template, error) {
	env := NewEnvironment()
	return env.ParseString(templateString, name)
}

func ensureEnvironment(env *Environment) error {
	if env == nil {
		return NewError(ErrorTypeTemplate, "environment must not be nil", nodes.Position{}, nil)
	}
	return nil
}

// GetTemplate retrieves a template by name using the provided environment.
func GetTemplate(env *Environment, name string) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.GetTemplate(name)
}

// SelectTemplate resolves the first available template from the provided list
// of names, mirroring Jinja2's select_template helper.
func SelectTemplate(env *Environment, names []string) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.SelectTemplate(names)
}

// GetOrSelectTemplate resolves templates passed as names or template objects,
// behaving like Jinja2's get_or_select_template helper.
func GetOrSelectTemplate(env *Environment, target interface{}) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.GetOrSelectTemplate(target)
}

// JoinPath combines a template path with its parent template using the
// environment's loader join semantics.
func JoinPath(env *Environment, template, parent string) (string, error) {
	if err := ensureEnvironment(env); err != nil {
		return "", err
	}
	return env.JoinPath(template, parent)
}

// ParseString parses a template string using this environment
func (env *Environment) ParseString(templateString, name string) (*Template, error) {
	parserEnv := &parser.Environment{
		TrimBlocks:          env.trimBlocks,
		LstripBlocks:        env.lstripBlocks,
		KeepTrailingNewline: env.keepTrailingNewline,
		LineStatementPrefix: env.lineStatementPrefix,
		LineCommentPrefix:   env.lineCommentPrefix,
		Extensions:          env.Extensions(),
		EnableAsync:         env.IsAsyncEnabled(),
	}

	// Parse template using the parser
	ast, err := parser.ParseTemplateWithEnv(parserEnv, templateString, name, name)
	if err != nil {
		return nil, WrapError(err, nodes.Position{}, nil)
	}

	// Create template from AST
	return env.NewTemplateFromAST(ast, name)
}

// ExecuteToString is a convenience function that parses and renders a template string
func ExecuteToString(templateString string, vars map[string]interface{}) (string, error) {
	template, err := ParseString(templateString)
	if err != nil {
		return "", err
	}
	return template.ExecuteToString(vars)
}

// Execute is a convenience function that parses and renders a template string to a writer
func Execute(templateString string, vars map[string]interface{}, writer io.Writer) error {
	template, err := ParseString(templateString)
	if err != nil {
		return err
	}
	return template.Execute(vars, writer)
}

// ExecuteToStringWithEnvironment mirrors ExecuteToString but renders using the
// provided environment. This allows callers to honour environment-specific
// settings such as autoescape policies, newline trimming, async support, and
// registered extensions while still using the convenience helper.
func ExecuteToStringWithEnvironment(env *Environment, templateString string, vars map[string]interface{}) (string, error) {
	if err := ensureEnvironment(env); err != nil {
		return "", err
	}

	tmpl, err := env.ParseString(templateString, "template")
	if err != nil {
		return "", err
	}
	return tmpl.ExecuteToString(vars)
}

// ExecuteWithEnvironment renders a template string using the provided
// environment and writes the output to the supplied writer. The helper mirrors
// Execute while ensuring the environment's configuration (such as
// keep_trailing_newline or sandboxing) is respected.
func ExecuteWithEnvironment(env *Environment, templateString string, vars map[string]interface{}, writer io.Writer) error {
	if err := ensureEnvironment(env); err != nil {
		return err
	}
	if writer == nil {
		return NewError(ErrorTypeTemplate, "writer must not be nil", nodes.Position{}, nil)
	}

	tmpl, err := env.ParseString(templateString, "template")
	if err != nil {
		return err
	}
	return tmpl.Execute(vars, writer)
}

// Generate parses a template string and returns a streaming renderer that
// yields fragments as they are produced during execution.
func Generate(templateString string, vars map[string]interface{}) (*TemplateStream, error) {
	template, err := ParseString(templateString)
	if err != nil {
		return nil, err
	}
	return template.Generate(vars)
}

// GenerateWithEnvironment parses a template string using the provided
// environment and returns a streaming renderer.
func GenerateWithEnvironment(env *Environment, templateString string, vars map[string]interface{}) (*TemplateStream, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	tmpl, err := env.ParseString(templateString, "template")
	if err != nil {
		return nil, err
	}
	return tmpl.Generate(vars)
}

// GenerateToWriter renders a template string and writes the streamed output to
// the provided writer. It is equivalent to calling Generate followed by WriteTo
// and mirrors Jinja2's generator convenience helpers.
func GenerateToWriter(templateString string, vars map[string]interface{}, writer io.Writer) (int64, error) {
	if writer == nil {
		return 0, NewError(ErrorTypeTemplate, "writer must not be nil", nodes.Position{}, nil)
	}

	stream, err := Generate(templateString, vars)
	if err != nil {
		return 0, err
	}
	return stream.WriteTo(writer)
}

// GenerateToWriterWithEnvironment mirrors GenerateToWriter but uses the
// provided environment for parsing and execution.
func GenerateToWriterWithEnvironment(env *Environment, templateString string, vars map[string]interface{}, writer io.Writer) (int64, error) {
	if err := ensureEnvironment(env); err != nil {
		return 0, err
	}
	if writer == nil {
		return 0, NewError(ErrorTypeTemplate, "writer must not be nil", nodes.Position{}, nil)
	}

	tmpl, err := env.ParseString(templateString, "template")
	if err != nil {
		return 0, err
	}

	stream, err := tmpl.Generate(vars)
	if err != nil {
		return 0, err
	}

	return stream.WriteTo(writer)
}

// ParseAST creates a template from an existing AST
func ParseAST(ast *nodes.Template) (*Template, error) {
	return ParseASTWithName(ast, "template")
}

// ParseASTWithName creates a template from an AST with a given name
func ParseASTWithName(ast *nodes.Template, name string) (*Template, error) {
	env := NewEnvironment()
	return env.NewTemplateFromAST(ast, name)
}

// ParseASTWithEnvironment creates a template from an AST using the given environment
func ParseASTWithEnvironment(env *Environment, ast *nodes.Template, name string) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.NewTemplateFromAST(ast, name)
}

// FromString creates a new template from a string using the default environment
func FromString(templateString string) (*Template, error) {
	return ParseString(templateString)
}

// FromStringWithEnvironment creates a new template from a string using a specific environment
func FromStringWithEnvironment(env *Environment, templateString string) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.FromString(templateString)
}

// FromAST creates a new template from an AST using the default environment
func FromAST(ast *nodes.Template) (*Template, error) {
	return ParseAST(ast)
}

// FromASTWithEnvironment creates a new template from an AST using a specific environment
func FromASTWithEnvironment(env *Environment, ast *nodes.Template) (*Template, error) {
	if err := ensureEnvironment(env); err != nil {
		return nil, err
	}
	return env.NewTemplateFromAST(ast, "template")
}

// TemplateChain represents a chain of templates (useful for template inheritance)
type TemplateChain struct {
	templates   []*Template
	environment *Environment
}

// NewTemplateChain creates a new template chain
func NewTemplateChain(env *Environment) *TemplateChain {
	if env == nil {
		env = NewEnvironment()
	}

	return &TemplateChain{
		templates:   make([]*Template, 0),
		environment: env,
	}
}

// Add adds a template to the chain
func (tc *TemplateChain) Add(template *Template) {
	tc.templates = append(tc.templates, template)
}

// AddFromString adds a template from string to the chain
func (tc *TemplateChain) AddFromString(templateString, name string) error {
	if tc.environment == nil {
		tc.environment = NewEnvironment()
	}

	template, err := tc.environment.ParseString(templateString, name)
	if err != nil {
		return err
	}
	tc.Add(template)
	return nil
}

// AddFromAST adds a template from AST to the chain
func (tc *TemplateChain) AddFromAST(ast *nodes.Template, name string) error {
	if tc.environment == nil {
		tc.environment = NewEnvironment()
	}

	template, err := tc.environment.NewTemplateFromAST(ast, name)
	if err != nil {
		return err
	}
	tc.Add(template)
	return nil
}

// Get gets a template by name from the chain
func (tc *TemplateChain) Get(name string) (*Template, bool) {
	for _, template := range tc.templates {
		if template.Name() == name {
			return template, true
		}
	}
	return nil, false
}

// Has checks if a template exists in the chain
func (tc *TemplateChain) Has(name string) bool {
	_, ok := tc.Get(name)
	return ok
}

// Remove removes a template by name from the chain
func (tc *TemplateChain) Remove(name string) {
	for i, template := range tc.templates {
		if template.Name() == name {
			tc.templates = append(tc.templates[:i], tc.templates[i+1:]...)
			break
		}
	}
}

// Clear removes all templates from the chain
func (tc *TemplateChain) Clear() {
	tc.templates = make([]*Template, 0)
}

// Size returns the number of templates in the chain
func (tc *TemplateChain) Size() int {
	return len(tc.templates)
}

// All returns all templates in the chain
func (tc *TemplateChain) All() []*Template {
	templates := make([]*Template, len(tc.templates))
	copy(templates, tc.templates)
	return templates
}

// Names returns all template names in the chain
func (tc *TemplateChain) Names() []string {
	names := make([]string, len(tc.templates))
	for i, template := range tc.templates {
		names[i] = template.Name()
	}
	return names
}

// String returns a string representation of the template chain
func (tc *TemplateChain) String() string {
	var names []string
	for _, template := range tc.templates {
		names = append(names, template.Name())
	}
	return "TemplateChain([" + joinStrings(names, ", ") + "])"
}

// Helper functions

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// RenderTemplate renders a template with the given context
func RenderTemplate(templateString string, context map[string]interface{}) (string, error) {
	return ExecuteToString(templateString, context)
}

// RenderTemplateWithEnvironment renders a template with the given context and environment
func RenderTemplateWithEnvironment(env *Environment, templateString string, context map[string]interface{}) (string, error) {
	if err := ensureEnvironment(env); err != nil {
		return "", err
	}

	template, err := env.ParseString(templateString, "template")
	if err != nil {
		return "", err
	}
	return template.ExecuteToString(context)
}

// RenderTemplateToWriter renders a template to the given writer
func RenderTemplateToWriter(templateString string, context map[string]interface{}, writer io.Writer) error {
	return Execute(templateString, context, writer)
}

// RenderTemplateToWriterWithEnvironment renders a template to the given writer with environment
func RenderTemplateToWriterWithEnvironment(env *Environment, templateString string, context map[string]interface{}, writer io.Writer) error {
	if err := ensureEnvironment(env); err != nil {
		return err
	}

	template, err := env.ParseString(templateString, "template")
	if err != nil {
		return err
	}
	return template.Execute(context, writer)
}

// BatchRenderer renders multiple templates efficiently
type BatchRenderer struct {
	environment *Environment
	templates   map[string]*Template
}

// NewBatchRenderer creates a new batch renderer
func NewBatchRenderer(env *Environment) *BatchRenderer {
	if env == nil {
		env = NewEnvironment()
	}

	return &BatchRenderer{
		environment: env,
		templates:   make(map[string]*Template),
	}
}

// AddTemplate adds a template to the batch renderer
func (br *BatchRenderer) AddTemplate(name, templateString string) error {
	if br.environment == nil {
		br.environment = NewEnvironment()
	}

	template, err := br.environment.ParseString(templateString, name)
	if err != nil {
		return err
	}
	br.templates[name] = template
	return nil
}

// AddTemplateFromAST adds a template from AST to the batch renderer
func (br *BatchRenderer) AddTemplateFromAST(name string, ast *nodes.Template) error {
	if br.environment == nil {
		br.environment = NewEnvironment()
	}

	template, err := br.environment.NewTemplateFromAST(ast, name)
	if err != nil {
		return err
	}
	br.templates[name] = template
	return nil
}

// Render renders a template by name
func (br *BatchRenderer) Render(name string, context map[string]interface{}) (string, error) {
	template, ok := br.templates[name]
	if !ok {
		return "", NewTemplateNotFound(name, []string{name}, nil)
	}
	return template.ExecuteToString(context)
}

// RenderToWriter renders a template by name to a writer
func (br *BatchRenderer) RenderToWriter(name string, context map[string]interface{}, writer io.Writer) error {
	template, ok := br.templates[name]
	if !ok {
		return NewTemplateNotFound(name, []string{name}, nil)
	}
	return template.Execute(context, writer)
}

// HasTemplate checks if a template exists in the batch renderer
func (br *BatchRenderer) HasTemplate(name string) bool {
	_, ok := br.templates[name]
	return ok
}

// RemoveTemplate removes a template by name
func (br *BatchRenderer) RemoveTemplate(name string) {
	delete(br.templates, name)
}

// Clear removes all templates
func (br *BatchRenderer) Clear() {
	br.templates = make(map[string]*Template)
}

// Size returns the number of templates
func (br *BatchRenderer) Size() int {
	return len(br.templates)
}

// Names returns all template names
func (br *BatchRenderer) Names() []string {
	names := make([]string, 0, len(br.templates))
	for name := range br.templates {
		names = append(names, name)
	}
	return names
}

// GetAll returns all templates
func (br *BatchRenderer) GetAll() map[string]*Template {
	templates := make(map[string]*Template)
	for name, template := range br.templates {
		templates[name] = template
	}
	return templates
}

// String returns a string representation of the batch renderer
func (br *BatchRenderer) String() string {
	return fmt.Sprintf("BatchRenderer(templates=%d)", len(br.templates))
}
