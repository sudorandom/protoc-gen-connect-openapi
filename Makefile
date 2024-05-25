ALL: internal/converter/fixtures/fileset.binpb test install
PHONY: test install buf-generate

PROTO_FILES=$(shell find internal/converter/fixtures -type f -name '*.proto')

internal/converter/fixtures/fileset.binpb: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	buf build -o internal/converter/fixtures/fileset.binpb

test: internal/converter/fixtures/fileset.binpb
	go test ./...

install:
	go install

buf-generate: install
	buf generate --path internal/
