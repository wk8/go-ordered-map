.DEFAULT_GOAL := all

.PHONY: all
all: test lint fuzz

# the TEST_FLAGS env var can be set to eg run only specific tests
.PHONY: test
test:
	go test -v -count=1 -race -cover $(TEST_FLAGS)

.PHONY: bench
bench:
	go test -bench=.

.PHONY: fuzz
fuzz:
	go test -fuzz=. -fuzztime=10s ./...

.PHONY: lint
lint:
	golangci-lint run
