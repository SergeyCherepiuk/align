.PHONY: build
build:
	@go build -o build/align main.go

.PHONY: run
run: # TODO: sc: Figure out how to avoid use of `sudo`.
	@sudo build/align | jq -c

.PHONY: test-unit
test-unit:
	@go test -v -cover -run "^.*Unit" ./...

.PHONY: test-integration
test-integration: # TODO: sc: Run inside container.
	@go test -v -cover -run "^.*Integration" ./...

.PHONY: test-cover
test-cover: # TODO: sc: Run inside container.
	@go test -v -cover -coverprofile cover.out ./...
	@go tool cover -html cover.out -o cover.html
	@rm cover.out
	@firefox cover.html
