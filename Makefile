.PHONY: ⚙️ 🤖
⚙️:
🤖:

install: ⚙️  # install code-index CLI and man pages
	@cp index.csv cmd/code-index/cmd/data.csv
	@cd cmd/code-index && go install -ldflags="-s -w" ./main.go
	@mv $$(go env GOBIN)/main $$(go env GOBIN)/code-index 2>/dev/null || mv $$(go env GOPATH)/bin/main $$(go env GOPATH)/bin/code-index
	@code-index man --install
	@echo "✓ Installed code-index to $$(go env GOPATH)/bin/code-index"
	@echo "  Add to PATH: export PATH=\$$HOME/go/bin:\$$PATH"
	@echo "  Try: code-index --help or man code-index"

install-mcode: ⚙️  # build and install the local mcode CLI globally
	@podman build --tag minimax-code:local --file Containerfile .
	@echo "✓ Built minimax-code:local; run it with: make container-run"

container-build: ⚙️  # build the application image without using host Node tooling
	podman build --tag minimax-code:local --file Containerfile .

container-run: ⚙️  # run mcode in the application container
	podman run --rm --interactive --tty --env MCODE_DISABLE_TELEMETRY=1 --env DO_NOT_TRACK=1 minimax-code:local

make-q1: 🤖 container-build  # run the verification gate inside the Container
	podman run --rm --network=none --env CI=1 --env MCODE_DISABLE_TELEMETRY=1 --env DO_NOT_TRACK=1 --entrypoint pnpm minimax-code:local verify

man: ⚙️  # generate roff man page to stdout
	@code-index man

test-q1: 🤖  # compatibility alias for the Container verification gate
	@$(MAKE) make-q1
