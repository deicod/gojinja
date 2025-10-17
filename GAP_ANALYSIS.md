# Gap Analysis vs. Jinja2 Reference

Severity scale: **High** – breaks common templates or core semantics, **Medium** – limits advanced features or ecosystem compatibility, **Low** – niche parity gaps or optional optimisations.

| Area | Gap | Severity | Notes / References |
| --- | --- | --- | --- |
| Parser | Translation/i18n tags (`{% trans %}`, `{% pluralize %}`, `{% blocktrans %}`) absent | Medium | Required for Django-style templates and Flask-Babel integrations |
| Parser | Async statements (`async for`, `async with`) unsupported | Medium | Needed for parity with `enable_async` templates |
| Runtime | Bytecode cache and loader invalidation still missing | Low | `runtime/cache.go` only caches templates in-memory without mtime checks |
| Runtime | Async rendering & streaming APIs unavailable | Medium | No equivalent to `generate()` or async render pipeline |
| Macros | Keyword-only/varargs validation, exported template modules incomplete | Medium | Macro registry executes but skips argument contract checks |
| Expressions | Async/await expressions (`await`, async filters/tests) unsupported | Medium | Blocks templates using `enable_async` helpers |
| Security | Sandbox coverage for filters/tests/globals incomplete | Medium | Policy builder exists but enforcement gaps remain in `runtime/security.go` |
| Errors | Stack traces lack full context chain seen in Python | Low | `runtime/errors.go` records position but not multi-frame call stacks |
| Tooling | No upstream conformance harness synced from Jinja2 | High | Without regression tests parity regressions go unnoticed |
