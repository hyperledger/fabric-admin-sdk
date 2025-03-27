base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node

go_bin_dir := $(shell go env GOPATH)/bin

.PHONY: unit-test
unit-test: unit-test-go unit-test-node

.PHONY: unit-test-go
unit-test-go:
	cd '$(base_dir)' && \
		go test -race -coverprofile='$(base_dir)/coverage.out' '$(go_dir)/...'

.PHONY: unit-test-node
unit-test-node:
	cd '$(node_dir)/admin' && \
		npm install

.PHONY: lint
lint: golangci-lint

.PHONY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go_bin_dir)

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONY: scan
scan: scan-go scan-node

.PHONY: scan-go
scan-go: scan-go-govulncheck

.PHONY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck '$(base_dir)/...'

.PHONY: scan-node
scan-node: scan-node-npm-audit

.PHONY: scan-node-npm-audit
scan-node-npm-audit:
	cd "$(node_dir)/admin" && \
		npm install --package-lock-only && \
		npm audit --omit=dev

.PHONY: escapes_detect
escapes_detect:
	@go build -gcflags="-m -l" ./... 2>&1 | grep "escapes to heap" || true

.PHONY: generate
generate:
	go install go.uber.org/mock/mockgen@latest
	go generate ./pkg/...
