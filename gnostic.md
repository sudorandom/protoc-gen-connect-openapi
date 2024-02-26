# Gnostic Support
protoc-gen-connect-openapi has support for the [OpenAPI v3 annotations](https://github.com/google/gnostic/blob/main/openapiv3/annotations.proto) provided by the [google/gnostic project](https://github.com/google/gnostic). Here's an example of what this looks like in a protobuf file:

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
| (gnostic.openapi.v3.document).openapi | ✅ | |
| (gnostic.openapi.v3.document).info | ✅ | |
| (gnostic.openapi.v3.document).servers | ✅ | |
| (gnostic.openapi.v3.document).paths | ✅ | |
| (gnostic.openapi.v3.document).components | 🟧 | Only security_schemes |
| (gnostic.openapi.v3.document).security | ✅ | |
| (gnostic.openapi.v3.document).tags | ✅ | |
| (gnostic.openapi.v3.document).external_docs | ✅ | |
| (gnostic.openapi.v3.document).specification_extension | ❌ | |

#### Method Options
| Option | Supported? |
|---|---|
| (gnostic.openapi.v3.schema).tags | ✅ |
| (gnostic.openapi.v3.schema).summary | ✅ |
| (gnostic.openapi.v3.schema).description | ✅ |
| (gnostic.openapi.v3.schema).external_docs | ✅ |
| (gnostic.openapi.v3.schema).operation_id | ✅ |
| (gnostic.openapi.v3.schema).parameters | ❌ |
| (gnostic.openapi.v3.schema).request_body | ❌ |
| (gnostic.openapi.v3.schema).responses | ❌ |
| (gnostic.openapi.v3.schema).callbacks | ❌ |
| (gnostic.openapi.v3.schema).deprecated  | ✅ |
| (gnostic.openapi.v3.schema).security  | ✅ |
| (gnostic.openapi.v3.schema).servers  | ✅ |
| (gnostic.openapi.v3.schema).specification_extension | ❌ |

#### Message Options
| Option | Supported? |
|---|---|
| (gnostic.openapi.v3.schema).nullable | ✅ |
| (gnostic.openapi.v3.schema).discriminator | ❌ |
| (gnostic.openapi.v3.schema).read_only | ✅ |
| (gnostic.openapi.v3.schema).write_only | ✅ |
| (gnostic.openapi.v3.schema).xml | ❌ |
| (gnostic.openapi.v3.schema).external_docs | ✅ |
| (gnostic.openapi.v3.schema).example | ✅ |
| (gnostic.openapi.v3.schema).deprecated | ✅ |
| (gnostic.openapi.v3.schema).title | ✅ |
| (gnostic.openapi.v3.schema).multiple_of | ✅ |
| (gnostic.openapi.v3.schema).maximum | ✅ |
| (gnostic.openapi.v3.schema).exclusive_maximum | ✅ |
| (gnostic.openapi.v3.schema).minimum | ✅ |
| (gnostic.openapi.v3.schema).exclusive_minimum | ✅ |
| (gnostic.openapi.v3.schema).max_length | ✅ |
| (gnostic.openapi.v3.schema).min_length | ✅ |
| (gnostic.openapi.v3.schema).pattern | ✅ |
| (gnostic.openapi.v3.schema).max_items | ✅ |
| (gnostic.openapi.v3.schema).min_items | ✅ |
| (gnostic.openapi.v3.schema).unique_items | ✅ |
| (gnostic.openapi.v3.schema).max_properties | ✅ |
| (gnostic.openapi.v3.schema).min_properties | ✅ |
| (gnostic.openapi.v3.schema).string required | ✅ |
| (gnostic.openapi.v3.schema).Any enum | ✅ |
| (gnostic.openapi.v3.schema).type | ✅ |
| (gnostic.openapi.v3.schema).all_of | ✅ |
| (gnostic.openapi.v3.schema).one_of | ✅ |
| (gnostic.openapi.v3.schema).any_of | ✅ |
| (gnostic.openapi.v3.schema).not | ✅ |
| (gnostic.openapi.v3.schema).items | ✅ |
| (gnostic.openapi.v3.schema).properties | ✅ |
| (gnostic.openapi.v3.schema).additional_properties | ✅ |
| (gnostic.openapi.v3.schema).default | ✅ |
| (gnostic.openapi.v3.schema).description | ✅ |
| (gnostic.openapi.v3.schema).format | ✅ |
| (gnostic.openapi.v3.schema).specification_extension | ❌ |

#### Field Options
| Option | Supported? |
|---|---|
| (gnostic.openapi.v3.property).nullable | ✅ |
| (gnostic.openapi.v3.property).discriminator | ❌ |
| (gnostic.openapi.v3.property).read_only | ✅ |
| (gnostic.openapi.v3.property).write_only | ✅ |
| (gnostic.openapi.v3.property).xml | ❌ |
| (gnostic.openapi.v3.property).external_docs | ✅ |
| (gnostic.openapi.v3.property).example | ✅ |
| (gnostic.openapi.v3.property).deprecated | ✅ |
| (gnostic.openapi.v3.property).title | ✅ |
| (gnostic.openapi.v3.property).multiple_of | ✅ |
| (gnostic.openapi.v3.property).maximum | ✅ |
| (gnostic.openapi.v3.property).exclusive_maximum | ✅ |
| (gnostic.openapi.v3.property).minimum | ✅ |
| (gnostic.openapi.v3.property).exclusive_minimum | ✅ |
| (gnostic.openapi.v3.property).max_length | ✅ |
| (gnostic.openapi.v3.property).min_length | ✅ |
| (gnostic.openapi.v3.property).pattern | ✅ |
| (gnostic.openapi.v3.property).max_items | ✅ |
| (gnostic.openapi.v3.property).min_items | ✅ |
| (gnostic.openapi.v3.property).unique_items | ✅ |
| (gnostic.openapi.v3.property).max_properties | ✅ |
| (gnostic.openapi.v3.property).min_properties | ✅ |
| (gnostic.openapi.v3.property).string required | ✅ |
| (gnostic.openapi.v3.property).Any enum | ✅ |
| (gnostic.openapi.v3.property).type | ✅ |
| (gnostic.openapi.v3.property).all_of | ✅ |
| (gnostic.openapi.v3.property).one_of | ✅ |
| (gnostic.openapi.v3.property).any_of | ✅ |
| (gnostic.openapi.v3.property).not | ✅ |
| (gnostic.openapi.v3.property).items | ✅ |
| (gnostic.openapi.v3.property).properties | ✅ |
| (gnostic.openapi.v3.property).additional_properties | ✅ |
| (gnostic.openapi.v3.property).default | ✅ |
| (gnostic.openapi.v3.property).description | ✅ |
| (gnostic.openapi.v3.property).format | ✅ |
| (gnostic.openapi.v3.property).specification_extension | ❌ |

For more information on how to use each option in your Protobuf file, you can reference [the gnostic.openapi.v3 module documentation](https://buf.build/gnostic/gnostic/docs/main:gnostic.openapi.v3) and the [google/gnostic repo](https://github.com/google/gnostic). Note that this is a new feature, so if find something that isn't supported that you need, please [create an issue](https://github.com/sudorandom/protoc-gen-connect-openapi/issues/new).