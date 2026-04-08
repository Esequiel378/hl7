.PHONY: install publish

install:
	go install ./cmd/hl7

publish:
	@latest=$$(git describe --tags --match "v[0-9]*" --abbrev=0 2>/dev/null || echo "v0.0.0"); \
	version=$${latest#v}; \
	major=$$(echo $$version | cut -d. -f1); \
	minor=$$(echo $$version | cut -d. -f2); \
	patch=$$(echo $$version | cut -d. -f3); \
	next="v$$major.$$minor.$$((patch + 1))"; \
	echo "Tagging $$next and pushing..."; \
	git tag $$next && git push origin $$next
