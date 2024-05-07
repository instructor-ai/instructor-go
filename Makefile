.DEFAULT_GOAL := instructor

GO_SOURCES := \
	$(wildcard *.go) \
	$(wildcard */*.go)

instructor: $(GO_SOURCES)
	go build -o instructor

.PHONY: clean
clean:
	rm -f instructor

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run
