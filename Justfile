specs_find := "find internal/converter/testdata -path '*/output/*.openapi.yaml' ! -name '*novalidate*' ! -path '*/disable_default_response/*'"

# The first recipe is the default, mirroring the 'ALL' target.
default: generate test install

# Generate fixture descriptor set and apply license headers.
# In `just`, there isn't a direct equivalent to Make's file-based prerequisites.
# The `go generate` commands inherently use the necessary source files.
generate:
    @echo "Generating fixture descriptor set"
    go generate ./...
    go generate ./internal/converter/testdata
    license-header \
        --license-type apache \
        --copyright-holder "The Connect Authors" \
        --year-range "2023-2026" \
        --ignore "demo/node_modules" \
        --ignore "demo/out" \
        --ignore "internal/converter/testdata/standard/protovalidate\.(numbers|strings)\.proto"

# Run tests after generating files.
test: generate
    go test -coverprofile=coverage.out -coverpkg=./internal/...,./converter/... ./...
    # To see coverage report:
    # go tool cover -html=coverage.out

# Verify that `just generate` and tests don't produce any uncommitted diffs (including license headers).
checkgenerate:
    test -z "$(git status --porcelain | tee /dev/stderr)"

lint:
    golangci-lint run ./...

# Install the application binary.
install:
    go install

# Build the application binary locally.
build:
    go build -v ./...
    go build -o protoc-gen-connect-openapi .

# Run buf to generate code from protobuf definitions.
buf-generate: install
    buf generate --path internal/proto

# Generate specs for a proto file with this plugin, run them through
# downstream tooling (documentation sites, type generators), and serve the
# results locally. Also accepts an already-generated OpenAPI/JSON Schema file.
# Example: just demo internal/converter/testdata/standard/helloworld.proto
demo input port="4173":
    just --justfile demo/Justfile demo "{{ absolute_path(input) }}" {{ port }}

# Run goreleaser to create a release or check configuration.
release *args="release --clean":
    goreleaser {{ args }}

clear-golden:
    rm -rf internal/converter/testdata/*/output

# Run vacuum OpenAPI linter on all golden specs.
# Usage:
#   just vacuum                    # lint all golden specs
#   just vacuum -e                 # errors only
vacuum *args:
    vacuum lint {{ args }} $({{ specs_find }})

# Show only OpenAPI errors across specs (suppresses verbose tables and passing specs).
vacuum-errors:
    @vacuum lint --github-annotations $({{ specs_find }}) 2>&1 | grep '^::error' || echo "All OpenAPI specs passed with 0 errors."

# List only filenames that have OpenAPI errors.
vacuum-failing-files:
    @vacuum lint --github-annotations $({{ specs_find }}) 2>&1 | sed -n 's/^::error.*file=\([^,]*\).*/\1/p' | sort -u

vacuum-dashboard spec="internal/converter/testdata/standard/output/helloworld.openapi.yaml":
    vacuum dashboard {{ spec }}
