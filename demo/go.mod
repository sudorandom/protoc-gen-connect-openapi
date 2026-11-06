// The demo builds generated Go files (quicktype output) under out/. This
// separate module keeps them out of the plugin's `go build ./...`, `go vet`,
// and lint runs.
module github.com/sudorandom/protoc-gen-connect-openapi/demo

go 1.26.0
