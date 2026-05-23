.PHONY: help serve site validate examples-test examples-list

help:
	@echo "LiveTemplate docs — local development targets:"
	@echo "  serve            Run the docs site (tinkerdown serve content/)"
	@echo "  site             Run cmd/site (recipes binary on :9091, what tinkerdown proxies to)"
	@echo "  validate         Validate every page (tinkerdown validate content/)"
	@echo "  examples-test    Run all example tests (go test ./examples/...)"
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

examples-test:
	go test ./examples/...

examples-list:
	@ls -1 examples/ | grep -v '^cmd$$'
