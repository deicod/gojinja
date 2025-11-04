# Template Inheritance System

This document describes the comprehensive template inheritance system implemented for the Go-based Jinja2 template engine.

## Overview

The inheritance system provides full Jinja2-compatible template inheritance with support for:
- `{% extends "parent.html" %}` - Template extension
- `{% block name %}...{% endblock %}` - Block definitions and overriding
- `{{ super() }}` - Calling parent block content
- Nested inheritance chains
- Template caching with TTL support
- Circular dependency detection

## Features

### 1. Template Extension

Templates can extend other templates using the `extends` tag:

```html
<!-- child.html -->
{% extends "base.html" %}

{% block content %}
    <h1>Hello World</h1>
{% endblock %}
```

### 2. Block System

Blocks allow templates to define sections that can be overridden by child templates:

```html
<!-- base.html -->
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    {% block header %}<header>Default Header</header>{% endblock %}
    {% block content %}{% endblock %}
    {% block footer %}<footer>Default Footer</footer>{% endblock %}
</body>
</html>
```

### 3. Super Function

The `super()` function allows child templates to call parent block content:

```html
{% extends "base.html" %}

{% block content %}
    {{ super() }}
    <p>Additional child content</p>
{% endblock %}
```

### 4. Nested Inheritance

Templates can extend other templates that extend yet more templates:

```html
<!-- grandchild.html -->
{% extends "child.html" %}

{% block content %}
    {{ super() }}
    <p>Grandchild content</p>
{% endblock %}
```

### 5. Block Scoping

Blocks can be scoped to prevent variable leakage:

```html
{% block content scoped %}
    {% set local_var = "This won't leak" %}
    {{ local_var }}
{% endblock %}
```

## Usage

### Basic Setup

```go
package main

import (
    "github.com/deicod/gojinja/runtime"
)

func main() {
    // Create environment
    env := runtime.NewEnvironment()

    // Set up loader
    loader := runtime.NewFileSystemLoader("./templates")
    env.SetLoader(loader)

    // Parse and render template
    tmpl, err := env.ParseFile("child.html")
    if err != nil {
        panic(err)
    }

    result, err := tmpl.ExecuteToString(map[string]interface{}{
        "title": "My Page",
        "content": "Hello World",
    })
    if err != nil {
        panic(err)
    }

    println(result)
}
```

### Using Map Loader

```go
templates := map[string]string{
    "base.html": `<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    {% block content %}{% endblock %}
</body>
</html>`,

    "page.html": `{% extends "base.html" %}
{% block title %}{{ page_title }}{% endblock %}
{% block content %}<h1>{{ page_title }}</h1><p>{{ content }}</p>{% endblock %}`,
}

loader := runtime.NewMapLoader(templates)
env.SetLoader(loader)
```

### Template Caching

```go
// Set cache TTL
env.SetCacheTTL(1 * time.Hour)

// Clear cache
env.ClearCache()

// Check cache size
size := env.CacheSize()
```

## Architecture

### Components

1. **TemplateCache** (`cache.go`)
   - Thread-safe template caching
   - TTL support
   - Dependency tracking
   - Cache invalidation

2. **InheritanceResolver** (`inheritance.go`)
   - Resolves inheritance chains
   - Handles circular dependency detection
   - Manages super() function context

3. **InheritanceContext** (`inheritance.go`)
   - Tracks current block context
   - Manages super() call stack
   - Stores parent block references

4. **Environment Extensions** (`environment.go`)
   - Template loading integration
   - Cache management
   - Parser integration

### Flow

1. **Template Loading**: Templates are loaded through the loader interface
2. **Parsing**: Templates are parsed into AST nodes
3. **Inheritance Resolution**: The inheritance chain is resolved recursively
4. **Cache Storage**: Resolved templates are cached with dependencies
5. **Execution**: Templates are executed with proper inheritance context

## API Reference

### Environment Methods

- `SetLoader(loader Loader)` - Set the template loader
- `ParseFile(name string) (*Template, error)` - Parse a template file
- `LoadTemplate(name string) (*Template, error)` - Load a template (with caching)
- `SetCacheTTL(ttl time.Duration)` - Set cache time-to-live
- `ClearCache()` - Clear the template cache
- `CacheSize() int` - Get current cache size

### Loader Types

- `FileSystemLoader` - Load templates from filesystem
- `MapLoader` - Load templates from in-memory map

### Template Methods

- `Execute(vars map[string]interface{}, writer io.Writer) error` - Execute template
- `ExecuteToString(vars map[string]interface{}) (string, error)` - Execute to string
- `RenderBlock(blockName string, vars map[string]interface{}, writer io.Writer) error` - Render specific block
- `RenderBlockToString(blockName string, vars map[string]interface{}) (string, error)` - Render block to string

## Error Handling

The inheritance system provides comprehensive error handling:

- **Circular Dependencies**: Detected and reported with chain information
- **Missing Templates**: Clear error messages for missing parent templates
- **Multiple Extends**: Error when multiple extends statements are found
- **Invalid Super Calls**: Error when super() is called outside valid context

## Performance Considerations

1. **Caching**: Templates are cached after inheritance resolution
2. **Lazy Loading**: Parent templates are only loaded when needed
3. **Dependency Tracking**: Cache invalidation tracks file dependencies
4. **Memory Management**: Cache size limits prevent memory bloat

## Testing

Comprehensive tests cover:

- Basic inheritance scenarios
- Nested inheritance chains
- Super() functionality
- Error conditions
- Cache behavior
- Block scoping
- Variable contexts

Run tests with:
```bash
go test ./runtime/... -v
```

## Limitations

1. **Dynamic Extends**: Template names in extends must be compile-time constants
2. **Include Support**: Template include functionality not yet implemented
3. **Macro Import**: Macro import/export functionality implemented via the import manager and module exports

## Future Enhancements

1. **Template Includes**: Support for `{% include %}` statements
2. **Macro System**: Full macro import/export system
3. **Dynamic Inheritance**: Support for dynamic template names in extends
4. **Advanced Caching**: File watching and automatic cache invalidation
5. **Performance Optimization**: Template bytecode compilation