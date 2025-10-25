# Feature Coverage Snapshot

## Statements and Tags

- Core control tags match Jinja2 semantics: `autoescape`, `block`, `break`, `continue`, `do`, `extends`, `for`, `if`, `import`, `include`, `from`, `macro`, `print`, `set`, and `with` are recognised by the parser (`parser/parser.go`).
- Ancillary blocks – including `call`, `filter`, and `spaceless` – reuse Python's stack discipline and end-token validation (`parser/core.go`).
- Raw/verbatim blocks, whitespace trimming markers, and comment controls all flow through the lexer with support for `trim_blocks`, `lstrip_blocks`, `keep_trailing_newline`, and the configurable line statement/comment prefixes (`lexer/lexer.go`, `parser/parser.go`, `runtime/environment.go`).
- Template inheritance implements block resolution and `super()` lookups in line with Jinja2 (`runtime/template.go`).
- Extension hooks allow custom tags to be registered at the environment level, and participate in parsing (`runtime/environment.go`, `parser/parser.go`).
- Translation tags (`{% trans %}`/`{% blocktrans %}`) mirror Jinja2's context, trimming, and pluralisation semantics with runtime gettext/npgettext dispatch (`parser/statements.go`, `runtime/evaluator.go`).

**Remaining gaps**: async constructs (`async for` / `async with`) are not yet implemented.

## Expression & Assignment Support

- Arithmetic, comparison, logical operators, slicing, attribute/item access, test/filter pipes, and ternary expressions are available through the node tree (`nodes/nodes.go`).
- Tuple/list/dict literals, macro calls, positional/keyword argument binding, unpacking assignment targets, and namespace references mirror Python Jinja behaviour (`parser/expressions.go`, `runtime/evaluator.go`).
- Helper expressions for inspecting runtime state are provided via the builtin `environment()` and `context()` globals, returning the active environment and a snapshot of the scope (`runtime/environment.go`, `runtime/context.go`).

**Remaining gaps**: async/await expressions are still missing.

## Built-in Filters

- String, list, numeric, and utility filters now cover the standard library, including `abs`, `attr`, `batch`, `capitalize`, `center`, `default`, `dictsort`, `dictsortcasesensitive`, `dictsortreversed`, `escape`/`e`, `escapejs`, `filesizeformat`, `filter`, `first`, `float`, `floatformat`, `forceescape`, `format`, `fromjson`, `groupby`, `indent`, `int`, `join`, `last`, `length`, `list`, `lower`, `ltrim`, `map`, `max`, `min`, `pprint`, `random`, `reject`, `rejectattr`, `replace`, `reverse`, `round`, `safe`, `select`, `selectattr`, `slice`, `sort`, `striptags`, `sum` (with `attribute` and `start` support), `title`, `trim`, `truncate`, `unique`, `upper`, `urlencode`, `urlize`, `wordcount`, `wordwrap`, `xmlattr`, `tojson`, and `do` (`runtime/filters.go`).
- Keyword argument handling matches Jinja for filters such as `wordwrap`, `filesizeformat`, `urlize`, the `dictsort` family, and the `sum` filter's `attribute`/`start` options, and the environment newline settings flow into wrapping behaviour (`runtime/filters.go`).

**Remaining gaps**: richer type coercions for some filters and async-aware filters remain unimplemented.

## Built-in Tests

- The environment registers numeric, sequence, mapping, callability, truthiness, string case, containment, regex, NaN/Inf, undefined, module, and rich comparison aliases (including the symbolic operators) (`runtime/filters.go`).

**Remaining gaps**: async predicates are not yet available.

## Global Functions

- Built-in globals include `range`, `lipsum`, `dict`, `cycler`, `joiner`, `namespace`, `class`, `_`/`gettext`/`ngettext`, `debug`, `self`, `context`, `environment`, and the configurable `url_for` hook (`runtime/environment.go`, `runtime/context.go`).

**Remaining gaps**: async-aware variants of these helpers remain TODO for full parity with `enable_async` templates.

## Macros, Imports, and Namespaces

- Macro declarations, caller blocks, imports (`import`/`from`), and namespace tracking are covered via the macro registry and import runtime (`parser/statements.go`, `runtime/macro.go`, `runtime/import.go`).
- `with context` / `without context` toggles and `call` blocks propagate scope data appropriately (`runtime/evaluator.go`).
- Variadic argument collectors (`*args`) and keyword dictionaries (`**kwargs`) bind with the same semantics as Jinja macros, including duplicate argument detection (`parser/statements.go`, `runtime/macro.go`).

**Remaining gaps**: macro argument validation (keyword-only ordering, default expression late binding) and template module exports remain TODOs.

## Whitespace & Data Control

- Environment switches `SetTrimBlocks`, `SetLStripBlocks`, `SetKeepTrailingNewline`, `SetLineStatementPrefix`, and `SetLineCommentPrefix` feed directly into the lexer/parser to match Jinja trimming semantics (`runtime/environment.go`, `parser/parser.go`).
- Markup raw data is preserved and whitespace trimming honours dash/plus syntax across statements, variables, and comments (`lexer/lexer.go`).

**Remaining gaps**: advanced edge cases around `lstrip_blocks` and preserving intentional blank lines still need coverage.

## Environment & Runtime

- File system and map loaders honour multi-path search order, provide `TemplateModTime`, and surface `TemplateNotFound` with tried paths (`runtime/environment.go`).
- Template caching, macro registries, extension registration, autoescape selection, and sandbox-aware execution line up with Python's API surface (`runtime/environment.go`, `runtime/template.go`, `runtime/sandbox.go`).

**Remaining gaps**: bytecode cache APIs, streaming writers, and async rendering modes are not yet implemented.

## Error Handling & Security

- Runtime errors capture positions and wrap underlying causes, and dedicated `TemplateNotFoundError` / `TemplatesNotFoundError` types align with Jinja expectations (`runtime/errors.go`).
- Security policy builders, sandbox environments, and policy enforcement hooks are in place for filter/test/global whitelisting and resource limits (`runtime/security.go`, `runtime/environment.go`).

**Remaining gaps**: the sandbox still needs comprehensive filter/test coverage and richer stack traces for parity with Python's error diagnostics.
