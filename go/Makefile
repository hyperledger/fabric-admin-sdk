.PHONY: unit-test
unit-test:
	@echo "Run unit test......"
	go test -coverpkg=./pkg/... -coverprofile=coverage.out ./pkg/...

include gotools.mk

.PHONY: basic-checks
basic-checks: gotools-install generate fmt

.PHONY: fmt
fmt:
	@echo "LINT: Running code checks......"
	./scripts/gofmt.sh

golint:
	./scripts/golinter.sh

set_govulncheck:
	@go install golang.org/x/vuln/cmd/govulncheck@latest

govulncheck: set_govulncheck
	@govulncheck -v ./... || true

escapes_detect:
	@go build -gcflags="-m -l" ./... 2>&1 | grep "escapes to heap" || true

.PHONEY: generate
generate:
	go install github.com/golang/mock/mockgen@v1.6
	go generate ./pkg/...
