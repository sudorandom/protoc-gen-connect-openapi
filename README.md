# protoc-gen-connect-openapi
Generate OpenAPI v3.1 from protobuf matching the [Connect protocol](https://connectrpc.com/docs/protocol).

With this OpenAPI file, you can:

- Generate Documentation (Elements, redoc, etc.)
- Generate HTTP Clients for places where you cannot use gRPC (openapi-generator)

```mermaid
flowchart LR

protobuf(Protobuf) -->|protoc-gen-connect-openapi| openapi(OpenAPI)
openapi -->|elements| elements(Gorgeous\nAPI Documentation)
openapi -->|openapi-generator| other-languages(Languages that\nConnect doesn't\n support yet)
openapi -->|?| ???(?)
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
```
go install github.com/sudorandom/protoc-gen-connect-openapi@main
```

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
    --connect-openapi_opt=base=example.base.yaml,without-streaming
```

See `protoc --help` for more protoc options.

### Using buf
With buf you can make a `buf.gen.yaml` with your options, like this:
```
version: v1
plugins:
  - plugin: connect-openapi
    out: out
    opt:
    - base=example.base.yaml
```
And then run `buf generate`. See [the documentation on buf generate](https://buf.build/docs/reference/cli/buf/generate#usage) for more help.

### Gnostic Support
protoc-gen-connect-openapi also has support for the [OpenAPI v3 annotations](https://github.com/google/gnostic/blob/main/openapiv3/annotations.proto) provided by the [google/gnostic project](https://github.com/google/gnostic). Here's an example of what this looks like in a protobuf file:

```protobuf
syntax = "proto3";

package example_with_gnostic;

import "gnostic/openapi/v3/annotations.proto";

option (gnostic.openapi.v3.document) = {
  info: {
    title: "Title from annotation";
    version: "Version from annotation";
    description: "Description from annotation";
    contact: {
      name: "Contact Name";
      url: "https://github.com/sudorandom/protoc-gen-connect-openapi";
      email: "hello@sudorandom.com";
    }
    license: {
      name: "MIT License";
      url: "https://github.com/sudorandom/protoc-gen-connect-openapi/blob/master/LICENSE";
    }
  }
  components: {
    security_schemes: {
      additional_properties: [
        {
          name: "BasicAuth";
          value: {
            security_scheme: {
              type: "http";
              scheme: "basic";
            }
          }
        }
      ]
    }
  }
};

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello(HelloRequest) returns (HelloReply) {
    option idempotency_level = NO_SIDE_EFFECTS;
    option (gnostic.openapi.v3.operation) = {
      deprecated: true,
      security: [
        {
          additional_properties: [
            {
              name: "BasicAuth";
              value: {
                value: []
              }
            }
          ]
        }
      ]
    };
  }
}

// The request message containing the user's name.
message HelloRequest {
  option (gnostic.openapi.v3.schema) = {title: "Custom title for a message"};

  string name = 1 [(gnostic.openapi.v3.property) = {title: "Custom title for a field"}];
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}

```
#### File Options
| Option | Supported? | Notes |
|---|---|---|
| gnostic.openapi.v3.document.openapi | ✅ | |
| gnostic.openapi.v3.document.info | ✅ | |
| gnostic.openapi.v3.document.servers | ✅ | |
| gnostic.openapi.v3.document.paths | ✅ | |
| gnostic.openapi.v3.document.components | 🟧 | Only security_schemes |
| gnostic.openapi.v3.document.security | ✅ | |
| gnostic.openapi.v3.document.tags | ✅ | |
| gnostic.openapi.v3.document.external_docs | ✅ | |
| gnostic.openapi.v3.document.specification_extension | ❌ | |

#### Method Options
| Option | Supported? |
|---|---|
| gnostic.openapi.v3.schema.tags | ✅ |
| gnostic.openapi.v3.schema.summary | ✅ |
| gnostic.openapi.v3.schema.description | ✅ |
| gnostic.openapi.v3.schema.external_docs | ✅ |
| gnostic.openapi.v3.schema.operation_id | ✅ |
| gnostic.openapi.v3.schema.parameters | ❌ |
| gnostic.openapi.v3.schema.request_body | ❌ |
| gnostic.openapi.v3.schema.responses | ❌ |
| gnostic.openapi.v3.schema.callbacks | ❌ |
| gnostic.openapi.v3.schema.deprecated  | ✅ |
| gnostic.openapi.v3.schema.security  | ✅ |
| gnostic.openapi.v3.schema.servers  | ✅ |
| gnostic.openapi.v3.schema.specification_extension | ❌ |

#### Message Options
| Option | Supported? |
|---|---|
| gnostic.openapi.v3.schema.nullable | ✅ |
| gnostic.openapi.v3.schema.discriminator | ❌ |
| gnostic.openapi.v3.schema.read_only | ✅ |
| gnostic.openapi.v3.schema.write_only | ✅ |
| gnostic.openapi.v3.schema.xml | ❌ |
| gnostic.openapi.v3.schema.external_docs | ✅ |
| gnostic.openapi.v3.schema.example | ✅ |
| gnostic.openapi.v3.schema.deprecated | ✅ |
| gnostic.openapi.v3.schema.title | ✅ |
| gnostic.openapi.v3.schema.multiple_of | ✅ |
| gnostic.openapi.v3.schema.maximum | ✅ |
| gnostic.openapi.v3.schema.exclusive_maximum | ✅ |
| gnostic.openapi.v3.schema.minimum | ✅ |
| gnostic.openapi.v3.schema.exclusive_minimum | ✅ |
| gnostic.openapi.v3.schema.max_length | ✅ |
| gnostic.openapi.v3.schema.min_length | ✅ |
| gnostic.openapi.v3.schema.pattern | ✅ |
| gnostic.openapi.v3.schema.max_items | ✅ |
| gnostic.openapi.v3.schema.min_items | ✅ |
| gnostic.openapi.v3.schema.unique_items | ✅ |
| gnostic.openapi.v3.schema.max_properties | ✅ |
| gnostic.openapi.v3.schema.min_properties | ✅ |
| gnostic.openapi.v3.schema.string required | ✅ |
| gnostic.openapi.v3.schema.Any enum | ✅ |
| gnostic.openapi.v3.schema.type | ✅ |
| gnostic.openapi.v3.schema.SchemaOrReference all_of | ❌ |
| gnostic.openapi.v3.schema.SchemaOrReference one_of | ❌ |
| gnostic.openapi.v3.schema.SchemaOrReference any_of | ❌ |
| gnostic.openapi.v3.schema.not | ❌ |
| gnostic.openapi.v3.schema.items | ❌ |
| gnostic.openapi.v3.schema.properties | ❌ |
| gnostic.openapi.v3.schema.additional_properties | ❌ |
| gnostic.openapi.v3.schema.default | ✅ |
| gnostic.openapi.v3.schema.description | ✅ |
| gnostic.openapi.v3.schema.format | ✅ |
| gnostic.openapi.v3.schema.NamedAny specification_extension | ❌ |

#### Field Options
| Option | Supported? |
|---|---|
| gnostic.openapi.v3.property.nullable | ✅ |
| gnostic.openapi.v3.property.discriminator | ❌ |
| gnostic.openapi.v3.property.read_only | ✅ |
| gnostic.openapi.v3.property.write_only | ✅ |
| gnostic.openapi.v3.property.xml | ❌ |
| gnostic.openapi.v3.property.external_docs | ✅ |
| gnostic.openapi.v3.property.example | ✅ |
| gnostic.openapi.v3.property.deprecated | ✅ |
| gnostic.openapi.v3.property.title | ✅ |
| gnostic.openapi.v3.property.multiple_of | ✅ |
| gnostic.openapi.v3.property.maximum | ✅ |
| gnostic.openapi.v3.property.exclusive_maximum | ✅ |
| gnostic.openapi.v3.property.minimum | ✅ |
| gnostic.openapi.v3.property.exclusive_minimum | ✅ |
| gnostic.openapi.v3.property.max_length | ✅ |
| gnostic.openapi.v3.property.min_length | ✅ |
| gnostic.openapi.v3.property.pattern | ✅ |
| gnostic.openapi.v3.property.max_items | ✅ |
| gnostic.openapi.v3.property.min_items | ✅ |
| gnostic.openapi.v3.property.unique_items | ✅ |
| gnostic.openapi.v3.property.max_properties | ✅ |
| gnostic.openapi.v3.property.min_properties | ✅ |
| gnostic.openapi.v3.property.string required | ✅ |
| gnostic.openapi.v3.property.Any enum | ✅ |
| gnostic.openapi.v3.property.type | ✅ |
| gnostic.openapi.v3.property.SchemaOrReference all_of | ❌ |
| gnostic.openapi.v3.property.SchemaOrReference one_of | ❌ |
| gnostic.openapi.v3.property.SchemaOrReference any_of | ❌ |
| gnostic.openapi.v3.property.not | ❌ |
| gnostic.openapi.v3.property.items | ❌ |
| gnostic.openapi.v3.property.properties | ❌ |
| gnostic.openapi.v3.property.additional_properties | ❌ |
| gnostic.openapi.v3.property.default | ✅ |
| gnostic.openapi.v3.property.description | ✅ |
| gnostic.openapi.v3.property.format | ✅ |
| gnostic.openapi.v3.property.NamedAny specification_extension | ❌ |

For more information on how to use each option in your Protobuf file, you can reference [the gnostic.openapi.v3 module documentation](https://buf.build/gnostic/gnostic/docs/main:gnostic.openapi.v3) and the [google/gnostic repo](https://github.com/google/gnostic). Note that this is a new feature, so if find something that isn't supported that you need, please [create an issue](https://github.com/sudorandom/protoc-gen-connect-openapi/issues/new).

## Options
| Option | Values | Description |
|---|---|---|
| path | `{filepath}` | Output filepath, defaults to per-protofile output if not given. |
| format | `yaml` or `json` | Which format to use for the OpenAPI file, defaults to `yaml`. |
| base | `{filepath}` | The path to a base OpenAPI file to populate fienlds that this tool doesn't populate. |
| with-streaming | - | Generate OpenAPI with content types related to streaming (can be messy). |
| only-string-enum-values | - | Only use strings for enum values, defaults to showing integers and strings |
| debug | - | Emit debug logs |
