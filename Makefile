.PHONY: help serve site validate test test-unit test-examples test-e2e test-e2e-local examples-test examples-list

help:
	@echo "LiveTemplate docs — local development targets:"
	@echo "  serve            Run the docs site (tinkerdown serve content/)"
	@echo "  site             Run cmd/site (recipes binary on :9091, what tinkerdown proxies to)"
	@echo "  validate         Validate every page (tinkerdown validate content/)"
	@echo "  test-unit        Run fast Go package tests with GOWORK=off"
	@echo "  test-examples    Run example package tests with GOWORK=off"
	@echo "  test-e2e         Run browser-backed docs e2e tests with GOWORK=off"
	@echo "  test-e2e-local   Run docs e2e against http://127.0.0.1:8084"
	@echo "  test             Run the full Go test suite with GOWORK=off"
	@echo "  examples-test    Alias for test-examples"
	@echo "  examples-list    Enumerate examples (one per line)"
	@echo ""
	@echo "  To run a single example standalone:"
	@echo "      go run ./examples/<slug>/cmd --dev"
	@echo ""
	@echo "  To run a single example's tests:"
	@echo "      go test ./examples/<slug>"

serve:
	tinkerdown serve content/

site:
	go run ./cmd/site

validate:
	tinkerdown validate content/

test-unit:
	GOWORK=off go test ./cmd/... ./examples/counter/... ./examples/counter-basic/...

test-examples:
	GOWORK=off go test ./examples/...

test-e2e:
	GOWORK=off go test ./e2e/...

test-e2e-local:
	GOWORK=off E2E_BASE_URL=http://127.0.0.1:8084 go test ./e2e/...

test:
	GOWORK=off go test ./...

examples-test:
	$(MAKE) test-examples

examples-list:
	@ls -1 examples/ | grep -v '^cmd$$'
