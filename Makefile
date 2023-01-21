# Ensure that all dependencies are installed using vendored sources (/vendor).
export GOFLAGS=-mod=vendor

# Default to the system 'go'.
GO?=$(shell which go)

.PHONY: benchmarks
benchmarks: ## Running file diff benchmarks
	$(GO) test -bench . -run notest

.PHONY: tests
tests: ## Run unit tests
	$(GO) test -v  ./...

.PHONY: cover
cover: ## Calculate coverage
	@go test -mod=vendor -coverprofile=coverage.out -tags=integration ./... ; \
	cat coverage.out | \
	awk 'BEGIN {cov=0; stat=0;} $$3!="" { cov+=($$3==1?$$2:0); stat+=$$2; } \
	END {printf("Total coverage: %.2f%% of statements\n", (cov/stat)*100);}'
	@go tool cover -html=coverage.out

.PHONY: tidy
tidy: ## Tidy go modules and re-vendor
	@go mod tidy
	@go mod vendor
