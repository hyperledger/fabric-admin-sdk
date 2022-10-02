.PHONY: unit-test
unit-test:
	@echo "Run unit test......"
	go test -v ./... --bench=. -cover --count=1

include gotools.mk

.PHONY: basic-checks
basic-checks: gotools-install linter

.PHONY: linter
linter:
	@echo "LINT: Running code checks......"
	./scripts/golinter.sh

dummy:
	./scripts/golinter.sh

set_govulncheck:
	@go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: set_govulncheck
	@govulncheck -v ./... || true

escapes_detect:
	@go build -gcflags="-m -l" ./... | grep "escapes to heap" || true
