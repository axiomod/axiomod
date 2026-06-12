# Audit Records

Audit reports and their remediation plans live here, **organized by audit-id** — one
directory per audit. Never place audit files directly under `docs/`.

```
docs/audit/<audit-id>/        # audit-id format: YYYY-MM-DD-<short-slug>
  feature-implementation-audit.md   # findings
  launch-task-plan.md               # remediation tasks (when applicable)
```

| Audit ID | Scope | Documents |
|---|---|---|
| `2026-06-12-launch-readiness` | Documented-vs-implemented feature audit + launch-readiness task plan | [audit](./2026-06-12-launch-readiness/feature-implementation-audit.md) · [task plan](./2026-06-12-launch-readiness/launch-task-plan.md) |
