# Go Jinja2 AST Node System

This package provides a comprehensive Abstract Syntax Tree (AST) node system for the Go Jinja2 implementation, based on the Python Jinja2 nodes.py structure.

## Overview

The AST system defines all the necessary node types for representing parsed Jinja2 templates, including:

- **Statements** (Stmt): Control flow and template structure nodes
- **Expressions** (Expr): Values and computations
- **Helpers** (Helper): Context-specific nodes
- **Template**: The root node of every AST

## Core Components

### Base Node Interface

All nodes implement the `Node` interface:

```go
type Node interface {
    GetPosition() Position
    SetPosition(pos Position)
    GetChildren() []Node
    Accept(visitor Visitor) interface{}
    String() string
    Type() string
}
```

### Position Tracking

Every node maintains position information for error reporting:

```go
type Position struct {
    Line   int `json:"line"`
    Column int `json:"column"`
}
```

### Visitor Pattern

The system implements the visitor pattern for AST traversal:

```go
type Visitor interface {
    Visit(node Node) interface{}
}

// Example: Count all Name nodes
visitor := NodeVisitorFunc(func(node Node) interface{} {
    if _, ok := node.(*Name); ok {
        nameCount++
    }
    return nil
})

Walk(visitor, template)
```

## Node Categories

### Statement Nodes (Stmt)

- **Template**: Root template node containing body
- **Output**: Multiple expressions to be printed
- **For**: Loop statements with target, iterable, body, and optional else
- **If**: Conditional statements with test, body, elifs, and else
- **Macro**: Function definitions with name, arguments, and body
- **CallBlock**: Anonymous function calls
- **FilterBlock**: Filter application to content blocks
- **With**: Context variable assignment
- **Block**: Named template blocks
- **Include**: Template inclusion
- **Import**: Template import with target
- **FromImport**: Selective import from templates
- **Assign**: Variable assignment
- **AssignBlock**: Block assignment with optional filtering
- **Continue/Break**: Loop control statements
- **Scope**: Artificial scope boundaries

### Expression Nodes (Expr)

- **Const**: Constant values (numbers, strings, booleans)
- **Name**: Variable references with context (load/store/param)
- **NSRef**: Namespace references
- **TemplateData**: Raw template strings
- **List**: List literals `[1, 2, 3]`
- **Dict**: Dictionary literals `{key: value}`
- **Tuple**: Tuple literals `(1, 2, 3)`
- **BinExpr**: Binary expressions (`+`, `-`, `*`, `/`, `and`, `or`, etc.)
- **UnaryExpr**: Unary expressions (`not`, `-`, `+`)
- **Call**: Function calls with arguments and keyword arguments
- **Getattr**: Attribute access `obj.attr`
- **Getitem**: Item access `obj[key]`
- **Slice**: Slice objects `start:stop:step`
- **Compare**: Comparison expressions `==`, `!=`, `<`, `>`, `in`, etc.
- **CondExpr**: Conditional expressions `a if condition else b`
- **Concat**: String concatenation
- **Filter/FilterTestCommon**: Filter and test applications
- **Filter**: Filter application
- **Test**: Test application

### Helper Nodes (Helper)

- **Pair**: Key-value pairs for dictionaries
- **Keyword**: Keyword arguments for function calls
- **Operand**: Comparison operands
- **ContextReference/DerivedContextReference**: Template context access
- **EnvironmentAttribute**: Environment attribute access
- **ExtensionAttribute**: Extension attribute access
- **ImportedName**: Imported name references
- **InternalName**: Compiler internal names
- **MarkSafe/MarkSafeIfAutoescape**: Safety marking for autoescaping
- **EvalContextModifier/ScopedEvalContextModifier**: Context modification

## Context Types

Expressions support different contexts:

```go
const (
    CtxLoad  = "load"  // Reading a value
    CtxStore = "store" // Writing a value
    CtxParam = "param" // Function parameter
)
```

## Usage Examples

### Creating a Simple AST

```go
// Create nodes: x + 5
x := NewName("x", CtxLoad, 1, 1)
five := NewConst(5, 1, 5)
add := NewAdd(x, five)

// Create output and template
output := NewOutput([]Expr{add}, 1, 1)
template := NewTemplate([]Node{output})
```

### Visitor Pattern Usage

```go
// Find all variables
var variables []string
visitor := NodeVisitorFunc(func(node Node) interface{} {
    if name, ok := node.(*Name); ok {
        variables = append(variables, name.Name)
    }
    return nil
})

Walk(visitor, template)
```

### Constant Evaluation

```go
ctx := NewEvalContext(env, "template.html")
value, err := node.AsConst(ctx)
if err != nil {
    // Handle error - node cannot be evaluated as constant
}
```

### AST Manipulation

```go
// Set context for assignment targets
SetCtx(template, CtxStore)

// Set line numbers
SetLineNo(template, 10, false)

// Find specific nodes
addNode := Find(template, "Add")
allNames := FindAll(template, "Name")
```

### Type Checking

```go
// Check node types
if IsExpr(node) {
    if expr, ok := AsExpr(node); ok {
        // Work with expression
    }
}

if IsStmt(node) {
    // Handle statement
}
```

## Utility Functions

### Dumping AST

```go
// Print AST structure for debugging
fmt.Println(Dump(template))
```

### Truthiness Evaluation

```go
// Evaluate Jinja2 truthiness rules
if isTruthy(value) {
    // Value is truthy in Jinja2 context
}
```

## Evaluation Context

The `EvalContext` provides runtime information for constant evaluation:

```go
type EvalContext struct {
    Environment interface{} // Will be *environment.Environment
    AutoEscape  bool        // Autoescaping enabled
    Volatile    bool        // Volatile context (prevents const evaluation)
}
```

## Design Principles

1. **Type Safety**: Strong typing with Go interfaces
2. **Extensibility**: Easy to add new node types
3. **Performance**: Efficient visitor pattern and traversal
4. **Maintainability**: Clear separation of concerns
5. **Compatibility**: Matches Python Jinja2 behavior
6. **Position Tracking**: Complete source position information
7. **Context Awareness**: Proper load/store context handling

## Integration

This AST system is designed to work with:
- **Lexer**: Provides token position information
- **Parser**: Builds AST from tokens
- **Compiler**: Generates code from AST
- **Runtime**: Executes compiled templates

## Testing

The package includes comprehensive tests covering:
- Node creation and manipulation
- Visitor pattern implementation
- Constant evaluation
- Type checking and casting
- AST traversal and searching
- Position tracking
- Context handling

Run tests with:
```bash
go test ./nodes -v
```

## Example

See `../examples/ast_demo.go` for a complete demonstration of the AST system in action.