#***** Generation
.PHONY: gen-proto
gen-proto:
	buf generate

.PHONY: gen-mocks
gen-mocks:
	go run github.com/vektra/mockery/v2@v2.52.1

.PHONY: generate
generate: gen-proto gen-mocks

#***** Build
.PHONY: build-calculator
build-calculator:
	CGO_ENABLED=0 go build -o ./bin/calculator ./cmd/calculator
.PHONY: build-agent
build-agent:
	CGO_ENABLED=0 go build -o ./bin/agent ./cmd/agent

#***** Docker
.PHONY: up
up:
	docker-compose up

#***** Lint
.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6 run ./...

#***** Tests
.PHONY: test
test:
	go test -v ./internal/...

.PHONY: test-cov
test-cov:
	mkdir -p coverage \
	&& go test ./internal/... -coverprofile=coverage/cover \
	&& go tool cover -html=coverage/cover -o coverage/cover.html
