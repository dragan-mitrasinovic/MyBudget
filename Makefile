# MyBudget quality gates — single source of truth for fmt/lint/vet/test. See ADR-0008.
# Every other layer (editor, agent hook, lefthook pre-commit) calls these targets.
.PHONY: help fmt fmt-check lint vet test check dev-api dev-web

help:
	@echo "make fmt        format web (Prettier) + server (gofmt)"
	@echo "make fmt-check  verify formatting without writing"
	@echo "make lint       angular-eslint + stylelint + golangci-lint"
	@echo "make vet        go vet"
	@echo "make test       web (Vitest) + go test"
	@echo "make check      fmt-check + lint (run before committing)"
	@echo ""
	@echo "Local dev (not part of the gate):"
	@echo "make dev-api    run the Go API with real Google login + the dev dummy-login route (DEV_AUTH=true)"
	@echo "make dev-web    run the Angular dev server (proxies /api to the API)"

fmt:
	cd web && npm run format
	cd server && gofmt -w .

fmt-check:
	cd web && npm run format:check
	cd server && test -z "$$(gofmt -l .)" || { echo "gofmt: run 'make fmt' — files need formatting:"; gofmt -l .; exit 1; }

lint:
	cd web && npm run lint
	cd web && npm run lint:css
	cd server && golangci-lint run

vet:
	cd server && go vet ./...

test:
	cd web && npm test
	cd server && go test ./...

# Fast pre-commit-style gate. Tests are intentionally excluded (too slow / can watch); run
# `make test` separately.
check: fmt-check lint

# ---- Local development (not part of the quality gate) ----
# Run the two processes in separate terminals, then open http://localhost:4200.
# dev-api uses the real Google client ID (matching the web app's dev config) so real
# "Sign in with Google" works locally, and *also* keeps the dev-only dummy-login route
# enabled (DEV_AUTH=true) so agents/automation can sign in headlessly via /api/dev/login.
dev-api:
	cd server && GOOGLE_CLIENT_ID=923945549782-d0tr117jt1lo0i4u2r5aq783flp7uo3g.apps.googleusercontent.com DEV_AUTH=true go run ./cmd/mybudget

dev-web:
	cd web && npm start
