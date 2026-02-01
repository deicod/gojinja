# Gap Analysis vs. Jinja2 Reference

Severity scale: **High** – breaks common templates or core semantics, **Medium** – limits advanced features or ecosystem compatibility, **Low** – niche parity gaps or optional optimisations.

| Area | Gap | Severity | Notes / References |
| --- | --- | --- | --- |
| Parser | Translation/i18n tags (`{% trans %}`, `{% pluralize %}`, `{% blocktrans %}`) align with Jinja2 (context strings, trimming, plural hooks) | Resolved | Implemented in `parser/statements.go` and `runtime/evaluator.go` |
| Parser | Async statements (`async for`, `async with`) unsupported | Resolved | Environment flag enables parsing with synchronous execution fallbacks |
| Runtime | Bytecode cache supports loader-aware invalidation | Resolved | Bytecode cache API with modtime validation in `runtime/environment.go`, `runtime/bytecode_cache.go` |
| Runtime | Streaming renderers and writer helpers | Resolved | `Template.Generate` and `Generate*` helpers stream fragments and honour trailing newline policy (`runtime/stream.go`, `runtime/template.go`, `runtime/api.go`) |
| Macros | Keyword-only/varargs validation, exported template modules incomplete | Resolved | Macro registry enforces argument contracts and module exports support shared contexts (`runtime/macro.go`, `runtime/template.go`) |
| Expressions | Async/await expressions (`await`, async filters/tests) supported | Resolved | Await operator and async-aware filters/tests implemented (`parser/expressions.go`, `runtime/evaluator.go`, `runtime/filters.go`) |
| Security | Sandbox enforcement now covers filters/tests/globals with explicit test allow/block policies | Resolved | Policy builder and evaluator enforcement in `runtime/policy.go`, `runtime/security.go`, `runtime/evaluator.go`, `runtime/filters.go` |
| Errors | Stack traces now include template frames | Resolved | Traceback frames (template name/line/column + macro context) are emitted via `runtime/errors.go`, `runtime/context.go`, `runtime/evaluator.go` |
| Tooling | No upstream conformance harness synced from Jinja2 | High | Without regression tests parity regressions go unnoticed |
