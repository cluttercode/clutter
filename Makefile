BINDIR?=bin

GO_BUILD_OPTS?=

GOTEST=go test

# Check if gotestsum is around. See https://github.com/gotestyourself/gotestsum if you want to use it.
ifneq (, $(shell which gotestsum))
GOTEST=gotestsum --format testname --
endif

# Test options. -count 1 disables test result caching.
GO_TEST_OPTS?=-v --race -count 1

.PHONY: all
all: clean shellcheck clutter lint test

.PHONY: clean
clean:
	rm -fR "$(BINDIR)/*"

.PHONY: bin
bin: clutter

.PHONY: clutter
clutter:
	go build -o $(BINDIR)/clutter $(GO_BUILD_OPTS) ./cmd/clutter

.PHONY: test
test: test-unit test-end-to-end

.PHONY: test-unit
test-unit:
	$(GOTEST) $(GO_TEST_OPTS) ./...

.PHONY: test-end-to-end
test-end-to-end: clutter
	./tests/cli/run.sh

.PHONY: lint
lint:
	docker run --rm -v "$(shell pwd):/code" -w /code golangci/golangci-lint:v1.26.0 golangci-lint run

.PHONY: shellcheck
shellcheck:
	find . -name \*.sh | xargs docker run --rm -v "$(shell pwd):/code" -w /code koalaman/shellcheck:stable -e SC2059

.PHONY: install
install:
	go install ./cmd/clutter
