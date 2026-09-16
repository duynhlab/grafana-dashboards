.PHONY: generate test coverage fmt lint generated-check validate e2e-kind all

generate:
	go run ./cmd/generate

test:
	go test ./...

coverage:
	go test ./... -coverpkg=./... -coverprofile=/tmp/grafana-dashboards-coverage.out
	@total="$$(go tool cover -func=/tmp/grafana-dashboards-coverage.out | awk '/^total:/ {gsub("%", "", $$3); print $$3}')"; \
	echo "total coverage: $${total}% (required: 90.0%)"; \
	awk -v actual="$${total}" 'BEGIN { if ((actual + 0) < 90.0) exit 1 }'

fmt:
	gofmt -w .

lint:
	test -z "$$(gofmt -l .)"
	go vet ./...

generated-check: generate
	git diff --exit-code -- generated/ deploy/

validate: lint coverage generated-check

all: validate

e2e-kind:
	./test/e2e/kind/run.sh
