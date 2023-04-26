base_dir := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))

go_dir := $(base_dir)/pkg
node_dir := $(base_dir)/node

go_bin_dir := $(shell go env GOPATH)/bin

.PHONY: unit-test
unit-test: unit-test-go unit-test-node

.PHONY: unit-test-go
unit-test-go:
	cd '$(base_dir)' && \
		go test -coverprofile='$(base_dir)/coverage.out' '$(go_dir)/...'

.PHONY: unit-test-node
unit-test-node:
	cd '$(node_dir)/admin' && \
		npm install

.PHONEY: lint
lint: staticcheck golangci-lint

.PHONEY: staticcheck
staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck -f stylish '$(base_dir)/...'

.PHONEY: install-golangci-lint
install-golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go_bin_dir)

$(go_bin_dir)/golangci-lint:
	$(MAKE) install-golangci-lint

.PHONEY: golangci-lint
golangci-lint: $(go_bin_dir)/golangci-lint
	golangci-lint run

.PHONEY: scan
scan: scan-go scan-node

.PHONEY: scan-go
scan-go: scan-go-govulncheck scan-go-osv-scanner

.PHONEY: scan-go-govulncheck
scan-go-govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck '$(base_dir)/...'

.PHONEY: scan-go-osv-scanner
scan-go-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	osv-scanner --lockfile='$(base_dir)/go.mod'

.PHONEY: scan-node
scan-node: scan-node-npm-audit scan-node-osv-scanner

.PHONEY: scan-node-npm-audit
scan-node-npm-audit:
	cd "$(node_dir)/admin" && \
		npm install --package-lock-only && \
		npm audit --omit=dev

.PHONEY: scan-node-osv-scanner
scan-node-osv-scanner:
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	cd "$(node_dir)/admin" && \
		npm install --package-lock-only && \
		osv-scanner --lockfile=package-lock.json

.PHONEY: escapes_detect
escapes_detect:
	@go build -gcflags="-m -l" ./... 2>&1 | grep "escapes to heap" || true

.PHONEY: generate
generate:
	go install github.com/golang/mock/mockgen@v1.6
	go generate ./pkg/...
