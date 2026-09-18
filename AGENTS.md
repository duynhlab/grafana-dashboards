# AGENTS.md

Grafana dashboards and alert rules as Go code (Grafana Foundation SDK, dashboardv2),
rendered by `go run ./cmd/generate` and delivered to Grafana 12 through Flux OCI
artifacts and the Grafana Operator.

## Skill

For any dashboard, alert, query, generator, CI, or e2e work, load the project skill
`grafana-foundation-sdk` (canonical copy at `.agents/skills/grafana-foundation-sdk/SKILL.md`,
symlinked into `.claude/skills/` and `.cursor/skills/`). It holds the workflows, helper
catalogue, metric models, and the contract for splitting work across several agents.

## Commands

```bash
make generate    # render generated/ and deploy/ from Go
make test        # go test ./...
make coverage    # repository-wide coverage, must stay >= 90%
make validate    # fmt + vet + coverage + generate + clean diff; the definition of done
make e2e-kind    # Kind + Grafana Operator smoke test, needs Docker
```

## Invariants

- Go under `internal/` is the source of truth. Never hand-edit `generated/` or `deploy/`.
- Never edit legacy JSON under `dashboard/`. It is read-only porting input.
- Register resources only in `internal/registry/`. Do not change `cmd/generate` for a new
  dashboard, alert, or domain.
- Reuse `internal/panels`, `internal/standards`, and existing query constants before
  adding abstractions.
- `make validate` must pass before a change is reported as done. Report its output.

## Parallel work

One orchestrator owns the registry, shared helpers, generated output, Makefile, and
workflows. Workers own one dashboard or alert group each and only their own query,
dashboard, and test files. See `references/multi-agent.md` in the skill.

## Git

Conventional commits (`feat(as-code): ...`, `fix(e2e): ...`, `docs(skill): ...`).
Branch from `main`, open a pull request; CI runs `make validate` equivalents plus the
Kind e2e.
