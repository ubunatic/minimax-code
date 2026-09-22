.PHONY: ⚙️ 🤖 install
⚙️:
🤖:

install:  # install code-index CLI using go install
	@cp index.csv cmd/code-index/cmd/data.csv
	@cd cmd/code-index && go install -ldflags="-s -w" ./main.go
	@mv $$(go env GOBIN)/main $$(go env GOBIN)/code-index 2>/dev/null || mv $$(go env GOPATH)/bin/main $$(go env GOPATH)/bin/code-index
	@echo "✓ Installed code-index to $$(go env GOPATH)/bin/code-index"
	@echo "  Add to PATH: export PATH=\$$HOME/go/bin:\$$PATH"
	@echo "  Try: code-index --help"

test-q1: 🤖  # run tests under Quota-1 enforcement
	harnez exec --quota-1 -- $(MAKE) test
