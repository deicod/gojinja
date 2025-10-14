# Go Jinja2 Runtime Engine

A production-ready runtime engine for evaluating Jinja2 templates in Go, providing comprehensive template rendering capabilities with excellent error handling and performance.

## Features

### Core Components
- **Environment**: Manages configuration, globals, filters, and template loading
- **Context**: Handles variable scopes and context during rendering
- **Template**: Compiled template ready for rendering
- **Evaluator**: AST visitor pattern implementation for template evaluation

### Template Features
- ✅ Variable substitution and expressions
- ✅ Control structures: `{% for %}`, `{% if %}`, `{% elif %}`, `{% else %}`
- ✅ Variable assignment: `{% set %}`
- ✅ Comprehensive filters and tests
- ✅ Arithmetic and logical operations
- ✅ Attribute and item access
- ✅ Loop variables and iteration
- ✅ Autoescaping support
- ✅ Custom filters and functions

### Built-in Filters

**String Filters:**
- `upper`, `lower`, `capitalize`, `title`
- `trim`, `ltrim`, `rtrim`, `strip`
- `striptags`, `replace`, `truncate`
- `wordcount`, `reverse`, `center`, `indent`

**Number Filters:**
- `round`, `abs`, `int`, `float`
- `default` - with sensible defaults

**List Filters:**
- `length`, `first`, `last`, `join`
- `sort`, `unique`, `min`, `max`, `sum`
- `list`, `slice`, `batch`, `groupby`

**Utility Filters:**
- `safe`, `escape`, `e`, `urlencode`
- `attr`, `map`, `select`, `reject`
- `selectattr`, `rejectattr`

### Built-in Tests

- `divisibleby`, `defined`, `undefined`, `none`
- `boolean`, `number`, `string`, `sequence`
- `mapping`, `iterable`, `callable`, `sameas`
- `lower`, `upper`, `even`, `odd`, `in`

### Global Functions

- `range(start, stop, step)` - Generate number sequences
- `dict(key1, val1, key2, val2, ...)` - Create dictionaries
- `lipsum()` - Generate lorem ipsum text
- `cycler(item1, item2, ...)` - Create cycling iterator
- `joiner(separator)` - Create string joiner

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/deicod/gojinja/runtime"
)

func main() {
    // Simple template rendering
    template := "Hello {{ name|upper }}!"
    context := map[string]interface{}{
        "name": "world",
    }

    result, err := runtime.ExecuteToString(template, context)
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // "Hello WORLD!"
}
```

### Using Environment

```go
// Create environment with custom configuration
env := runtime.NewEnvironment()
env.SetAutoescape(true)

// Add custom filter
env.AddFilter("shout", func(ctx *runtime.Context, value interface{}, args ...interface{}) (interface{}, error) {
    if str, ok := value.(string); ok {
        return strings.ToUpper(str) + "!!!", nil
    }
    return value, nil
})

// Parse and render template
template, err := env.ParseString("Hello {{ name|shout }}", "greeting")
if err != nil {
    panic(err)
}

result, err := template.ExecuteToString(map[string]interface{}{"name": "go"})
if err != nil {
    panic(err)
}

fmt.Println(result) // "Hello GO!!!"
```

### Batch Rendering

```go
// Create batch renderer for multiple templates
renderer := runtime.NewBatchRenderer(env)

// Add templates
renderer.AddTemplate("welcome", "Welcome, {{ name }}!")
renderer.AddTemplate("goodbye", "Goodbye, {{ name }}!")

// Render templates
welcome, _ := renderer.Render("welcome", map[string]interface{}{"name": "Alice"})
goodbye, _ := renderer.Render("goodbye", map[string]interface{}{"name": "Bob"})

fmt.Println(welcome) // "Welcome, Alice!"
fmt.Println(goodbye) // "Goodbye, Bob!"
```

### Error Handling

```go
result, err := runtime.ExecuteToString("Hello {{ undefined_var }}!", nil)
if err != nil {
    // Error includes position information and detailed context
    fmt.Printf("Error: %v\n", err)

    // Check error types
    if runtime.IsUndefinedError(err) {
        fmt.Println("Variable was undefined")
    }
}
```

## Architecture

### Component Overview

1. **Environment**: Central configuration hub
   - Manages filters, tests, and globals
   - Handles template loading and caching
   - Controls autoescaping and security settings

2. **Context**: Runtime state management
   - Variable scoping with nested contexts
   - Loop state tracking (loop.index, loop.first, etc.)
   - Error collection and context inheritance

3. **Template**: Compiled template representation
   - AST-based evaluation
   - Thread-safe rendering
   - Memory efficient execution

4. **Evaluator**: AST visitor implementation
   - Traverses template AST
   - Evaluates expressions and statements
   - Handles control flow and variable resolution

### Error Handling

The runtime provides comprehensive error handling with:

- **Position Information**: Template line and column numbers
- **Error Types**: Specific error types for different failure modes
- **Error Context**: Rich error messages with debugging information
- **Graceful Degradation**: Handles missing values and undefined variables

## Performance

- **Thread-Safe**: Multiple goroutines can render templates concurrently
- **Memory Efficient**: Minimal allocations during rendering
- **AST Caching**: Templates can be compiled once and reused
- **Lazy Evaluation**: Expressions evaluated only when needed

## Examples

See the `examples/` directory for comprehensive usage examples:

- `simple_demo.go` - Basic usage examples
- `runtime_demo.go` - Advanced features and API demonstrations

## Testing

Run the test suite:

```bash
go test ./runtime -v
```

Run benchmarks:

```bash
go test ./runtime -bench=.
```

## Compatibility

This runtime engine is compatible with Jinja2 syntax and provides most core functionality. Some advanced features like template inheritance, macros, and custom tags are planned for future releases.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[Add your license information here]