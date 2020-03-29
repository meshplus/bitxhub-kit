
help: Makefile
	@printf "${BLUE}Choose a command run:${NC}\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/    /'

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports -E bodyclose --skip-dirs-use-default
