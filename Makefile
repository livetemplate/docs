.PHONY: help serve run-firstapp validate

help:
	@echo "LiveTemplate docs — local development targets:"
	@echo "  serve         Run the docs site (tinkerdown serve content/)"
	@echo "  run-firstapp  Run the tutorial counter locally on :9090 (use when"
	@echo "                iterating on _app/counter/; the docs site's embeds"
	@echo "                point at lt-firstapp.fly.dev by default)"
	@echo "  validate      Validate every page (tinkerdown validate content/)"

serve:
	tinkerdown serve content/

run-firstapp:
	cd content/getting-started/_app/counter && PORT=9090 go run .

validate:
	tinkerdown validate content/
