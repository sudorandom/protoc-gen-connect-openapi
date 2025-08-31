# Protovalidate Support
protoc-gen-connect-openapi has support for many [Protovalidate](https://github.com/bufbuild/protovalidate) annotations. Note that not every Protovalidate constraint translates clearly to OpenAPI.

Your protobuf that looks like this:
```protobuf
syntax = "proto3";

import "buf/validate/validate.proto";

message User {
  int32 age = 1 [(buf.validate.field).int32.gte = 18];
}
```

... results in an OpenAPI file that contains the gte constraint as OpenAPI's "minimum" property:
```yaml
components:
  schemas:
    User:
      properties:
        age:
          additionalProperties: false
          description: ""
          minimum: 18
          title: age
          type: integer
```

For custom CEL expressions, it will be added at the end of the description.
```protobuf
syntax = "proto3";

import "buf/validate/validate.proto";

package custom;

message User {
  int32 age = 1 [(buf.validate.field).cel = {
    id: "user.age",
    message: "The user can't be a minor (younger than 18 years old)",
    expression: "this < 18 ? 'User must be at least 18 years old': ''"
  }];
}
```

Result:
```yaml
components:
  schemas:
    custom.User:
      properties:
        age:
          description: |+
            The user can't be a minor (younger than 18 years old):
            ```
            this < 18 ? 'User must be at least 18 years old': ''
            ```

          additionalProperties: false
          description: ""
          title: age
          type: integer
```


## Message Options
| Option | Supported? | Notes |
|---|---|---|
| (buf.validate.message).cel | ✅ | Appended to the 'description' field |
| (buf.validate.message).disabled | ✅ | |
| (buf.validate.message).oneOf | ✅ | |

## Field Options
| Option | Supported? | Notes |
|---|---|---|
| (buf.validate.field).cel | ✅ | Appended to the 'description' field |
| (buf.validate.field).any.in | ✅ | |
| (buf.validate.field).any.not_in | ✅ | |
| (buf.validate.field).bool.const | ✅ | |
| (buf.validate.field).bool.example | ✅ | |
| (buf.validate.field).bytes.const | ✅ | |
| (buf.validate.field).bytes.contains | | |
| (buf.validate.field).bytes.in | ✅ | |
| (buf.validate.field).bytes.ip | ✅ | |
| (buf.validate.field).bytes.ipv4 | ✅ | |
| (buf.validate.field).bytes.ipv6 | ✅ | |
| (buf.validate.field).bytes.len | ✅ | |
| (buf.validate.field).bytes.max_len | ✅ | |
| (buf.validate.field).bytes.min_len | ✅ | |
| (buf.validate.field).bytes.not_in | ✅ | |
| (buf.validate.field).bytes.pattern | ✅ | |
| (buf.validate.field).bytes.prefix | ❌ | |
| (buf.validate.field).bytes.suffix | ❌ | |
| (buf.validate.field).bytes.example | ✅ | |
| (buf.validate.field).double.const | ✅ | |
| (buf.validate.field).double.gt | ✅ | |
| (buf.validate.field).double.gte | ✅ | |
| (buf.validate.field).double.lt | ✅ | |
| (buf.validate.field).double.lte | ✅ | |
| (buf.validate.field).double.example | ✅ | |
| (buf.validate.field).duration.const | ✅ | |
| (buf.validate.field).duration.gt | ❌ | |
| (buf.validate.field).duration.gte | ❌ | |
| (buf.validate.field).duration.in | ✅ | |
| (buf.validate.field).duration.lt | ❌ | |
| (buf.validate.field).duration.lte | ❌ | |
| (buf.validate.field).duration.not_in | ❌ | |
| (buf.validate.field).duration.example | ✅ | |
| (buf.validate.field).enum.const | ✅ | |
| (buf.validate.field).enum.defined_only | ❌ | |
| (buf.validate.field).enum.example | ✅ | |
| (buf.validate.field).fixed32.const | ✅ | |
| (buf.validate.field).fixed32.gt | ✅ | |
| (buf.validate.field).fixed32.gte | ✅ | |
| (buf.validate.field).fixed32.lt | ✅ | |
| (buf.validate.field).fixed32.lte | ✅ | |
| (buf.validate.field).fixed32.example | ✅ | |
| (buf.validate.field).fixed64.const | ✅ | |
| (buf.validate.field).fixed64.gt | ✅ | |
| (buf.validate.field).fixed64.gte | ✅ | |
| (buf.validate.field).fixed64.lt | ✅ | |
| (buf.validate.field).fixed64.lte | ✅ | |
| (buf.validate.field).fixed64.example | ✅ | |
| (buf.validate.field).float.const | ✅ | |
| (buf.validate.field).float.gt | ✅ | |
| (buf.validate.field).float.gte | ✅ | |
| (buf.validate.field).float.lt | ✅ | |
| (buf.validate.field).float.lte | ✅ | |
| (buf.validate.field).float.example | ✅ | |
| (buf.validate.field).int32.const | ✅ | |
| (buf.validate.field).int32.gt | ✅ | |
| (buf.validate.field).int32.gte | ✅ | |
| (buf.validate.field).int32.lt | ✅ | |
| (buf.validate.field).int32.lte | ✅ | |
| (buf.validate.field).int32.example | ✅ | |
| (buf.validate.field).int64.const | ✅ | |
| (buf.validate.field).int64.gt | ✅ | |
| (buf.validate.field).int64.gte | ✅ | |
| (buf.validate.field).int64.lt | ✅ | |
| (buf.validate.field).int64.lte | ✅ | |
| (buf.validate.field).int64.example | ✅ | |
| (buf.validate.field).map.keys | ❌ | |
| (buf.validate.field).map.max_pairs | ✅ | |
| (buf.validate.field).map.min_pairs | ✅ | |
| (buf.validate.field).map.values | ✅ | |
| (buf.validate.field).repeated.items | ✅ | |
| (buf.validate.field).repeated.max_items | ✅ | |
| (buf.validate.field).repeated.min_items | ✅ | |
| (buf.validate.field).repeated.unique | ✅ | |
| (buf.validate.field).required | ✅ | |
| (buf.validate.field).sfixed32.const | ✅ | |
| (buf.validate.field).sfixed32.gt | ✅ | |
| (buf.validate.field).sfixed32.gte | ✅ | |
| (buf.validate.field).sfixed32.lt | ✅ | |
| (buf.validate.field).sfixed32.lte | ✅ | |
| (buf.validate.field).sfixed32.example | ✅ | |
| (buf.validate.field).sfixed64.const | ✅ | |
| (buf.validate.field).sfixed64.gt | ✅ | |
| (buf.validate.field).sfixed64.gte | ✅ | |
| (buf.validate.field).sfixed64.lt | ✅ | |
| (buf.validate.field).sfixed64.lte | ✅ | |
| (buf.validate.field).sfixed64.example | ✅ | |
| (buf.validate.field).sint32.const | ✅ | |
| (buf.validate.field).sint32.gt | ✅ | |
| (buf.validate.field).sint32.gte | ✅ | |
| (buf.validate.field).sint32.lt | ✅ | |
| (buf.validate.field).sint32.lte | ✅ | |
| (buf.validate.field).sint32.example | ✅ | |
| (buf.validate.field).sint64.const | ✅ | |
| (buf.validate.field).sint64.gt | ✅ | |
| (buf.validate.field).sint64.gte | ✅ | |
| (buf.validate.field).sint64.lt | ✅ | |
| (buf.validate.field).sint64.lte | ✅ | |
| (buf.validate.field).sint64.example | ✅ | |
| (buf.validate.field).string.address | | |
| (buf.validate.field).string.const | ✅ | |
| (buf.validate.field).string.contains | ❌ | |
| (buf.validate.field).string.email | ✅ | |
| (buf.validate.field).string.hostname | ✅ | |
| (buf.validate.field).string.in | ✅ | |
| (buf.validate.field).string.ip | ✅ | |
| (buf.validate.field).string.ip_prefix | ❌ | |
| (buf.validate.field).string.ip_with_prefixlen | ❌ | |
| (buf.validate.field).string.ipv4 | ✅ | |
| (buf.validate.field).string.ipv4_prefix | ❌ | |
| (buf.validate.field).string.ipv4_with_prefixlen | ❌ | |
| (buf.validate.field).string.ipv6 | ✅ | |
| (buf.validate.field).string.ipv6_prefix | ❌ | |
| (buf.validate.field).string.ipv6_with_prefixlen | ❌ | |
| (buf.validate.field).string.len | ✅ | |
| (buf.validate.field).string.len_bytes | ❌ | |
| (buf.validate.field).string.max_bytes | ❌ | |
| (buf.validate.field).string.max_len | ✅ | |
| (buf.validate.field).string.min_bytes | ❌ | |
| (buf.validate.field).string.min_len | ✅ | |
| (buf.validate.field).string.not_contains | ❌ | |
| (buf.validate.field).string.not_in | ✅ | |
| (buf.validate.field).string.pattern | ✅ | |
| (buf.validate.field).string.prefix | ❌ | |
| (buf.validate.field).string.strict | ❌ | |
| (buf.validate.field).string.suffix | ❌ | |
| (buf.validate.field).string.uri | ✅ | |
| (buf.validate.field).string.uri_ref | ✅ | |
| (buf.validate.field).string.uuid | ✅ | |
| (buf.validate.field).string.well_known_regex | ❌ | |
| (buf.validate.field).string.example | ✅ | |
| (buf.validate.field).timestamp.const | ✅ | |
| (buf.validate.field).timestamp.gt | ❌ | |
| (buf.validate.field).timestamp.gt_now | ❌ | |
| (buf.validate.field).timestamp.gte | ❌ | |
| (buf.validate.field).timestamp.lt_now | ❌ | |
| (buf.validate.field).timestamp.lte | ❌ | |
| (buf.validate.field).timestamp.within | ❌ | |
| (buf.validate.field).timestamp.example | ✅ | |
| (buf.validate.field).uint32.const | ✅ | |
| (buf.validate.field).uint32.gt | ✅ | |
| (buf.validate.field).uint32.gte | ✅ | |
| (buf.validate.field).uint32.lt | ✅ | |
| (buf.validate.field).uint32.lte | ✅ | |
| (buf.validate.field).uint32.example | ✅ | |
| (buf.validate.field).uint64.const | ✅ | |
| (buf.validate.field).uint64.gt | ✅ | |
| (buf.validate.field).uint64.gte | ✅ | |
| (buf.validate.field).uint64.lt | ✅ | |
| (buf.validate.field).uint64.lte | ✅ | |
| (buf.validate.field).uint64.example | ✅ | |

## OneOf Options
| Option | Supported? | Notes |
|---|---|---|
| (buf.validate.oneof).required | ❌ | |
