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