.PHONY: test-unit
test-unit:
	@go test -v -cover -run "^.*Unit" ./...

.PHONY: test-integration
test-integration: # run inside container
	@go test -v -cover -run "^.*Integration" ./...

.PHONY: test-cover
test-cover: # run inside container
	@go test -v -cover -coverprofile cover.out ./...
	@go tool cover -html cover.out -o cover.html
	@rm cover.out
	@firefox cover.html
