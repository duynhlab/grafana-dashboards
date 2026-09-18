# Testing

## Commands

```bash
make test        # go test ./...
make coverage    # repository-wide coverage, fails under 90%
make lint        # gofmt -l must be empty, go vet
make generate    # go run ./cmd/generate
make validate    # lint + coverage + generate + git diff --exit-code generated/ deploy/
```

`make validate` is the definition of done. CI runs the same steps on every pull request.
The 90% gate is repository-wide (`-coverpkg=./...`), so an untested new dashboard file
lowers the total. Every dashboard ships with a test.

## Test layout

```text
test/dashboards/<name>_test.go     one per dashboard (package dashboards_test)
test/dashboards/alerts_test.go     TestAlertRegistry, TestDomainFolders
internal/generate/output_test.go   renderer: determinism, stale-file removal, OCI env, error paths
```

Use `test/dashboards/pgdog_test.go` as the template for a dashboard test.

## What to assert

Dashboards:
- `Build()` succeeds, `Title` matches
- `len(dash.Variables)` and `len(dash.Elements)` (panel count) match the design
- `dashboardv2.Manifest(uid, builder).Build()` gives `Metadata.Name == uid`
- marshalled JSON contains the key metric names (two or three fragments, not every query)

Alerts (covered by `TestAlertRegistry` for all rules):
- UID unique, required fields set
- labels `severity`, `domain`, `component`
- annotations `summary`, `description`, `dashboard_uid` resolving to a registered dashboard

Folders: `TestDomainFolders` asserts each `standards.Folder*` constant is produced by
`registry.Folders()` and that no dashboard uses a catch-all folder. Extend the `want`
map when adding a folder.

Queries: metric name, aggregation, label set, no environment-specific values.

Golden JSON is not used. `generated/` committed output plus `git diff --exit-code` in
`make validate` plays that role.

## Related

- [cicd.md](cicd.md) for what CI runs and publishes
- [e2e-kind.md](e2e-kind.md) for the operator smoke test
