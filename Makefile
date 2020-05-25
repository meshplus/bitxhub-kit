
help: Makefile
	@printf "${BLUE}Choose a command run:${NC}\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/    /'

TEST_PKGS := $(shell go list ./...)

GO = go

## make test: Run go unittest
test:
	@$(GO) test ${TEST_PKGS} -count=1

## make linter: Run golanci-lint
linter:
	golangci-lint run -E goimports -E bodyclose --skip-dirs-use-default
