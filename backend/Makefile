# GreenMetrics backend Makefile.
#
# Doctrine refs: Rule 14 (CLI contract), Rule 17 (DX), Rule 23 (tooling discipline).
# CLI contract documented in docs/CLI-CONTRACT.md.

.PHONY: help \
        run debug-backend build clean docker \
        test test-unit test-integration test-e2e test-property test-security test-conformance test-static \
        bench cover lint vet fmt tidy precommit verify \
        migrate migrate-up migrate-down migrate-status migrate-create \
        openapi-validate openapi-bundle policy-check audit-rejections layers-doc

APP := greenmetrics-backend
PKG := ./...
BIN := bin/$(APP)
GO  ?= go

# DATABASE_URL is sourced from environment; Makefile never sets defaults for secrets.
DB ?= $(DATABASE_URL)

help:  ## Print available targets and short description.
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2}'

# --- run / build -----------------------------------------------------------

run:  ## Start backend server (local dev).
	$(GO) run ./cmd/server

debug-backend:  ## Run backend under dlv (delve debugger) on port 2345.
	dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient ./cmd/server

build:  ## Build static binary (CGO disabled, stripped).
	mkdir -p bin
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/server

clean:  ## Remove build artefacts.
	rm -rf bin coverage.out coverage.html

docker:  ## Build local Docker image.
	docker build -t $(APP):local .

# --- testing ---------------------------------------------------------------

test: test-unit  ## Alias: run unit tests.

test-unit:  ## Unit tests with race detector.
	$(GO) test -race -cover $(PKG)

test-integration:  ## Integration tests using testcontainers (Timescale 16).
	$(GO) test -race -tags=integration ./tests/integration/...

test-e2e:  ## Full-stack E2E tests via docker-compose.
	$(GO) test -tags=e2e ./tests/e2e/...

test-property:  ## Property-based tests via gopter.
	$(GO) test -tags=property ./tests/property/...

test-security:  ## RBAC / RLS / boot-refusal security tests.
	$(GO) test -tags=security ./tests/security/...

test-conformance:  ## RFC 7807 / RFC 3339 UTC / UUIDv4 / money / health envelope.
	$(GO) test -tags=conformance ./tests/conformance/...

test-static:  ## Static AST checks (no float for money, no panic on user input).
	$(GO) test -tags=static ./tests/static/...

bench:  ## Benchmarks (CopyFrom throughput, JWT verify, etc.).
	$(GO) test -bench=. -benchmem -run=^$$ ./tests/bench/...

cover:  ## Generate HTML coverage report.
	$(GO) test -race -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html

# --- lint / format ---------------------------------------------------------

lint:  ## Run golangci-lint (fast).
	golangci-lint run --fast

vet:  ## go vet ./...
	$(GO) vet $(PKG)

fmt:  ## gofmt + goimports.
	gofmt -s -w .
	goimports -w .

tidy:  ## go mod tidy.
	$(GO) mod tidy

precommit:  ## Run all pre-commit hooks against the entire repo.
	cd .. && pre-commit run --all-files

verify:  ## Full continuous-verification loop (lint + tests + security + policy).
	@$(MAKE) lint vet test-unit test-integration test-property test-security test-conformance test-static
	@$(MAKE) -C .. policy-check

# --- migrations (pressly/goose) -------------------------------------------

migrate: migrate-up  ## Alias: apply all pending migrations.

migrate-up:  ## Apply all pending migrations (forward-only).
	@if [ -z "$(DB)" ]; then echo "DATABASE_URL not set"; exit 1; fi
	@if command -v goose >/dev/null 2>&1; then \
		goose -dir migrations postgres "$(DB)" up; \
	else \
		echo "[goose] not installed; falling back to legacy psql shell-out"; \
		psql "$(DB)" -f migrations/0001_init.sql; \
		psql "$(DB)" -f migrations/0002_hypertables.sql; \
		psql "$(DB)" -f migrations/0003_continuous_aggregates.sql; \
		psql "$(DB)" -f migrations/0004_retention.sql; \
		psql "$(DB)" -f migrations/0005_emission_factors.sql; \
	fi

migrate-down:  ## Roll back N migrations (default 1). Usage: make migrate-down N=1
	@if [ -z "$(DB)" ]; then echo "DATABASE_URL not set"; exit 1; fi
	@N=$${N:-1}; \
	for i in $$(seq 1 $$N); do \
		goose -dir migrations postgres "$(DB)" down; \
	done

migrate-status:  ## Show migration status.
	@if [ -z "$(DB)" ]; then echo "DATABASE_URL not set"; exit 1; fi
	goose -dir migrations postgres "$(DB)" status

migrate-create:  ## Create a new migration. Usage: make migrate-create NAME=add_idempotency
	@if [ -z "$(NAME)" ]; then echo "NAME=<snake_case> required"; exit 1; fi
	goose -dir migrations create $(NAME) sql

# --- contracts -------------------------------------------------------------

openapi-validate:  ## Lint canonical OpenAPI spec.
	npx -y @redocly/cli@latest lint ../api/openapi/v1.yaml

openapi-bundle:  ## Bundle OpenAPI for distribution (single-file JSON).
	npx -y @redocly/cli@latest bundle ../api/openapi/v1.yaml -o ../api/openapi/v1.bundle.json

# --- policies + housekeeping ----------------------------------------------

policy-check:  ## Run conftest gates locally (audit mode).
	@cd .. && \
	if [ -d policies/conftest/k8s ]; then conftest test --policy policies/conftest/k8s/ k8s/ || true; fi && \
	if [ -d policies/conftest/dockerfile ]; then echo "(hadolint runs in pre-commit)"; fi

audit-rejections:  ## Scan ADRs and PRs for keywords from docs/adr/REJECTED.md.
	@cd .. && \
	keywords="service mesh|GraphQL gateway|Kafka for telemetry|microservice split|Snyk|Vault for app config|SPIRE|custom KMS|unpinned tag"; \
	echo "Scanning ADRs and PRs for rejected patterns..."; \
	grep -rEn "$$keywords" docs/adr/ docs/decisions/ docs/backend/decisions/ 2>/dev/null || echo "(no hits)"

layers-doc:  ## Regenerate Mermaid diagram in docs/LAYERS.md from docs/layers.yaml.
	@echo "TODO: render Mermaid from docs/layers.yaml; manual sync until templating script lands."
