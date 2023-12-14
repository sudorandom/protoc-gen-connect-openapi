ALL: fixtures/fileset.pb
PHONY: test

PROTO_FILES=$(shell find fixtures -type f -name '*.proto')

fixtures/fileset.pb: $(PROTO_FILES)
	@echo "Generating fixtures"
	protoc --descriptor_set_out=fixtures/fileset.pb --include_imports --include_source_info -I. fixtures/*.proto

test: fixtures/fileset.pb
	go test ./...
