base_dir = $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
TMPDIR ?= /tmp

client_cert = $(TMPDIR)/certificate.pem
client_key = $(TMPDIR)/private-key.pem
ca_cert = $(TMPDIR)/ca.pem

.PHONY: unit-test
unit-test:
	@echo "Run unit test......"
	go test -v -coverpkg=./pkg/... -coverprofile=coverage.out ./pkg/...

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

.PHONEY: fabric-up
fabric-up:
	"$(base_dir)/scripts/microfab.sh" down
	"$(base_dir)/scripts/microfab.sh" up "$(base_dir)/test/microfab/two-org.json"

.PHONEY: fabric-down
fabric-down:
	"$(base_dir)/scripts/microfab.sh" down

.PHONEY: integration-test
integration-test: fabric-up
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .cert' | sed -e 's/"//g' | base64 --decode > "$(client_cert)"
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .private_key' | sed -e 's/"//g' | base64 --decode > "$(client_key)"
	curl http://console.127.0.0.1.nip.io:8080/ak/api/v1/components | jq '.[] | select(.type == "identity") | select(.id == "org1admin") | .ca' | sed -e 's/"//g' | base64 --decode > "$(ca_cert)"
	CHAINCODE_PACKAGE="$(base_dir)/test/chaincode/basic.tar.gz" CLIENT_CERT="$(client_cert)" CLIENT_KEY="$(client_key)" CA_CERT="$(ca_cert)" ENDPOINT="org1peer-api.127-0-0-1.nip.io:8080" go test -v  "$(base_dir)/test/integration/..."
