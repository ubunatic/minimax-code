.PHONY: ⚙️ 🤖 install man
⚙️:
🤖:

install:  # install code-index CLI and man pages
	@cp index.csv cmd/code-index/cmd/data.csv
	@cd cmd/code-index && go install -ldflags="-s -w" ./main.go
	@mv $$(go env GOBIN)/main $$(go env GOBIN)/code-index 2>/dev/null || mv $$(go env GOPATH)/bin/main $$(go env GOPATH)/bin/code-index
	@code-index man --install
	@echo "✓ Installed code-index to $$(go env GOPATH)/bin/code-index"
	@echo "  Add to PATH: export PATH=\$$HOME/go/bin:\$$PATH"
	@echo "  Try: code-index --help or man code-index"

man:  # generate roff man page to stdout
	@code-index man

test-q1: 🤖  # run tests under Quota-1 enforcement
	harnez exec --quota-1 -- $(MAKE) test
