# Feature Coverage Snapshot

## Statements and Tags

- Core keywords (`parser/parser.go:196`): `autoescape`, `block`, `break`, `continue`, `do`, `extends`, `for`, `if`, `import`, `include`, `from`, `macro`, `print`, `set`, `with`
- Additional handlers (`parser/core.go:63`): `call` blocks, `filter` blocks; tag stack and end-token validation mirror Python semantics
- Control helpers: `elif`/`else` branches inside `if`, `else` inside `for`, scoped `block` with `required` flag (`parser/core.go:246`)
- Comment/whitespace control: lexer honors dash/plus trimming syntax and environment flags `trim_blocks`, `lstrip_blocks`, `keep_trailing_newline` (`lexer/lexer.go:682`, `runtime/environment.go:76`)
- Raw/verbatim blocks (including whitespace-controlled variants) are preserved as literal template data in both parser and runtime
- Inheritance flow matches block resolution and `super()` support (`runtime/template.go:64`)
- Missing/partial tags: i18n tags (`{% trans %}`, `{% pluralize %}`, `{% blocktrans %}`), `{% spaceless %}`, line statements (`#` prefix), async constructs (`async for/with`), and extension hook registration

## Expression & Assignment Support

- Arithmetic and logical expressions (`nodes/nodes.go`): add/sub/mul/div/floor div/mod/pow, `and`, `or`, `not`, chained comparisons, ternary (`if ... else ...`)
- Data structures: tuple, list, dict literals; slices; attribute/item access (`nodes.Getattr`, `nodes.Getitem`), namespace refs, tests (`is`)
- Function calls filter/test pipelines, macro invocation, keyword/dynamic args; tuple unpacking assignment
- Missing: explicit async/await expressions, constant folding helpers, keyword-only enforcement, environment()/context expression helpers, expression-level whitespace control parity

## Built-in Filters (`runtime/filters.go`)

Implemented: `abs`, `attr`, `batch`, `capitalize`, `center`, `default`, `escape`/`e`, `first`, `float`, `forceescape`, `format` (including dict/keyword arguments), `groupby`, `indent`, `int`, `join`, `last`, `length`, `list`, `lower`, `ltrim`, `map` (positional and `attribute=` form), `max`, `min`, `reverse`, `replace`, `reject`, `rejectattr`, `rtrim`, `safe`, `select`, `selectattr`, `slice`, `sort`, `strip`, `striptags`, `sum`, `title`, `trim`, `truncate`, `unique`, `upper`, `urlencode`, `urlize` (policies, extra schemes, email handling), `wordcount`, `xmlattr`

Remaining/high priority: statement filters (`do`), completeness for collection helpers (`slice` column balancing edge cases), async-compatible filters, formatting utilities like `wordwrap`, and parity for environment-driven policies beyond `urlize`

## Built-in Tests (`runtime/filters.go`)

Implemented: `boolean`, `true`, `false`, `callable`, `defined`, `divisibleby`, `even`, `float`, `in`, `infinite`, `integer`, `iterable`, `list`, `lower`, `mapping`, `matching`, `nan`, `none`/`null`, `number`, `odd`, `sameas`, `search`, `sequence`, `string`, `startingwith`, `endingwith`, `containing`, `tuple`, `dict`, `undefined`, `upper`, `escaped`, `filter`, `test`, rich comparison aliases (`eq`, `ne`, `lt`, `le`, `gt`, `ge`, `equalto`)

Remaining/high priority: symbolic operator aliases such as `>` / `<` rely on parser support and remain pending.

## Global Functions (`runtime/environment.go`)

Implemented: `range`, `lipsum`, `dict`, `cycler`, `joiner`, `namespace`, `_`/`gettext`/`ngettext`, runtime-injected `super`

Missing/high priority: environment/context accessors

## Macros, Imports, and Namespaces

- Macro declarations, call blocks, caller support, import/from import handled with macro registry (`runtime/macro.go`)
- Namespaces tracked for nested imports; evaluation respects `with context` / `without context` toggles (`parser/statements.go`, `runtime/import.go`)
- Missing: macro argument validation parity (keyword-only, positional varargs ordering), `contextfunction` semantics, `namespace()` global, template module exports, deferred evaluation for macro defaults

## Whitespace & Data Control

- Dash/plus trimming recognized for blocks/variables/comments (`lexer/lexer.go:682`)
- Environment options exist but not exposed via public API toggles; default behavior may differ from Python Jinja
- Missing: `keep_trailing_newline` enforcement, line-statement prefix configuration, whitespace preservation in `lstrip_blocks` edge cases

## Environment & Runtime

- Loader abstraction (filesystem/map), template cache, security policy hooks, sandbox evaluator, context stack with loop tracking (`runtime/context.go`, `runtime/environment.go`, `runtime/sandbox.go`)
- Template class supports inheritance, macro/block registries, autoescape decisions (`runtime/template.go`)
- Missing/high priority: bytecode cache API, extension registration/loading, async rendering mode, streaming writer API, policy enforcement parity

## Error Handling & Security

- Errors capture position info and wrap causes (`runtime/errors.go`); security manager skeleton handles recursion/timeouts (`runtime/security.go`)
- Missing: full exception hierarchy alignment (TemplateRuntimeError, TemplateNotFound, TemplatesNotFound), undefined variable strategy parity, attribute access controls, sandbox policy coverage for filters/tests/globals, line/column accuracy, context-rich stack traces
