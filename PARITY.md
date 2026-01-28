# Feature Coverage Snapshot

## Statements and Tags

- Core control tags match Jinja2 semantics: `autoescape`, `block`, `break`, `continue`, `do`, `extends`, `for`, `if`, `import`, `include`, `from`, `macro`, `print`, `set`, and `with` are recognised by the parser (`parser/parser.go`).
- Ancillary blocks – including `call`, `filter`, and `spaceless` – reuse Python's stack discipline and end-token validation (`parser/core.go`).
- Raw/verbatim blocks, whitespace trimming markers, and comment controls all flow through the lexer with support for `trim_blocks`, `lstrip_blocks`, `keep_trailing_newline`, and the configurable line statement/comment prefixes (`lexer/lexer.go`, `parser/parser.go`, `runtime/environment.go`).
- Template inheritance implements block resolution and `super()` lookups in line with Jinja2 (`runtime/template.go`).
- Extension hooks allow custom tags to be registered at the environment level, and participate in parsing (`runtime/environment.go`, `parser/parser.go`).
- Translation tags (`{% trans %}`/`{% blocktrans %}`) mirror Jinja2's context, trimming, and pluralisation semantics with runtime gettext/npgettext dispatch (`parser/statements.go`, `runtime/evaluator.go`).
- Async control flow tags (`async for`, `async with`) are parsed when `enable_async` is activated on the environment and execute with synchronous fallbacks that match Jinja2's behaviour in non-async contexts (`parser/core.go`, `runtime/environment.go`, `runtime/evaluator.go`).

## Expression & Assignment Support

- Arithmetic, comparison, logical operators, slicing, attribute/item access, test/filter pipes, and ternary expressions are available through the node tree (`nodes/nodes.go`).
- Tuple/list/dict literals, macro calls, positional/keyword argument binding, unpacking assignment targets, and namespace references mirror Python Jinja behaviour (`parser/expressions.go`, `runtime/evaluator.go`).
- Helper expressions for inspecting runtime state are provided via the builtin `environment()` and `context()` globals, returning the active environment and a snapshot of the scope (`runtime/environment.go`, `runtime/context.go`).

**Remaining gaps**: async rendering modes are still missing.

## Built-in Filters

- String, list, numeric, and utility filters now cover the standard library, including `abs`, `attr`, `batch`, `capitalize`, `center`, `default`, `dictsort`, `dictsortcasesensitive`, `dictsortreversed`, `escape`/`e`, `escapejs`, `filesizeformat`, `filter`, `first`, `float`, `floatformat`, `forceescape`, `format`, `fromjson`, `groupby`, `indent`, `int`, `join`, `last`, `length`, `list`, `lower`, `ltrim`, `map`, `max`, `min`, `pprint`, `random`, `reject`, `rejectattr`, `replace`, `reverse`, `round`, `safe`, `select`, `selectattr`, `slice`, `sort`, `striptags`, `sum` (with `attribute` and `start` support), `title`, `trim`, `truncate`, `unique`, `upper`, `urlencode`, `urlize`, `wordcount`, `wordwrap`, `xmlattr`, `tojson`, and `do` (`runtime/filters.go`).
- Keyword argument handling matches Jinja for filters such as `wordwrap`, `filesizeformat`, `urlize`, the `dictsort` family, and the `sum` filter's `attribute`/`start` options, and the environment newline settings flow into wrapping behaviour (`runtime/filters.go`). When `enable_async` is active, filter results implementing awaitable semantics are resolved automatically, mirroring Python Jinja's async behaviour (`runtime/evaluator.go`).

## Built-in Tests

- The environment registers numeric, sequence, mapping, callability, truthiness, string case, containment, regex, NaN/Inf, undefined, module, and rich comparison aliases (including the symbolic operators). Async-enabled templates transparently await predicate results before truthiness checks (`runtime/filters.go`, `runtime/evaluator.go`).

## Global Functions

- Built-in globals include `range`, `lipsum`, `dict`, `cycler`, `joiner`, `namespace`, `class`, `_`/`gettext`/`ngettext`, `debug`, `self`, `context`, `environment`, and the configurable `url_for` hook, with async-aware results automatically awaited when `enable_async` is set (`runtime/environment.go`, `runtime/context.go`, `runtime/evaluator.go`).

## Macros, Imports, and Namespaces

- Template module exports, including explicit `{% export %}` statements and module creation that shares existing contexts, mirror Jinja's module API (`runtime/evaluator.go`, `runtime/template.go`).

## Whitespace & Data Control

- Environment switches `SetTrimBlocks`, `SetLStripBlocks`, `SetKeepTrailingNewline`, `SetLineStatementPrefix`, and `SetLineCommentPrefix` feed directly into the lexer/parser to match Jinja trimming semantics (`runtime/environment.go`, `parser/parser.go`).
- Markup raw data is preserved and whitespace trimming honours dash/plus syntax across statements, variables, and comments (`lexer/lexer.go`).

**Remaining gaps**: advanced edge cases around `lstrip_blocks` and preserving intentional blank lines still need coverage.

## Environment & Runtime

- File system and map loaders honour multi-path search order, provide `TemplateModTime`, and surface `TemplateNotFound` with tried paths (`runtime/environment.go`).
- Template caching, macro registries, extension registration, autoescape selection, and sandbox-aware execution line up with Python's API surface (`runtime/environment.go`, `runtime/template.go`, `runtime/sandbox.go`).

**Remaining gaps**: streaming writers and async rendering modes are not yet implemented.

## Error Handling & Security

- Runtime errors capture positions and wrap underlying causes, and dedicated `TemplateNotFoundError` / `TemplatesNotFoundError` types align with Jinja expectations (`runtime/errors.go`).
- Security policy builders now include explicit test allow/block controls, and sandbox execution enforces filter/test/global access alongside resource limits (`runtime/policy.go`, `runtime/security.go`, `runtime/evaluator.go`).

**Remaining gaps**: richer stack traces for parity with Python's error diagnostics.
