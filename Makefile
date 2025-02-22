ALL: generate test install
PHONY: test install buf-generate

PROTO_FILES=$(shell find internal/converter/testdata -type f -name '*.proto')

generate: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	go generate ./...
	go generate ./internal/converter/testdata

test: generate
	go test -coverprofile=coverage.out -coverpkg=./internal/...,./converter/... ./...
	# To see coverage report:
	# go tool cover -html=coverage.out

install:
	go install

buf-generate: install
	buf generate --path internal/
