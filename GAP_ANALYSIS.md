# Gap Analysis vs. Jinja2 Reference

Severity scale: **High** – breaks common templates or core semantics, **Medium** – limits advanced features or ecosystem compatibility, **Low** – niche parity gaps or optional optimisations.

| Area | Gap | Severity | Notes / References |
| --- | --- | --- | --- |
| Parser | Namespace declaration and line-statement prefix unsupported | Medium | Jinja’s `{% namespace %}` and `#` line statements missing in `parser/core.go`, impacting macros and compact templates |
| Runtime | Bytecode cache / template cache policy differs (no loader integration for search paths, mtime checks) | Low | `runtime/environment.go` cache lacks filesystem invalidation; affects production caching parity |
| Runtime | Async rendering (`async for/with`, async filters/tests/globals) absent | Low | Python Jinja supports optional async; Go port executes synchronously |
| Filters | Missing key filters (`filesizeformat`, `floatformat`, `escapejs`, `tojson`, `urlize`, `wordwrap`, `random`, etc.) | High | Many templates depend on these; inventory in `runtime/filters.go` lacks these implementations |
| Tests | Missing comparison/type/pattern tests (`eq`, `ne`, `lt`, `integer`, `mapping`, `escaped`, `matching`, `search`, `infinite`, `nan`) | Medium | Limits existing templates/tests relying on `is` checks |
| Globals | `namespace`, `class`, `gettext`/`ngettext`, debug helpers not exposed | Medium | Blocks advanced macro patterns and i18n usage |
| Extensions | No extension API wiring | Medium | `parser.Environment` lacks registration for filters/tests/globals via extensions, limiting reuse of Jinja ecosystem |
| Security | Sandbox policy incomplete (attribute/filter whitelisting) | Medium | `runtime/security.go` provides stubs but lacks enforcement parity; potential security divergence |
| Errors | Exception hierarchy diverges; `TemplateNotFound`, `TemplatesNotFound` missing | Medium | Error handling compatibility needed for frameworks expecting canonical types |
| Tooling | No conformance test harness vs. Jinja upstream suite | High | Without regression tests parity regressions go unnoticed; required for confidence |
