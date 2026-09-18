# Working with several agents

Read this when a task is large enough to split (porting several boards, adding a domain
with several dashboards and alerts) or when you are a subagent that received a slice.
The rules are harness-neutral: they hold for Claude Code subagents, Codex subagents,
Cursor, or several humans on one branch.

## Why the split is safe

Each dashboard is one Go file plus one query file plus one test file, and nothing else
imports them until the registry does. The only shared write points are
`internal/registry/*.go`, the shared helper packages, and the generated output. Keep
those in one pair of hands and parallel work does not conflict.

## Roles

**Orchestrator** (one agent, usually the one talking to the user)
- owns `internal/registry/`, `internal/standards/`, `internal/panels/`, `internal/generate/`,
  `generated/`, `deploy/`, `Makefile`, `.github/workflows/`, `test/e2e/`
- decides domain, folder, UID, and metric model for every slice before delegating
- hands each worker a brief (below), collects reports, registers resources, runs
  `make generate` once, runs `make validate` once, fixes cross-slice conflicts
- never has two workers on the same dashboard or the same query file

**Worker** (one per dashboard or alert group)
- may create or edit only:
  `internal/queries/prometheus/<domain>/<name>.go`,
  `internal/dashboards/<domain>/<name>.go` (and `<name>_sections.go`),
  `internal/alerts/<domain>.go` only when the brief says the alert group is theirs,
  `test/dashboards/<name>_test.go`
- may read anything; must read the skill's SKILL.md and `domains.md` first
- reuses existing helpers; if a new shared helper or a standards change is needed, it
  writes the proposal in its report and works around it locally
- runs `go build ./... && go test ./test/dashboards/ -run <TestName>` for its slice,
  never `make generate` (that rewrites shared output)
- does not register its resource. It reports the exact registry line instead.

**Reviewer** (optional, read-only)
- checks a finished slice against `testing.md`: panel count matches design, PromQL
  uses the right metric model, no duplicated constants, no environment values, test exists
- reports findings; does not edit

## Brief the orchestrator gives a worker

```text
Slice: <dashboard UID> in folder <Folder> (domain package <domain>)
Source: dashboard/<path>.json (read-only) | new board from spec: <bullets>
Metric model: <CNPG | Pigsty | PGDog | kube-state-metrics | ...>
Files you own: internal/queries/prometheus/<domain>/<name>.go,
               internal/dashboards/<domain>/<name>.go,
               test/dashboards/<name>_test.go
Do not edit: internal/registry, internal/standards, internal/panels, generated, deploy
Exported builder: func <Name>() cog.Builder[dashboardv2.Dashboard]
Done when: go build ./... passes and your test passes; reply with the report format.
```

## Report a worker returns

```text
Slice: <UID>
Builder: <package>.<Name>
Files: <list>
Registry line: {UID: "<uid>", Folder: standards.Folder<X>, Build: <package>.<Name>},
Panels: <n>, Variables: <n>, Test: Test<Name> (passing)
Queries added: <constant names>
Proposed shared changes: <none | description, not applied>
Open questions: <none | ...>
```

## Orchestrator merge steps

1. Add every registry line in one edit of `internal/registry/dashboards.go` (and
   `alerts.go`).
2. Apply accepted shared-helper proposals yourself, then tell affected workers to rebase.
3. `make generate`, review the diff in `generated/` and `deploy/`.
4. `make validate`. If coverage drops under 90%, the slice missing a test goes back
   to its worker.
5. `make e2e-kind` only if `deploy/` shape changed (new folder, new alert group).

## Do not parallelise

- changes to `internal/generate/` or the Makefile
- changes to `internal/standards/` or `internal/panels/` signatures
- anything under `deploy/`, `generated/`, `.github/workflows/`, `test/e2e/`
- two slices that need the same new query constant (merge them or sequence them)

## Delegation hints per harness

| Harness | How to delegate | Notes |
|---|---|---|
| Claude Code | `Agent` tool, one call per slice in the same message so they run concurrently; use a read-only agent type for the reviewer | subagents can invoke this skill through the Skill tool; paste the brief in the prompt |
| Codex | ask for parallel subagents in the prompt ("spawn one subagent per dashboard ..."); inspect with `/agent` | subagents follow AGENTS.md and `$grafana-foundation-sdk`; paste the brief |
| Cursor or a single agent | run slices sequentially in the order above; still keep registry edits for the end | the contract is the same, only the concurrency is gone |

The skill body and the files a worker owns are identical across harnesses. Only the
mechanism that starts a worker differs.
