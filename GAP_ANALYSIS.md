# Gap Analysis vs. Jinja2 Reference

Severity scale: **High** – breaks common templates or core semantics, **Medium** – limits advanced features or ecosystem compatibility, **Low** – niche parity gaps or optional optimisations.

| Area | Gap | Severity | Notes / References |
| --- | --- | --- | --- |
| Parser | Translation/i18n tags (`{% trans %}`, `{% pluralize %}`, `{% blocktrans %}`) align with Jinja2 (context strings, trimming, plural hooks) | Resolved | Implemented in `parser/statements.go` and `runtime/evaluator.go` |
| Parser | Async statements (`async for`, `async with`) unsupported | Resolved | Environment flag enables parsing with synchronous execution fallbacks |
| Runtime | Bytecode cache supports loader-aware invalidation | Resolved | Bytecode cache API with modtime validation in `runtime/environment.go`, `runtime/bytecode_cache.go` |
| Runtime | Async rendering & streaming APIs unavailable | Medium | No equivalent to `generate()` or async render pipeline |
| Macros | Keyword-only/varargs validation, exported template modules incomplete | Resolved | Macro registry enforces argument contracts and module exports support shared contexts (`runtime/macro.go`, `runtime/template.go`) |
| Expressions | Async/await expressions (`await`, async filters/tests) unsupported | Resolved | Parser/evaluator support await expressions and auto-await async helpers |
| Security | Sandbox enforcement now covers filters/tests/globals with explicit test allow/block policies | Resolved | Policy builder and evaluator enforcement in `runtime/policy.go`, `runtime/security.go`, `runtime/evaluator.go`, `runtime/filters.go` |
| Errors | Stack traces lack full context chain seen in Python | Low | `runtime/errors.go` records position but not multi-frame call stacks |
| Tooling | No upstream conformance harness synced from Jinja2 | High | Without regression tests parity regressions go unnoticed |
