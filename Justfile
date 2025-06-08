# The first recipe is the default, mirroring the 'ALL' target.
default: generate test install

# Generate fixture descriptor set.
# In `just`, there isn't a direct equivalent to Make's file-based prerequisites.
# The `go generate` commands inherently use the necessary source files.
generate:
    @echo "Generating fixture descriptor set"
    go generate ./...
    go generate ./internal/converter/testdata

# Run tests after generating files.
test: generate
    go test -coverprofile=coverage.out -coverpkg=./internal/...,./converter/... ./...
    # To see coverage report:
    # go tool cover -html=coverage.out

# Install the application binary.
install:
    go install

# Run buf to generate code from protobuf definitions.
buf-generate: install
    buf generate --path internal/