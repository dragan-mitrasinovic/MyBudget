# MyBudget quality gates — single source of truth for fmt/lint/vet/test. See ADR-0008.
# Every other layer (editor, agent hook, lefthook pre-commit) calls these targets.
.PHONY: help fmt fmt-check lint vet test check

help:
	@echo "make fmt        format web (Prettier) + server (gofmt)"
	@echo "make fmt-check  verify formatting without writing"
	@echo "make lint       angular-eslint + golangci-lint"
	@echo "make vet        go vet"
	@echo "make test       web (Vitest) + go test"
	@echo "make check      fmt-check + lint (run before committing)"

fmt:
	cd web && npm run format
	cd server && gofmt -w .

fmt-check:
	cd web && npm run format:check
	cd server && test -z "$$(gofmt -l .)" || { echo "gofmt: run 'make fmt' — files need formatting:"; gofmt -l .; exit 1; }

lint:
	cd web && npm run lint
	cd server && golangci-lint run

vet:
	cd server && go vet ./...

test:
	cd web && npm test
	cd server && go test ./...

# Fast pre-commit-style gate. Tests are intentionally excluded (too slow / can watch); run
# `make test` separately.
check: fmt-check lint
