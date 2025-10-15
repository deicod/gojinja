# Feature Parity Roadmap

## Stage 1 – Core Compatibility (High Severity items)

1. **Translation Tags**
   - Implement `{% trans %}`, `{% pluralize %}`, and `{% blocktrans %}` with runtime pluralisation hooks.
   - Add regression tests covering Django/Flask i18n templates.
2. **Async Control Flow**
   - Parse and evaluate `async for` / `async with` blocks under an `enable_async` switch.
   - Provide no-op fallbacks in synchronous environments to ease migration.
3. **Conformance Harness**
   - Port a representative slice of the upstream Jinja test suite (statements, filters, runtime) and wire into CI.
   - Track parity metrics per feature area to surface regressions quickly.

## Stage 2 – Advanced Runtime Alignment

1. **Caching & Bytecode**
   - Introduce a bytecode cache abstraction with filesystem mtime invalidation.
   - Allow pluggable cache stores that mirror Jinja's `FileSystemBytecodeCache` and friends.
2. **Async & Streaming Rendering**
   - Add async-aware filters/tests/globals and expose streaming APIs akin to `Template.generate()`.
   - Evaluate goroutine-based rendering helpers for concurrent output.
3. **Undefined Policies & Expressions**
   - Expand undefined variants (chainable, silent) and wire expression helpers like `environment()` / `context()`.
4. **Security Hardening**
   - Extend sandbox enforcement to cover the full filter/test/global matrix and improve violation diagnostics.

## Stage 3 – Ecosystem Enhancements

1. **Macro Contracts & Module Exports**
   - Enforce keyword-only/varargs ordering and expose compiled template modules for reuse.
2. **Filter Polish**
   - Add remaining keyword behaviours (e.g. `sum(attribute=...)`) and richer coercions where Python allows them.
3. **Whitespace Edge Cases**
   - Cover nuanced `lstrip_blocks`/`keep_trailing_newline` scenarios with fixture-backed tests.
4. **Documentation Refresh**
   - Produce migration notes for Python users covering new features and behavioural differences.

## Stage 4 – Continuous Validation

- Automate sync runs against upstream Jinja2 tests on a schedule.
- Publish parity dashboards that highlight newly covered or regressed features.
- Establish a deprecation policy aligned with Python releases.
