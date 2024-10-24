# protoc-gen-connect-openapi
[![Go](https://github.com/sudorandom/protoc-gen-connect-openapi/actions/workflows/go.yml/badge.svg)](https://github.com/sudorandom/protoc-gen-connect-openapi/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/sudorandom/protoc-gen-connect-openapi)](https://goreportcard.com/report/github.com/sudorandom/protoc-gen-connect-openapi) [![Go Reference](https://pkg.go.dev/badge/github.com/sudorandom/protoc-gen-connect-openapi.svg)](https://pkg.go.dev/github.com/sudorandom/protoc-gen-connect-openapi)

Generate OpenAPI v3.1 from protobuf matching the [Connect protocol](https://connectrpc.com/docs/protocol). With these [OpenAPI](https://www.openapis.org/what-is-openapi), you can:

- Generate Documentation (Elements, redoc, etc.)
- Generate HTTP Clients for places where you cannot use gRPC (openapi-generator)
- Datasource for automated endpoint validation/security testing
- Datasource for monitoring dashboards
- Many other things

Features:
- Support for OpenAPIv3.1 (which has support for jsonschema)
- Support for many [protovalidate](https://github.com/bufbuild/protovalidate) options ([more info](protovalidate.md))
- Support for many [OpenAPIv3](https://github.com/google/gnostic/blob/main/openapiv3/annotations.proto) options from the [google/gnostic project](https://github.com/google/gnostic) protobufs ([more info](gnostic.md))
- Support for [gRPC-Gateway annotations](https://github.com/grpc-ecosystem/grpc-gateway) ([more info](grpcgateway.md))
- Has [an easy interface](https://pkg.go.dev/github.com/sudorandom/protoc-gen-connect-openapi/converter) for generating OpenAPI specs within the process

Example Pipeline:
- Protobuf file: [example](examples/basic.proto)
- OpenAPI file: [example](examples/basic.openapi.yaml)
- Generate documentation: [redocly example](examples/basic.png)

```mermaid
flowchart LR

protobuf(Protobuf) -->|protoc-gen-connect-openapi| openapi(OpenAPI)
openapi -->|elements| elements(API Documentation)
openapi -->|openapi-generator| other-languages(Other Language Support)
openapi -->|???| ???(Other Tooling!)
click elements "https://github.com/stoplightio/elements" _blank
click openapi-generator "https://github.com/OpenAPITools/openapi-generator" _blank
```

## Why?
[Connect](https://connectrpc.com/docs/introduction) makes your gRPC service look and feel like a normal HTTP/JSON API, at least for non-streaming RPC calls. It does this without an extra network hop and an extra proxy layer because the same Connect server can speak [the Connect, gRPC and gRPC-Web protocols in a single port](https://connectrpc.com/docs/multi-protocol).

This is what a GET request looks like. Note that GET requests are available for methods with an option of `idempotency_level=NO_SIDE_EFFECTS`.
```
> GET /connectrpc.greet.v1.GreetService/Greet?encoding=json&message=%7B%22name%22%3A%22Buf%22%7D HTTP/1.1
> Host: demo.connectrpc.com

< HTTP/1.1 200 OK
< Content-Type: application/json
<
< {"greeting": "Hello, Buf!"}
```
We can document this API as if it's a real JSON/HTTP API... because it is, and the gRPC "flavor" isn't so noticable due to Connect. With protoc-gen-connect-openapi you can declare your API using protobuf, serve it using gRPC and Connect and fully document it without the API consumers ever knowing what protobuf is or how to read it.

## Install
```shell
go install github.com/sudorandom/protoc-gen-connect-openapi@main
```

or you can download pre-built binaries from the [Github releases page](https://github.com/sudorandom/protoc-gen-connect-openapi/releases/latest).

## Usage
### With protoc
This tool works as a plugin for protoc. Here's a basic example:
```shell
protoc internal/converter/fixtures/helloworld.proto --connect-openapi_out=gen
```

With the JSON format:
```shell
protoc internal/converter/fixtures/helloworld.proto \
    --connect-openapi_out=gen \
--connect-openapi_opt=format=json
```

With a base OpenAPI file and without all of the streaming content type:
```shell
protoc internal/converter/fixtures/helloworld.proto \
    --connect-openapi_out=gen \
    --connect-openapi_opt=base=example.base.yaml,content-types=json
```

See `protoc --help` for more protoc options.

### Using buf
With buf you can make a `buf.gen.yaml` with your options, like this:
```yaml
version: v2
plugins:
  - local: protoc-gen-connect-openapi
    out: out
    opt:
    - base=example.base.yaml
```
And then run `buf generate`. See [the documentation on buf generate](https://buf.build/docs/reference/cli/buf/generate#usage) for more help.

### Proto Validate Support
protoc-gen-connect-openapi also has support for many [protovalidate](https://github.com/bufbuild/protovalidate) annotations. Note that not every protovalidate constraint translates clearly to OpenAPI.

[See the protovalidate documentation page for more information](protovalidate.md)

### gRPC-Gateway annotations
protoc-gen-connect-openapi also has support for the [gRPC-Gateway annotations](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/) provided by the [google/api/annotations.proto](https://github.com/googleapis/googleapis/blob/master/google/api/annotations.proto).

[See the gRPC-Gateway annotation documentation page for more information](grpcgateway.md)

### Gnostic Support
protoc-gen-connect-openapi also has support for the [OpenAPI v3 annotations](https://github.com/google/gnostic/blob/main/openapiv3/annotations.proto) provided by the [google/gnostic project](https://github.com/google/gnostic).

[See the gnostic documentation page for more information](gnostic.md)

## Options
| Option | Values | Description |
|---|---|---|
| allow-get | - | For methods that have `IdempotencyLevel=IDEMPOTENT`, this option will generate HTTP `GET` requests instead of `POST`. |
| base | `{filepath}` | The path to a base OpenAPI file to populate fields that this tool doesn't populate. |
| content-types | `json;proto` | Semicolon-separated content types to generate requests/repsonses |
| debug | - | Emit debug logs |
| format | `yaml` or `json` | Which format to use for the OpenAPI file, defaults to `yaml`. |
| include-number-enum-values | - | Include number enum values beside the string versions, defaults to only showing strings |
| path | `{filepath}` | Output filepath, defaults to per-protofile output if not given. |
| proto | - | Generate requests/repsonses with the protobuf content type |
| trim-unused-types | - | Remove types that aren't references from any method request or response. |
| with-proto-annotations | - | Add protobuf type annotations to the end of descriptions so users know the protobuf type that the field converts to. |
| with-proto-names | - | Use protobuf field names instead of the camelCase JSON names for property names. |
| with-streaming | - | Generate OpenAPI for client/server/bidirectional streaming RPCs (can be messy). |
