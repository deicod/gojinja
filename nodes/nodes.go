package nodes

import (
	"fmt"
	"strings"
)

// Position represents source code position information
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// NewPosition creates a new Position
func NewPosition(line, column int) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

// Node represents the base interface for all AST nodes
type Node interface {
	// GetPosition returns the position information for this node
	GetPosition() Position

	// SetPosition sets the position information for this node
	SetPosition(pos Position)

	// GetChildren returns all child nodes
	GetChildren() []Node

	// Accept implements the visitor pattern
	Accept(visitor Visitor) interface{}

	// String returns a string representation of the node
	String() string

	// Type returns the node type for identification
	Type() string
}

// BaseNode provides common functionality for all nodes
type BaseNode struct {
	Pos Position `json:"pos"`
}

// GetPosition returns the position information
func (n *BaseNode) GetPosition() Position {
	return n.Pos
}

// SetPosition sets the position information
func (n *BaseNode) SetPosition(pos Position) {
	n.Pos = pos
}

// GetChildren returns the base implementation (empty slice)
func (n *BaseNode) GetChildren() []Node {
	return []Node{}
}

// Type returns the node type name
func (n *BaseNode) Type() string {
	return "BaseNode"
}

// Visitor implements the visitor pattern for AST traversal
type Visitor interface {
	Visit(node Node) interface{}
}

// NodeVisitorFunc is a function adapter for Visitor interface
type NodeVisitorFunc func(node Node) interface{}

func (f NodeVisitorFunc) Visit(node Node) interface{} {
	return f(node)
}

// Walk traverses the AST using the visitor pattern
func Walk(visitor Visitor, node Node) {
	if node == nil {
		return
	}

	result := visitor.Visit(node)
	if result != nil {
		// If visitor returns non-nil, stop traversal
		return
	}

	// Visit children
	for _, child := range node.GetChildren() {
		Walk(visitor, child)
	}
}

// Node categories based on Python Jinja2 implementation

// Stmt represents statement nodes
type Stmt interface {
	Node
	isStmt()
}

func (n *BaseStmt) isStmt() {}

// BaseStmt provides common functionality for statement nodes
type BaseStmt struct {
	BaseNode
}

func (n *BaseStmt) Type() string {
	return "Stmt"
}

// Expr represents expression nodes
type Expr interface {
	Node
	isExpr()

	// AsConst attempts to evaluate the expression as a constant
	AsConst(ctx *EvalContext) (interface{}, error)

	// CanAssign checks if the expression can be assigned to
	CanAssign() bool
}

func (n *BaseExpr) isExpr() {}

// BaseExpr provides common functionality for expression nodes
type BaseExpr struct {
	BaseNode
}

func (n *BaseExpr) Type() string {
	return "Expr"
}

func (n *BaseExpr) AsConst(ctx *EvalContext) (interface{}, error) {
	return nil, fmt.Errorf("expression cannot be evaluated as constant")
}

func (n *BaseExpr) CanAssign() bool {
	return false
}

// Helper represents helper nodes that exist in specific contexts
type Helper interface {
	Node
	isHelper()
}

func (n *BaseHelper) isHelper() {}

// BaseHelper provides common functionality for helper nodes
type BaseHelper struct {
	BaseNode
}

func (n *BaseHelper) Type() string {
	return "Helper"
}

// Context types for expressions
const (
	CtxLoad  = "load"
	CtxStore = "store"
	CtxParam = "param"
)

// EvalContext holds evaluation time information
type EvalContext struct {
	Environment interface{} // Will be *environment.Environment when implemented
	AutoEscape  bool
	Volatile    bool
}

// NewEvalContext creates a new evaluation context
func NewEvalContext(env interface{}, templateName string) *EvalContext {
	// For now, basic implementation - will be enhanced when environment is implemented
	return &EvalContext{
		Environment: env,
		AutoEscape:  true, // Default value
		Volatile:    false,
	}
}

// Save returns a copy of the current context state
func (ctx *EvalContext) Save() map[string]interface{} {
	return map[string]interface{}{
		"autoescape": ctx.AutoEscape,
		"volatile":   ctx.Volatile,
	}
}

// Revert restores the context to a previous state
func (ctx *EvalContext) Revert(old map[string]interface{}) {
	if autoescape, ok := old["autoescape"].(bool); ok {
		ctx.AutoEscape = autoescape
	}
	if volatile, ok := old["volatile"].(bool); ok {
		ctx.Volatile = volatile
	}
}

// Template represents the outermost template node
type Template struct {
	BaseStmt
	Body []Node `json:"body"`
}

func (t *Template) Accept(visitor Visitor) interface{} {
	return visitor.Visit(t)
}

func (t *Template) GetChildren() []Node {
	return t.Body
}

func (t *Template) String() string {
	return fmt.Sprintf("Template(body=%v)", t.Body)
}

// Output holds multiple expressions to be printed
type Output struct {
	BaseStmt
	Nodes []Expr `json:"nodes"`
}

func (o *Output) Accept(visitor Visitor) interface{} {
	return visitor.Visit(o)
}

func (o *Output) GetChildren() []Node {
	children := make([]Node, len(o.Nodes))
	for i, node := range o.Nodes {
		children[i] = node
	}
	return children
}

func (o *Output) String() string {
	return fmt.Sprintf("Output(nodes=%v)", o.Nodes)
}

// Extends represents an extends statement
type Extends struct {
	BaseStmt
	Template Expr `json:"template"`
}

func (e *Extends) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *Extends) GetChildren() []Node {
	if e.Template != nil {
		return []Node{e.Template}
	}
	return []Node{}
}

func (e *Extends) String() string {
	return fmt.Sprintf("Extends(template=%v)", e.Template)
}

// For represents a for loop
type For struct {
	BaseStmt
	Target    Expr   `json:"target"`
	Iter      Expr   `json:"iter"`
	Body      []Node `json:"body"`
	Else      []Node `json:"else"`
	Test      Expr   `json:"test"`
	Recursive bool   `json:"recursive"`
}

func (f *For) Accept(visitor Visitor) interface{} {
	return visitor.Visit(f)
}

func (f *For) GetChildren() []Node {
	var children []Node

	if f.Target != nil {
		children = append(children, f.Target)
	}
	if f.Iter != nil {
		children = append(children, f.Iter)
	}
	if f.Test != nil {
		children = append(children, f.Test)
	}

	children = append(children, f.Body...)
	children = append(children, f.Else...)

	return children
}

func (f *For) String() string {
	return fmt.Sprintf("For(target=%v, iter=%v, body=%v, else=%v, test=%v, recursive=%t)",
		f.Target, f.Iter, f.Body, f.Else, f.Test, f.Recursive)
}

// If represents an if statement
type If struct {
	BaseStmt
	Test Expr   `json:"test"`
	Body []Node `json:"body"`
	Elif []*If  `json:"elif"`
	Else []Node `json:"else"`
}

func (i *If) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *If) GetChildren() []Node {
	var children []Node

	if i.Test != nil {
		children = append(children, i.Test)
	}

	children = append(children, i.Body...)

	for _, elif := range i.Elif {
		children = append(children, elif)
	}

	children = append(children, i.Else...)

	return children
}

func (i *If) String() string {
	return fmt.Sprintf("If(test=%v, body=%v, elif=%v, else=%v)",
		i.Test, i.Body, i.Elif, i.Else)
}

// Macro represents a macro definition
type Macro struct {
	BaseStmt
	Name     string  `json:"name"`
	Args     []*Name `json:"args"`
	Defaults []Expr  `json:"defaults"`
	VarArg   *Name   `json:"vararg"`
	KwArg    *Name   `json:"kwarg"`
	Body     []Node  `json:"body"`
}

func (m *Macro) Accept(visitor Visitor) interface{} {
	return visitor.Visit(m)
}

func (m *Macro) GetChildren() []Node {
	var children []Node

	for _, arg := range m.Args {
		children = append(children, arg)
	}

	if m.VarArg != nil {
		children = append(children, m.VarArg)
	}

	if m.KwArg != nil {
		children = append(children, m.KwArg)
	}

	for _, def := range m.Defaults {
		children = append(children, def)
	}

	children = append(children, m.Body...)

	return children
}

func (m *Macro) String() string {
	return fmt.Sprintf("Macro(name=%s, args=%v, vararg=%v, kwarg=%v, defaults=%v, body=%v)",
		m.Name, m.Args, m.VarArg, m.KwArg, m.Defaults, m.Body)
}

// CallBlock represents a call block (like a macro without a name)
type CallBlock struct {
	BaseStmt
	Call     *Call   `json:"call"`
	Args     []*Name `json:"args"`
	Defaults []Expr  `json:"defaults"`
	VarArg   *Name   `json:"vararg"`
	KwArg    *Name   `json:"kwarg"`
	Body     []Node  `json:"body"`
}

func (c *CallBlock) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *CallBlock) GetChildren() []Node {
	var children []Node

	if c.Call != nil {
		children = append(children, c.Call)
	}

	for _, arg := range c.Args {
		children = append(children, arg)
	}

	if c.VarArg != nil {
		children = append(children, c.VarArg)
	}

	if c.KwArg != nil {
		children = append(children, c.KwArg)
	}

	for _, def := range c.Defaults {
		children = append(children, def)
	}

	children = append(children, c.Body...)

	return children
}

func (c *CallBlock) String() string {
	return fmt.Sprintf("CallBlock(call=%v, args=%v, vararg=%v, kwarg=%v, defaults=%v, body=%v)",
		c.Call, c.Args, c.VarArg, c.KwArg, c.Defaults, c.Body)
}

// FilterBlock represents a filter section
type FilterBlock struct {
	BaseStmt
	Body   []Node  `json:"body"`
	Filter *Filter `json:"filter"`
}

func (f *FilterBlock) Accept(visitor Visitor) interface{} {
	return visitor.Visit(f)
}

func (f *FilterBlock) GetChildren() []Node {
	var children []Node

	children = append(children, f.Body...)

	if f.Filter != nil {
		children = append(children, f.Filter)
	}

	return children
}

func (f *FilterBlock) String() string {
	return fmt.Sprintf("FilterBlock(body=%v, filter=%v)", f.Body, f.Filter)
}

// Spaceless represents a spaceless block that collapses whitespace between HTML tags.
type Spaceless struct {
	BaseStmt
	Body []Node `json:"body"`
}

func (s *Spaceless) Accept(visitor Visitor) interface{} {
	return visitor.Visit(s)
}

func (s *Spaceless) GetChildren() []Node {
	return append([]Node(nil), s.Body...)
}

func (s *Spaceless) String() string {
	return fmt.Sprintf("Spaceless(body=%v)", s.Body)
}

// With represents a with statement
type With struct {
	BaseStmt
	Targets []Expr `json:"targets"`
	Values  []Expr `json:"values"`
	Body    []Node `json:"body"`
}

func (w *With) Accept(visitor Visitor) interface{} {
	return visitor.Visit(w)
}

func (w *With) GetChildren() []Node {
	var children []Node

	for _, target := range w.Targets {
		if target != nil {
			children = append(children, target)
		}
	}

	for _, value := range w.Values {
		if value != nil {
			children = append(children, value)
		}
	}

	children = append(children, w.Body...)

	return children
}

func (w *With) String() string {
	return fmt.Sprintf("With(targets=%v, values=%v, body=%v)",
		w.Targets, w.Values, w.Body)
}

// Namespace represents a namespace block that exposes a mutable namespace value
// to the surrounding scope after executing its body.
type Namespace struct {
	BaseStmt
	Name  string `json:"name"`
	Value Expr   `json:"value"`
	Body  []Node `json:"body"`
}

func (n *Namespace) Accept(visitor Visitor) interface{} {
	return visitor.Visit(n)
}

func (n *Namespace) GetChildren() []Node {
	var children []Node

	if n.Value != nil {
		children = append(children, n.Value)
	}

	children = append(children, n.Body...)
	return children
}

func (n *Namespace) String() string {
	return fmt.Sprintf("Namespace(name=%s, value=%v, body=%v)", n.Name, n.Value, n.Body)
}

// Export marks one or more names for export from the compiled template module.
type Export struct {
	BaseStmt
	Names []*Name `json:"names"`
}

func (e *Export) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *Export) GetChildren() []Node {
	children := make([]Node, len(e.Names))
	for i, name := range e.Names {
		children[i] = name
	}
	return children
}

func (e *Export) String() string {
	names := make([]string, len(e.Names))
	for i, name := range e.Names {
		names[i] = name.Name
	}
	return fmt.Sprintf("Export(%s)", strings.Join(names, ", "))
}

// Trans represents a translation block with optional pluralization support.
type Trans struct {
	BaseStmt
	Singular   []Node          `json:"singular"`
	Plural     []Node          `json:"plural"`
	Variables  map[string]Expr `json:"variables"`
	CountExpr  Expr            `json:"count_expr"`
	CountName  string          `json:"count_name"`
	Context    string          `json:"context"`
	HasContext bool            `json:"has_context"`
	Trimmed    bool            `json:"trimmed"`
	TrimmedSet bool            `json:"trimmed_set"`
}

func (t *Trans) Accept(visitor Visitor) interface{} {
	return visitor.Visit(t)
}

func (t *Trans) GetChildren() []Node {
	var children []Node

	if t.CountExpr != nil {
		children = append(children, t.CountExpr)
	}

	for _, expr := range t.Variables {
		if expr != nil {
			children = append(children, expr)
		}
	}

	children = append(children, t.Singular...)
	children = append(children, t.Plural...)

	return children
}

func (t *Trans) String() string {
	return fmt.Sprintf("Trans(count=%v, count_name=%s, context=%q, trimmed=%t, vars=%v, singular=%v, plural=%v)",
		t.CountExpr, t.CountName, t.Context, t.Trimmed, t.Variables, t.Singular, t.Plural)
}

// Block represents a block
type Block struct {
	BaseStmt
	Name     string `json:"name"`
	Body     []Node `json:"body"`
	Scoped   bool   `json:"scoped"`
	Required bool   `json:"required"`
}

func (b *Block) Accept(visitor Visitor) interface{} {
	return visitor.Visit(b)
}

func (b *Block) GetChildren() []Node {
	return b.Body
}

func (b *Block) String() string {
	return fmt.Sprintf("Block(name=%s, body=%v, scoped=%t, required=%t)",
		b.Name, b.Body, b.Scoped, b.Required)
}

// Include represents an include tag
type Include struct {
	BaseStmt
	Template      Expr `json:"template"`
	WithContext   bool `json:"with_context"`
	IgnoreMissing bool `json:"ignore_missing"`
}

func (i *Include) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *Include) GetChildren() []Node {
	if i.Template != nil {
		return []Node{i.Template}
	}
	return []Node{}
}

func (i *Include) String() string {
	return fmt.Sprintf("Include(template=%v, with_context=%t, ignore_missing=%t)",
		i.Template, i.WithContext, i.IgnoreMissing)
}

// Import represents an import tag
type Import struct {
	BaseStmt
	Template    Expr   `json:"template"`
	Target      string `json:"target"`
	WithContext bool   `json:"with_context"`
}

func (i *Import) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *Import) GetChildren() []Node {
	if i.Template != nil {
		return []Node{i.Template}
	}
	return []Node{}
}

func (i *Import) String() string {
	return fmt.Sprintf("Import(template=%v, target=%s, with_context=%t)",
		i.Template, i.Target, i.WithContext)
}

// FromImport represents a from import tag
type FromImport struct {
	BaseStmt
	Template    Expr         `json:"template"`
	Names       []ImportName `json:"names"`
	WithContext bool         `json:"with_context"`
}

// ImportName represents either a simple name or an alias (name, alias)
type ImportName struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
}

func (i *FromImport) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *FromImport) GetChildren() []Node {
	if i.Template != nil {
		return []Node{i.Template}
	}
	return []Node{}
}

func (i *FromImport) String() string {
	return fmt.Sprintf("FromImport(template=%v, names=%v, with_context=%t)",
		i.Template, i.Names, i.WithContext)
}

// ExprStmt represents a statement that evaluates an expression and discards the result
type ExprStmt struct {
	BaseStmt
	Node Expr `json:"node"`
}

func (e *ExprStmt) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *ExprStmt) GetChildren() []Node {
	if e.Node != nil {
		return []Node{e.Node}
	}
	return []Node{}
}

func (e *ExprStmt) String() string {
	return fmt.Sprintf("ExprStmt(node=%v)", e.Node)
}

// Assign represents an assignment statement
type Assign struct {
	BaseStmt
	Target Expr `json:"target"`
	Node   Expr `json:"node"`
}

func (a *Assign) Accept(visitor Visitor) interface{} {
	return visitor.Visit(a)
}

func (a *Assign) GetChildren() []Node {
	var children []Node
	if a.Target != nil {
		children = append(children, a.Target)
	}
	if a.Node != nil {
		children = append(children, a.Node)
	}
	return children
}

func (a *Assign) String() string {
	return fmt.Sprintf("Assign(target=%v, node=%v)", a.Target, a.Node)
}

// AssignBlock represents a block assignment
type AssignBlock struct {
	BaseStmt
	Target Expr    `json:"target"`
	Filter *Filter `json:"filter"`
	Body   []Node  `json:"body"`
}

func (a *AssignBlock) Accept(visitor Visitor) interface{} {
	return visitor.Visit(a)
}

func (a *AssignBlock) GetChildren() []Node {
	var children []Node

	if a.Target != nil {
		children = append(children, a.Target)
	}

	if a.Filter != nil {
		children = append(children, a.Filter)
	}

	children = append(children, a.Body...)

	return children
}

func (a *AssignBlock) String() string {
	return fmt.Sprintf("AssignBlock(target=%v, filter=%v, body=%v)",
		a.Target, a.Filter, a.Body)
}

// Do represents a do statement
type Do struct {
	BaseStmt
	Expr Expr `json:"expr"`
}

func (d *Do) Accept(visitor Visitor) interface{} {
	return visitor.Visit(d)
}

func (d *Do) GetChildren() []Node {
	if d.Expr != nil {
		return []Node{d.Expr}
	}
	return nil
}

func (d *Do) String() string {
	return fmt.Sprintf("Do(expr=%v)", d.Expr)
}

func (d *Do) Type() string {
	return "Do"
}

// BinExpr represents binary expressions
type BinExpr struct {
	BaseExpr
	Left     Expr   `json:"left"`
	Right    Expr   `json:"right"`
	Operator string `json:"operator"`
}

func (b *BinExpr) Accept(visitor Visitor) interface{} {
	return visitor.Visit(b)
}

func (b *BinExpr) GetChildren() []Node {
	var children []Node
	if b.Left != nil {
		children = append(children, b.Left)
	}
	if b.Right != nil {
		children = append(children, b.Right)
	}
	return children
}

func (b *BinExpr) String() string {
	return fmt.Sprintf("BinExpr(left=%v, right=%v, operator=%s)",
		b.Left, b.Right, b.Operator)
}

func (b *BinExpr) Type() string {
	return "BinExpr"
}

// UnaryExpr represents unary expressions
type UnaryExpr struct {
	BaseExpr
	Node     Expr   `json:"node"`
	Operator string `json:"operator"`
}

func (u *UnaryExpr) Accept(visitor Visitor) interface{} {
	return visitor.Visit(u)
}

func (u *UnaryExpr) GetChildren() []Node {
	if u.Node != nil {
		return []Node{u.Node}
	}
	return []Node{}
}

func (u *UnaryExpr) String() string {
	return fmt.Sprintf("UnaryExpr(node=%v, operator=%s)", u.Node, u.Operator)
}

func (u *UnaryExpr) Type() string {
	return "UnaryExpr"
}

// Name represents a name lookup
type Name struct {
	BaseExpr
	Name string `json:"name"`
	Ctx  string `json:"ctx"`
}

func (n *Name) Accept(visitor Visitor) interface{} {
	return visitor.Visit(n)
}

func (n *Name) GetChildren() []Node {
	return []Node{}
}

func (n *Name) String() string {
	return fmt.Sprintf("Name(name=%s, ctx=%s)", n.Name, n.Ctx)
}

func (n *Name) Type() string {
	return "Name"
}

func (n *Name) CanAssign() bool {
	// Names like true, false, none cannot be assigned to
	reservedNames := map[string]bool{
		"true": true, "false": true, "none": true,
		"True": true, "False": true, "None": true,
	}
	return !reservedNames[n.Name]
}

func (n *Name) AsConst(ctx *EvalContext) (interface{}, error) {
	// Only evaluate constants if we're in load context and it's a reserved name
	if n.Ctx == CtxLoad {
		switch n.Name {
		case "true", "True":
			return true, nil
		case "false", "False":
			return false, nil
		case "none", "None":
			return nil, nil
		}
	}
	return nil, fmt.Errorf("cannot evaluate name '%s' as constant", n.Name)
}

// NSRef represents a namespace reference
type NSRef struct {
	BaseExpr
	Name string `json:"name"`
	Attr string `json:"attr"`
}

func (n *NSRef) Accept(visitor Visitor) interface{} {
	return visitor.Visit(n)
}

func (n *NSRef) GetChildren() []Node {
	return []Node{}
}

func (n *NSRef) String() string {
	return fmt.Sprintf("NSRef(name=%s, attr=%s)", n.Name, n.Attr)
}

func (n *NSRef) Type() string {
	return "NSRef"
}

func (n *NSRef) CanAssign() bool {
	return true
}

// Literal represents literal values
type Literal interface {
	Expr
	isLiteral()
}

func (n *BaseLiteral) isLiteral() {}

// BaseLiteral provides common functionality for literal nodes
type BaseLiteral struct {
	BaseExpr
}

func (n *BaseLiteral) Type() string {
	return "Literal"
}

// Const represents a constant value
type Const struct {
	BaseLiteral
	Value interface{} `json:"value"`
}

func (c *Const) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *Const) GetChildren() []Node {
	return []Node{}
}

func (c *Const) String() string {
	return fmt.Sprintf("Const(value=%v)", c.Value)
}

func (c *Const) Type() string {
	return "Const"
}

func (c *Const) AsConst(ctx *EvalContext) (interface{}, error) {
	return c.Value, nil
}

// TemplateData represents constant template string
type TemplateData struct {
	BaseLiteral
	Data string `json:"data"`
}

func (t *TemplateData) Accept(visitor Visitor) interface{} {
	return visitor.Visit(t)
}

func (t *TemplateData) GetChildren() []Node {
	return []Node{}
}

func (t *TemplateData) String() string {
	return fmt.Sprintf("TemplateData(data=%s)", t.Data)
}

func (t *TemplateData) Type() string {
	return "TemplateData"
}

func (t *TemplateData) AsConst(ctx *EvalContext) (interface{}, error) {
	if ctx != nil && ctx.Volatile {
		return nil, fmt.Errorf("template data cannot be evaluated as constant in volatile context")
	}

	// In autoescape mode, this should be wrapped in Markup
	// For now, just return the string - will be enhanced with Markup implementation
	return t.Data, nil
}

// Tuple represents a tuple literal
type Tuple struct {
	BaseLiteral
	Items []Expr `json:"items"`
	Ctx   string `json:"ctx"`
}

func (t *Tuple) Accept(visitor Visitor) interface{} {
	return visitor.Visit(t)
}

func (t *Tuple) GetChildren() []Node {
	children := make([]Node, len(t.Items))
	for i, item := range t.Items {
		children[i] = item
	}
	return children
}

func (t *Tuple) String() string {
	return fmt.Sprintf("Tuple(items=%v, ctx=%s)", t.Items, t.Ctx)
}

func (t *Tuple) Type() string {
	return "Tuple"
}

func (t *Tuple) AsConst(ctx *EvalContext) (interface{}, error) {
	result := make([]interface{}, len(t.Items))
	for i, item := range t.Items {
		value, err := item.AsConst(ctx)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

func (t *Tuple) CanAssign() bool {
	for _, item := range t.Items {
		if !item.CanAssign() {
			return false
		}
	}
	return true
}

// List represents a list literal
type List struct {
	BaseLiteral
	Items []Expr `json:"items"`
}

func (l *List) Accept(visitor Visitor) interface{} {
	return visitor.Visit(l)
}

func (l *List) GetChildren() []Node {
	children := make([]Node, len(l.Items))
	for i, item := range l.Items {
		children[i] = item
	}
	return children
}

func (l *List) String() string {
	return fmt.Sprintf("List(items=%v)", l.Items)
}

func (l *List) Type() string {
	return "List"
}

func (l *List) AsConst(ctx *EvalContext) (interface{}, error) {
	result := make([]interface{}, len(l.Items))
	for i, item := range l.Items {
		value, err := item.AsConst(ctx)
		if err != nil {
			return nil, err
		}
		result[i] = value
	}
	return result, nil
}

// Dict represents a dictionary literal
type Dict struct {
	BaseLiteral
	Items []*Pair `json:"items"`
}

func (d *Dict) Accept(visitor Visitor) interface{} {
	return visitor.Visit(d)
}

func (d *Dict) GetChildren() []Node {
	children := make([]Node, len(d.Items))
	for i, item := range d.Items {
		children[i] = item
	}
	return children
}

func (d *Dict) String() string {
	return fmt.Sprintf("Dict(items=%v)", d.Items)
}

func (d *Dict) Type() string {
	return "Dict"
}

func (d *Dict) AsConst(ctx *EvalContext) (interface{}, error) {
	result := make(map[interface{}]interface{})
	for _, pair := range d.Items {
		keyValue, err := pair.Key.AsConst(ctx)
		if err != nil {
			return nil, err
		}
		value, err := pair.Value.AsConst(ctx)
		if err != nil {
			return nil, err
		}
		result[keyValue] = value
	}
	return result, nil
}

// Pair represents a key-value pair for dictionaries
type Pair struct {
	BaseHelper
	Key   Expr `json:"key"`
	Value Expr `json:"value"`
}

func (p *Pair) Accept(visitor Visitor) interface{} {
	return visitor.Visit(p)
}

func (p *Pair) GetChildren() []Node {
	var children []Node
	if p.Key != nil {
		children = append(children, p.Key)
	}
	if p.Value != nil {
		children = append(children, p.Value)
	}
	return children
}

func (p *Pair) String() string {
	return fmt.Sprintf("Pair(key=%v, value=%v)", p.Key, p.Value)
}

func (p *Pair) Type() string {
	return "Pair"
}

func (p *Pair) AsConst(ctx *EvalContext) (interface{}, error) {
	keyValue, err := p.Key.AsConst(ctx)
	if err != nil {
		return nil, err
	}
	value, err := p.Value.AsConst(ctx)
	if err != nil {
		return nil, err
	}
	return []interface{}{keyValue, value}, nil
}

// Keyword represents a keyword argument
type Keyword struct {
	BaseHelper
	Key   string `json:"key"`
	Value Expr   `json:"value"`
}

func (k *Keyword) Accept(visitor Visitor) interface{} {
	return visitor.Visit(k)
}

func (k *Keyword) GetChildren() []Node {
	if k.Value != nil {
		return []Node{k.Value}
	}
	return []Node{}
}

func (k *Keyword) String() string {
	return fmt.Sprintf("Keyword(key=%s, value=%v)", k.Key, k.Value)
}

func (k *Keyword) Type() string {
	return "Keyword"
}

func (k *Keyword) AsConst(ctx *EvalContext) (interface{}, error) {
	value, err := k.Value.AsConst(ctx)
	if err != nil {
		return nil, err
	}
	return []interface{}{k.Key, value}, nil
}

// CondExpr represents a conditional expression (inline if)
type CondExpr struct {
	BaseExpr
	Test  Expr `json:"test"`
	Expr1 Expr `json:"expr1"`
	Expr2 Expr `json:"expr2"`
}

func (c *CondExpr) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *CondExpr) GetChildren() []Node {
	var children []Node
	if c.Test != nil {
		children = append(children, c.Test)
	}
	if c.Expr1 != nil {
		children = append(children, c.Expr1)
	}
	if c.Expr2 != nil {
		children = append(children, c.Expr2)
	}
	return children
}

func (c *CondExpr) String() string {
	return fmt.Sprintf("CondExpr(test=%v, expr1=%v, expr2=%v)",
		c.Test, c.Expr1, c.Expr2)
}

func (c *CondExpr) Type() string {
	return "CondExpr"
}

func (c *CondExpr) AsConst(ctx *EvalContext) (interface{}, error) {
	testValue, err := c.Test.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	if testValue != nil && testValue != false && testValue != 0 && testValue != "" {
		return c.Expr1.AsConst(ctx)
	}

	if c.Expr2 != nil {
		return c.Expr2.AsConst(ctx)
	}

	return nil, fmt.Errorf("conditional expression has no else clause")
}

// FilterTestCommon represents common functionality for filters and tests
type FilterTestCommon struct {
	BaseExpr
	Node      Expr    `json:"node"`
	Name      string  `json:"name"`
	Args      []Expr  `json:"args"`
	Kwargs    []*Pair `json:"kwargs"`
	DynArgs   Expr    `json:"dyn_args"`
	DynKwargs Expr    `json:"dyn_kwargs"`
	IsFilter  bool    `json:"is_filter"`
}

func (f *FilterTestCommon) Accept(visitor Visitor) interface{} {
	return visitor.Visit(f)
}

func (f *FilterTestCommon) GetChildren() []Node {
	var children []Node

	if f.Node != nil {
		children = append(children, f.Node)
	}

	for _, arg := range f.Args {
		children = append(children, arg)
	}

	for _, kwarg := range f.Kwargs {
		children = append(children, kwarg)
	}

	if f.DynArgs != nil {
		children = append(children, f.DynArgs)
	}

	if f.DynKwargs != nil {
		children = append(children, f.DynKwargs)
	}

	return children
}

func (f *FilterTestCommon) String() string {
	nodeType := "Test"
	if f.IsFilter {
		nodeType = "Filter"
	}
	return fmt.Sprintf("%s(node=%v, name=%s, args=%v, kwargs=%v, dyn_args=%v, dyn_kwargs=%v)",
		nodeType, f.Node, f.Name, f.Args, f.Kwargs, f.DynArgs, f.DynKwargs)
}

func (f *FilterTestCommon) Type() string {
	if f.IsFilter {
		return "Filter"
	}
	return "Test"
}

// Filter represents a filter application
type Filter struct {
	FilterTestCommon
}

func (f *Filter) AsConst(ctx *EvalContext) (interface{}, error) {
	if f.Node == nil {
		return nil, fmt.Errorf("filter cannot be evaluated as constant without node")
	}

	// Basic implementation - will be enhanced with actual filter lookup
	// For now, just return error to indicate this cannot be evaluated at compile time
	return nil, fmt.Errorf("filter evaluation requires runtime environment")
}

// Test represents a test application
type Test struct {
	FilterTestCommon
}

func (t *Test) AsConst(ctx *EvalContext) (interface{}, error) {
	if t.Node == nil {
		return nil, fmt.Errorf("test cannot be evaluated as constant without node")
	}

	// Basic implementation - will be enhanced with actual test lookup
	// For now, just return error to indicate this cannot be evaluated at compile time
	return nil, fmt.Errorf("test evaluation requires runtime environment")
}

// Call represents a function call
type Call struct {
	BaseExpr
	Node      Expr       `json:"node"`
	Args      []Expr     `json:"args"`
	Kwargs    []*Keyword `json:"kwargs"`
	DynArgs   Expr       `json:"dyn_args"`
	DynKwargs Expr       `json:"dyn_kwargs"`
}

func (c *Call) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *Call) GetChildren() []Node {
	var children []Node

	if c.Node != nil {
		children = append(children, c.Node)
	}

	for _, arg := range c.Args {
		children = append(children, arg)
	}

	for _, kwarg := range c.Kwargs {
		children = append(children, kwarg)
	}

	if c.DynArgs != nil {
		children = append(children, c.DynArgs)
	}

	if c.DynKwargs != nil {
		children = append(children, c.DynKwargs)
	}

	return children
}

func (c *Call) String() string {
	return fmt.Sprintf("Call(node=%v, args=%v, kwargs=%v, dyn_args=%v, dyn_kwargs=%v)",
		c.Node, c.Args, c.Kwargs, c.DynArgs, c.DynKwargs)
}

func (c *Call) Type() string {
	return "Call"
}

// Getitem represents item access (obj[key])
type Getitem struct {
	BaseExpr
	Node Expr   `json:"node"`
	Arg  Expr   `json:"arg"`
	Ctx  string `json:"ctx"`
}

func (g *Getitem) Accept(visitor Visitor) interface{} {
	return visitor.Visit(g)
}

func (g *Getitem) GetChildren() []Node {
	var children []Node
	if g.Node != nil {
		children = append(children, g.Node)
	}
	if g.Arg != nil {
		children = append(children, g.Arg)
	}
	return children
}

func (g *Getitem) String() string {
	return fmt.Sprintf("Getitem(node=%v, arg=%v, ctx=%s)", g.Node, g.Arg, g.Ctx)
}

func (g *Getitem) Type() string {
	return "Getitem"
}

func (g *Getitem) CanAssign() bool {
	return g.Ctx != CtxLoad
}

func (g *Getitem) AsConst(ctx *EvalContext) (interface{}, error) {
	if g.Ctx != CtxLoad {
		return nil, fmt.Errorf("getitem cannot be evaluated as constant in non-load context")
	}

	// Basic implementation - will be enhanced with actual environment.getitem
	return nil, fmt.Errorf("getitem evaluation requires runtime environment")
}

// Getattr represents attribute access (obj.attr)
type Getattr struct {
	BaseExpr
	Node Expr   `json:"node"`
	Attr string `json:"attr"`
	Ctx  string `json:"ctx"`
}

func (g *Getattr) Accept(visitor Visitor) interface{} {
	return visitor.Visit(g)
}

func (g *Getattr) GetChildren() []Node {
	if g.Node != nil {
		return []Node{g.Node}
	}
	return []Node{}
}

func (g *Getattr) String() string {
	return fmt.Sprintf("Getattr(node=%v, attr=%s, ctx=%s)", g.Node, g.Attr, g.Ctx)
}

func (g *Getattr) Type() string {
	return "Getattr"
}

func (g *Getattr) CanAssign() bool {
	return g.Ctx != CtxLoad
}

func (g *Getattr) AsConst(ctx *EvalContext) (interface{}, error) {
	if g.Ctx != CtxLoad {
		return nil, fmt.Errorf("getattr cannot be evaluated as constant in non-load context")
	}

	// Basic implementation - will be enhanced with actual environment.getattr
	return nil, fmt.Errorf("getattr evaluation requires runtime environment")
}

// Slice represents a slice object
type Slice struct {
	BaseExpr
	Start Expr `json:"start"`
	Stop  Expr `json:"stop"`
	Step  Expr `json:"step"`
}

func (s *Slice) Accept(visitor Visitor) interface{} {
	return visitor.Visit(s)
}

func (s *Slice) GetChildren() []Node {
	var children []Node
	if s.Start != nil {
		children = append(children, s.Start)
	}
	if s.Stop != nil {
		children = append(children, s.Stop)
	}
	if s.Step != nil {
		children = append(children, s.Step)
	}
	return children
}

func (s *Slice) String() string {
	return fmt.Sprintf("Slice(start=%v, stop=%v, step=%v)", s.Start, s.Stop, s.Step)
}

func (s *Slice) Type() string {
	return "Slice"
}

func (s *Slice) AsConst(ctx *EvalContext) (interface{}, error) {
	var start, stop, step interface{}
	var err error

	if s.Start != nil {
		start, err = s.Start.AsConst(ctx)
		if err != nil {
			return nil, err
		}
	}

	if s.Stop != nil {
		stop, err = s.Stop.AsConst(ctx)
		if err != nil {
			return nil, err
		}
	}

	if s.Step != nil {
		step, err = s.Step.AsConst(ctx)
		if err != nil {
			return nil, err
		}
	}

	// Convert to Go slice indices
	var startInt, stopInt, stepInt int
	if start != nil {
		if val, ok := start.(int); ok {
			startInt = val
		} else {
			return nil, fmt.Errorf("slice start must be integer")
		}
	}

	if stop != nil {
		if val, ok := stop.(int); ok {
			stopInt = val
		} else {
			return nil, fmt.Errorf("slice stop must be integer")
		}
	}

	if step != nil {
		if val, ok := step.(int); ok {
			stepInt = val
		} else {
			return nil, fmt.Errorf("slice step must be integer")
		}
	}

	// Return a representation that can be used for slicing
	// In a full implementation, this would create a proper slice object
	return map[string]interface{}{
		"start": startInt,
		"stop":  stopInt,
		"step":  stepInt,
	}, nil
}

// Concat represents string concatenation
type Concat struct {
	BaseExpr
	Nodes []Expr `json:"nodes"`
}

func (c *Concat) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *Concat) GetChildren() []Node {
	children := make([]Node, len(c.Nodes))
	for i, node := range c.Nodes {
		children[i] = node
	}
	return children
}

func (c *Concat) String() string {
	return fmt.Sprintf("Concat(nodes=%v)", c.Nodes)
}

func (c *Concat) Type() string {
	return "Concat"
}

func (c *Concat) AsConst(ctx *EvalContext) (interface{}, error) {
	var result strings.Builder
	for _, node := range c.Nodes {
		value, err := node.AsConst(ctx)
		if err != nil {
			return nil, err
		}
		result.WriteString(fmt.Sprintf("%v", value))
	}
	return result.String(), nil
}

// Compare represents comparison expressions
type Compare struct {
	BaseExpr
	Expr Expr       `json:"expr"`
	Ops  []*Operand `json:"ops"`
}

func (c *Compare) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *Compare) GetChildren() []Node {
	var children []Node

	if c.Expr != nil {
		children = append(children, c.Expr)
	}

	for _, op := range c.Ops {
		children = append(children, op)
	}

	return children
}

func (c *Compare) String() string {
	return fmt.Sprintf("Compare(expr=%v, ops=%v)", c.Expr, c.Ops)
}

func (c *Compare) Type() string {
	return "Compare"
}

func (c *Compare) AsConst(ctx *EvalContext) (interface{}, error) {
	if c.Expr == nil || len(c.Ops) == 0 {
		return nil, fmt.Errorf("compare expression incomplete")
	}

	// Evaluate left side
	leftValue, err := c.Expr.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	// Evaluate each comparison
	currentValue := leftValue
	for _, op := range c.Ops {
		if op.Expr == nil {
			return nil, fmt.Errorf("comparison operand missing expression")
		}

		rightValue, err := op.Expr.AsConst(ctx)
		if err != nil {
			return nil, err
		}

		// Perform comparison based on operator
		result, err := performComparison(op.Op, currentValue, rightValue)
		if err != nil {
			return nil, err
		}

		if !result {
			return false, nil
		}

		currentValue = rightValue
	}

	return true, nil
}

// performComparison performs a comparison operation
func performComparison(op string, left, right interface{}) (bool, error) {
	switch op {
	case "eq":
		return left == right, nil
	case "ne":
		return left != right, nil
	case "gt":
		return compareValues(left, right) > 0, nil
	case "gteq":
		return compareValues(left, right) >= 0, nil
	case "lt":
		return compareValues(left, right) < 0, nil
	case "lteq":
		return compareValues(left, right) <= 0, nil
	case "in":
		return isInCollection(left, right), nil
	case "notin":
		return !isInCollection(left, right), nil
	default:
		return false, fmt.Errorf("unsupported comparison operator: %s", op)
	}
}

// compareValues compares two values for ordering
func compareValues(left, right interface{}) int {
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

	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	}
	return 0
}

// toFloat64 attempts to convert a value to float64 for comparison
func toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		return 0, false
	}
}

// isInCollection checks if a value is in a collection
func isInCollection(item, collection interface{}) bool {
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

// Operand represents a comparison operator and expression
type Operand struct {
	BaseHelper
	Op   string `json:"op"`
	Expr Expr   `json:"expr"`
}

func (o *Operand) Accept(visitor Visitor) interface{} {
	return visitor.Visit(o)
}

func (o *Operand) GetChildren() []Node {
	if o.Expr != nil {
		return []Node{o.Expr}
	}
	return []Node{}
}

func (o *Operand) String() string {
	return fmt.Sprintf("Operand(op=%s, expr=%v)", o.Op, o.Expr)
}

func (o *Operand) Type() string {
	return "Operand"
}

// Binary operation specific nodes

// Mul represents multiplication
type Mul struct {
	BinExpr
}

func (m *Mul) Type() string {
	return "Mul"
}

func NewMul(left, right Expr) *Mul {
	return &Mul{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "*",
		},
	}
}

// Div represents division
type Div struct {
	BinExpr
}

func (d *Div) Type() string {
	return "Div"
}

func NewDiv(left, right Expr) *Div {
	return &Div{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "/",
		},
	}
}

// FloorDiv represents floor division
type FloorDiv struct {
	BinExpr
}

func (f *FloorDiv) Type() string {
	return "FloorDiv"
}

func NewFloorDiv(left, right Expr) *FloorDiv {
	return &FloorDiv{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "//",
		},
	}
}

// Add represents addition
type Add struct {
	BinExpr
}

func (a *Add) Type() string {
	return "Add"
}

func NewAdd(left, right Expr) *Add {
	return &Add{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "+",
		},
	}
}

// Sub represents subtraction
type Sub struct {
	BinExpr
}

func (s *Sub) Type() string {
	return "Sub"
}

func NewSub(left, right Expr) *Sub {
	return &Sub{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "-",
		},
	}
}

// Mod represents modulo
type Mod struct {
	BinExpr
}

func (m *Mod) Type() string {
	return "Mod"
}

func NewMod(left, right Expr) *Mod {
	return &Mod{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "%",
		},
	}
}

// Pow represents exponentiation
type Pow struct {
	BinExpr
}

func (p *Pow) Type() string {
	return "Pow"
}

func NewPow(left, right Expr) *Pow {
	return &Pow{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "**",
		},
	}
}

// And represents logical AND
type And struct {
	BinExpr
}

func (a *And) Type() string {
	return "And"
}

func NewAnd(left, right Expr) *And {
	return &And{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "and",
		},
	}
}

func (a *And) AsConst(ctx *EvalContext) (interface{}, error) {
	leftValue, err := a.Left.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	if !isTruthy(leftValue) {
		return false, nil
	}

	return a.Right.AsConst(ctx)
}

// Or represents logical OR
type Or struct {
	BinExpr
}

func (o *Or) Type() string {
	return "Or"
}

func NewOr(left, right Expr) *Or {
	return &Or{
		BinExpr: BinExpr{
			Left:     left,
			Right:    right,
			Operator: "or",
		},
	}
}

func (o *Or) AsConst(ctx *EvalContext) (interface{}, error) {
	leftValue, err := o.Left.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	if isTruthy(leftValue) {
		return true, nil
	}

	return o.Right.AsConst(ctx)
}

// Unary operation specific nodes

// Not represents logical NOT
type Not struct {
	UnaryExpr
}

func (n *Not) Type() string {
	return "Not"
}

func NewNot(node Expr) *Not {
	return &Not{
		UnaryExpr: UnaryExpr{
			Node:     node,
			Operator: "not",
		},
	}
}

func (n *Not) AsConst(ctx *EvalContext) (interface{}, error) {
	value, err := n.Node.AsConst(ctx)
	if err != nil {
		return nil, err
	}
	return !isTruthy(value), nil
}

// Neg represents arithmetic negation
type Neg struct {
	UnaryExpr
}

func (n *Neg) Type() string {
	return "Neg"
}

func NewNeg(node Expr) *Neg {
	return &Neg{
		UnaryExpr: UnaryExpr{
			Node:     node,
			Operator: "-",
		},
	}
}

func (n *Neg) AsConst(ctx *EvalContext) (interface{}, error) {
	value, err := n.Node.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case int:
		return -v, nil
	case int64:
		return -v, nil
	case float64:
		return -v, nil
	case float32:
		return -v, nil
	default:
		return nil, fmt.Errorf("cannot negate value of type %T", value)
	}
}

// Pos represents arithmetic positive (noop)
type Pos struct {
	UnaryExpr
}

func (p *Pos) Type() string {
	return "Pos"
}

func NewPos(node Expr) *Pos {
	return &Pos{
		UnaryExpr: UnaryExpr{
			Node:     node,
			Operator: "+",
		},
	}
}

func (p *Pos) AsConst(ctx *EvalContext) (interface{}, error) {
	return p.Node.AsConst(ctx)
}

// Helper function to determine truthiness
func isTruthy(value interface{}) bool {
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

// Extension helper nodes

// EnvironmentAttribute loads an attribute from the environment
type EnvironmentAttribute struct {
	BaseExpr
	Name string `json:"name"`
}

func (e *EnvironmentAttribute) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *EnvironmentAttribute) GetChildren() []Node {
	return []Node{}
}

func (e *EnvironmentAttribute) String() string {
	return fmt.Sprintf("EnvironmentAttribute(name=%s)", e.Name)
}

func (e *EnvironmentAttribute) Type() string {
	return "EnvironmentAttribute"
}

// ExtensionAttribute returns an attribute from an extension
type ExtensionAttribute struct {
	BaseExpr
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

func (e *ExtensionAttribute) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *ExtensionAttribute) GetChildren() []Node {
	return []Node{}
}

func (e *ExtensionAttribute) String() string {
	return fmt.Sprintf("ExtensionAttribute(identifier=%s, name=%s)", e.Identifier, e.Name)
}

func (e *ExtensionAttribute) Type() string {
	return "ExtensionAttribute"
}

// ImportedName represents an imported name
type ImportedName struct {
	BaseExpr
	ImportName string `json:"importname"`
}

func (i *ImportedName) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *ImportedName) GetChildren() []Node {
	return []Node{}
}

func (i *ImportedName) String() string {
	return fmt.Sprintf("ImportedName(importname=%s)", i.ImportName)
}

func (i *ImportedName) Type() string {
	return "ImportedName"
}

// InternalName represents an internal name in the compiler
type InternalName struct {
	BaseExpr
	Name string `json:"name"`
}

func (i *InternalName) Accept(visitor Visitor) interface{} {
	return visitor.Visit(i)
}

func (i *InternalName) GetChildren() []Node {
	return []Node{}
}

func (i *InternalName) String() string {
	return fmt.Sprintf("InternalName(name=%s)", i.Name)
}

func (i *InternalName) Type() string {
	return "InternalName"
}

// MarkSafe marks an expression as safe (wraps as Markup)
type MarkSafe struct {
	BaseExpr
	Expr Expr `json:"expr"`
}

func (m *MarkSafe) Accept(visitor Visitor) interface{} {
	return visitor.Visit(m)
}

func (m *MarkSafe) GetChildren() []Node {
	if m.Expr != nil {
		return []Node{m.Expr}
	}
	return []Node{}
}

func (m *MarkSafe) String() string {
	return fmt.Sprintf("MarkSafe(expr=%v)", m.Expr)
}

func (m *MarkSafe) Type() string {
	return "MarkSafe"
}

// MarkSafeIfAutoescape marks expression as safe only if autoescaping is active
type MarkSafeIfAutoescape struct {
	BaseExpr
	Expr Expr `json:"expr"`
}

func (m *MarkSafeIfAutoescape) Accept(visitor Visitor) interface{} {
	return visitor.Visit(m)
}

func (m *MarkSafeIfAutoescape) GetChildren() []Node {
	if m.Expr != nil {
		return []Node{m.Expr}
	}
	return []Node{}
}

func (m *MarkSafeIfAutoescape) String() string {
	return fmt.Sprintf("MarkSafeIfAutoescape(expr=%v)", m.Expr)
}

func (m *MarkSafeIfAutoescape) Type() string {
	return "MarkSafeIfAutoescape"
}

func (m *MarkSafeIfAutoescape) AsConst(ctx *EvalContext) (interface{}, error) {
	if ctx != nil && ctx.Volatile {
		return nil, fmt.Errorf("cannot evaluate in volatile context")
	}

	value, err := m.Expr.AsConst(ctx)
	if err != nil {
		return nil, err
	}

	if ctx != nil && ctx.AutoEscape {
		// In a full implementation, this would wrap in Markup
		// For now, just return the value
		return value, nil
	}

	return value, nil
}

// ContextReference returns the current template context
type ContextReference struct {
	BaseExpr
}

func (c *ContextReference) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *ContextReference) GetChildren() []Node {
	return []Node{}
}

func (c *ContextReference) String() string {
	return "ContextReference()"
}

func (c *ContextReference) Type() string {
	return "ContextReference"
}

// DerivedContextReference returns the current context including locals
type DerivedContextReference struct {
	BaseExpr
}

func (d *DerivedContextReference) Accept(visitor Visitor) interface{} {
	return visitor.Visit(d)
}

func (d *DerivedContextReference) GetChildren() []Node {
	return []Node{}
}

func (d *DerivedContextReference) String() string {
	return "DerivedContextReference()"
}

func (d *DerivedContextReference) Type() string {
	return "DerivedContextReference"
}

// Continue represents a continue statement
type Continue struct {
	BaseStmt
}

func (c *Continue) Accept(visitor Visitor) interface{} {
	return visitor.Visit(c)
}

func (c *Continue) GetChildren() []Node {
	return []Node{}
}

func (c *Continue) String() string {
	return "Continue()"
}

func (c *Continue) Type() string {
	return "Continue"
}

// Break represents a break statement
type Break struct {
	BaseStmt
}

func (b *Break) Accept(visitor Visitor) interface{} {
	return visitor.Visit(b)
}

func (b *Break) GetChildren() []Node {
	return []Node{}
}

func (b *Break) String() string {
	return "Break()"
}

func (b *Break) Type() string {
	return "Break"
}

// Scope represents an artificial scope
type Scope struct {
	BaseStmt
	Body []Node `json:"body"`
}

func (s *Scope) Accept(visitor Visitor) interface{} {
	return visitor.Visit(s)
}

func (s *Scope) GetChildren() []Node {
	return s.Body
}

func (s *Scope) String() string {
	return fmt.Sprintf("Scope(body=%v)", s.Body)
}

func (s *Scope) Type() string {
	return "Scope"
}

// OverlayScope represents an overlay scope for extensions
type OverlayScope struct {
	BaseStmt
	Context Expr   `json:"context"`
	Body    []Node `json:"body"`
}

func (o *OverlayScope) Accept(visitor Visitor) interface{} {
	return visitor.Visit(o)
}

func (o *OverlayScope) GetChildren() []Node {
	var children []Node

	if o.Context != nil {
		children = append(children, o.Context)
	}

	children = append(children, o.Body...)

	return children
}

func (o *OverlayScope) String() string {
	return fmt.Sprintf("OverlayScope(context=%v, body=%v)", o.Context, o.Body)
}

func (o *OverlayScope) Type() string {
	return "OverlayScope"
}

// EvalContextModifier modifies the eval context
type EvalContextModifier struct {
	BaseStmt
	Options []*Keyword `json:"options"`
}

func (e *EvalContextModifier) Accept(visitor Visitor) interface{} {
	return visitor.Visit(e)
}

func (e *EvalContextModifier) GetChildren() []Node {
	children := make([]Node, len(e.Options))
	for i, option := range e.Options {
		children[i] = option
	}
	return children
}

func (e *EvalContextModifier) String() string {
	return fmt.Sprintf("EvalContextModifier(options=%v)", e.Options)
}

func (e *EvalContextModifier) Type() string {
	return "EvalContextModifier"
}

// ScopedEvalContextModifier modifies eval context and reverts it later
type ScopedEvalContextModifier struct {
	EvalContextModifier
	Body []Node `json:"body"`
}

func (s *ScopedEvalContextModifier) Accept(visitor Visitor) interface{} {
	return visitor.Visit(s)
}

func (s *ScopedEvalContextModifier) GetChildren() []Node {
	var children []Node

	for _, option := range s.Options {
		children = append(children, option)
	}

	children = append(children, s.Body...)

	return children
}

func (s *ScopedEvalContextModifier) String() string {
	return fmt.Sprintf("ScopedEvalContextModifier(options=%v, body=%v)", s.Options, s.Body)
}

func (s *ScopedEvalContextModifier) Type() string {
	return "ScopedEvalContextModifier"
}

// Node utility functions

// SetCtx sets the context for a node and all its children
func SetCtx(node Node, ctx string) Node {
	visitor := NodeVisitorFunc(func(n Node) interface{} {
		if nameNode, ok := n.(*Name); ok {
			nameNode.Ctx = ctx
		} else if tupleNode, ok := n.(*Tuple); ok {
			tupleNode.Ctx = ctx
		} else if getattrNode, ok := n.(*Getattr); ok {
			getattrNode.Ctx = ctx
		} else if getitemNode, ok := n.(*Getitem); ok {
			getitemNode.Ctx = ctx
		}
		return nil
	})

	Walk(visitor, node)
	return node
}

// SetLineNo sets the line number for a node and all its children
func SetLineNo(node Node, line int, override bool) Node {
	visitor := NodeVisitorFunc(func(n Node) interface{} {
		if override || n.GetPosition().Line == 0 {
			currentPos := n.GetPosition()
			n.SetPosition(Position{Line: line, Column: currentPos.Column})
		}
		return nil
	})

	Walk(visitor, node)
	return node
}

// Find finds the first node of the given type
func Find(node Node, nodeType interface{}) Node {
	var result Node
	visitor := NodeVisitorFunc(func(n Node) interface{} {
		if result != nil {
			return true // Stop traversal
		}

		// Check if node matches the requested type
		switch target := nodeType.(type) {
		case string:
			if n.Type() == target {
				result = n
				return true
			}
		case func(Node) bool:
			if target(n) {
				result = n
				return true
			}
		default:
			// Type assertion for specific node types
			if n == target {
				result = n
				return true
			}
		}

		return nil
	})

	Walk(visitor, node)
	return result
}

// FindAll finds all nodes of the given type
func FindAll(node Node, nodeType interface{}) []Node {
	var results []Node
	visitor := NodeVisitorFunc(func(n Node) interface{} {
		// Check if node matches the requested type
		switch target := nodeType.(type) {
		case string:
			if n.Type() == target {
				results = append(results, n)
			}
		case func(Node) bool:
			if target(n) {
				results = append(results, n)
			}
		default:
			// Type assertion for specific node types
			if n == target {
				results = append(results, n)
			}
		}

		return nil
	})

	Walk(visitor, node)
	return results
}

// Dump creates a string representation of the AST for debugging
func Dump(node Node) string {
	if node == nil {
		return "nil"
	}

	var buf strings.Builder
	dumpNode(&buf, node, 0)
	return buf.String()
}

// dumpNode recursively dumps a node
func dumpNode(buf *strings.Builder, node Node, indent int) {
	if node == nil {
		buf.WriteString(strings.Repeat("  ", indent))
		buf.WriteString("nil\n")
		return
	}

	buf.WriteString(strings.Repeat("  ", indent))
	buf.WriteString(node.Type())
	buf.WriteString("(")

	// Handle different node types based on their structure
	switch n := node.(type) {
	case *Template:
		buf.WriteString(fmt.Sprintf("body=%v", n.Body))
	case *Output:
		buf.WriteString(fmt.Sprintf("nodes=%v", n.Nodes))
	case *Const:
		buf.WriteString(fmt.Sprintf("value=%v", n.Value))
	case *Name:
		buf.WriteString(fmt.Sprintf("name=%s, ctx=%s", n.Name, n.Ctx))
	case *BinExpr:
		buf.WriteString(fmt.Sprintf("left=%v, right=%v, operator=%s", n.Left, n.Right, n.Operator))
	case *UnaryExpr:
		buf.WriteString(fmt.Sprintf("node=%v, operator=%s", n.Node, n.Operator))
	case *Call:
		buf.WriteString(fmt.Sprintf("node=%v, args=%v, kwargs=%v", n.Node, n.Args, n.Kwargs))
	case *Getattr:
		buf.WriteString(fmt.Sprintf("node=%v, attr=%s, ctx=%s", n.Node, n.Attr, n.Ctx))
	case *Getitem:
		buf.WriteString(fmt.Sprintf("node=%v, arg=%v, ctx=%s", n.Node, n.Arg, n.Ctx))
	case *List:
		buf.WriteString(fmt.Sprintf("items=%v", n.Items))
	case *Dict:
		buf.WriteString(fmt.Sprintf("items=%v", n.Items))
	case *For:
		buf.WriteString(fmt.Sprintf("target=%v, iter=%v, body=%v", n.Target, n.Iter, n.Body))
	case *If:
		buf.WriteString(fmt.Sprintf("test=%v, body=%v", n.Test, n.Body))
	case *Assign:
		buf.WriteString(fmt.Sprintf("target=%v, node=%v", n.Target, n.Node))
	default:
		// Generic case - just use String() representation
		buf.WriteString(n.String())
	}

	buf.WriteString(")\n")

	// Dump children
	for _, child := range node.GetChildren() {
		dumpNode(buf, child, indent+1)
	}
}

// Type checking and casting utilities

// AsStmt attempts to cast a Node to Stmt
func AsStmt(node Node) (Stmt, bool) {
	if stmt, ok := node.(Stmt); ok {
		return stmt, true
	}
	return nil, false
}

// AsExpr attempts to cast a Node to Expr
func AsExpr(node Node) (Expr, bool) {
	if expr, ok := node.(Expr); ok {
		return expr, true
	}
	return nil, false
}

// AsHelper attempts to cast a Node to Helper
func AsHelper(node Node) (Helper, bool) {
	if helper, ok := node.(Helper); ok {
		return helper, true
	}
	return nil, false
}

// IsStmt checks if a node is a statement
func IsStmt(node Node) bool {
	_, ok := node.(Stmt)
	return ok
}

// IsExpr checks if a node is an expression
func IsExpr(node Node) bool {
	_, ok := node.(Expr)
	return ok
}

// IsHelper checks if a node is a helper
func IsHelper(node Node) bool {
	_, ok := node.(Helper)
	return ok
}

// Node creation helpers

// NewConst creates a new Const node
func NewConst(value interface{}, line, column int) *Const {
	node := &Const{Value: value}
	node.SetPosition(Position{Line: line, Column: column})
	return node
}

// NewName creates a new Name node
func NewName(name, ctx string, line, column int) *Name {
	node := &Name{Name: name, Ctx: ctx}
	node.SetPosition(Position{Line: line, Column: column})
	return node
}

// NewTemplate creates a new Template node
func NewTemplate(body []Node) *Template {
	return &Template{Body: body}
}

// NewOutput creates a new Output node
func NewOutput(nodes []Expr, line, column int) *Output {
	node := &Output{Nodes: nodes}
	node.SetPosition(Position{Line: line, Column: column})
	return node
}
