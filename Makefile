ALL: internal/converter/testdata/fileset.binpb test install
PHONY: test install buf-generate

PROTO_FILES=$(shell find internal/converter/testdata -type f -name '*.proto')

internal/converter/testdata/fileset.binpb: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	buf build -o internal/converter/testdata/fileset.binpb

test: internal/converter/testdata/fileset.binpb
	go test -coverprofile=coverage.out -coverpkg=./internal/...,./converter/... ./...
	# To see coverage report:
	# go tool cover -html=coverage.out

install:
	go install

buf-generate: install
	buf generate --path internal/
