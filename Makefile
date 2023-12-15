ALL: fixtures/fileset.binpb
PHONY: test install buf-generate

PROTO_FILES=$(shell find fixtures -type f -name '*.proto')

fixtures/fileset.binpb: $(PROTO_FILES)
	@echo "Generating fixture descriptor set"
	@cd fixtures; buf build -o fileset.binpb

fixtures/googleapis.binpb:
	@echo "Generating googleapis descriptor set"
	@cd fixtures; buf build buf.build/googleapis/googleapis -o googleapis.binpb

test: fixtures/fileset.binpb fixtures/googleapis.binpb
	go test ./...

install:
	go install

buf-generate: install
	cd fixtures; buf generate