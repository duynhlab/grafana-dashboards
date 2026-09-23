.PHONY: generate test coverage fmt lint generated-check validate e2e-kind all

generate:
	go run ./cmd/generate

test:
	go test ./...

# Every test binary built with -coverpkg=./... reports a block for every package
# in the module, so the raw profile carries one entry per (block, test binary)
# and go tool cover counts each of them separately. A block covered by one
# package's tests and untouched by four others reads as 20% rather than covered,
# and the dilution changes with the test cache, so the same commit measured
# 94.6% warm and 69.3% cold. mergeprofile keeps the highest count per block,
# which is what "did any test execute this line" means.
COVERPROFILE := /tmp/grafana-dashboards-coverage.out

coverage:
	go test ./... -coverpkg=./... -coverprofile=$(COVERPROFILE).raw
	@head -n 1 $(COVERPROFILE).raw > $(COVERPROFILE)
	@awk 'NR == 1 { next } { if ($$3 + 0 > count[$$1] + 0) count[$$1] = $$3; stmts[$$1] = $$2 } \
		END { for (block in stmts) print block, stmts[block], count[block] + 0 }' \
		$(COVERPROFILE).raw | sort >> $(COVERPROFILE)
	@total="$$(go tool cover -func=$(COVERPROFILE) | awk '/^total:/ {gsub("%", "", $$3); print $$3}')"; \
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
