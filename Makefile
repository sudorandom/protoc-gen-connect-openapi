ALL: fixtures/fileset.binpb
PHONY: test install buf-generate

PROTO_FILES=$(shell find fixtures -type f -name '*.proto')

fixtures/fileset.binpb: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	@cd fixtures; buf build -o fileset.binpb

test: fixtures/fileset.binpb
	go test -v ./...

install:
	go install

buf-generate: install
	cd fixtures; buf generate