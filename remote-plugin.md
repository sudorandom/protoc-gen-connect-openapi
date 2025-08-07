# Remote Plugin Usage

`protoc-gen-connect-openapi` can be used as a remote plugin with tools like `buf`. However, when running in a remote context, there are certain security restrictions that apply. Specifically, remote plugins are not allowed to access the local filesystem or network.

For more details on these restrictions, please see the Buf documentation on custom plugin requirements.

## Disabled Options

Due to the filesystem access restrictions, the following options are disabled when using `protoc-gen-connect-openapi` as a remote plugin:

| Option   | Description                                                                         |
|----------|-------------------------------------------------------------------------------------|
| `base`     | The path to a base OpenAPI file to populate fields that this tool doesn't populate. |
| `override` | The path to an override OpenAPI file to override schema components.                 |

Attempting to use these options with a remote plugin will result in an error.

## Alternative: Gnostic Annotations

As an alternative to using `base` and `override` files, you can use OpenAPI v3 annotations from the gnostic project. These annotations allow you to embed OpenAPI specification details directly within your `.proto` files.

This approach is fully compatible with remote plugins as it doesn't require any filesystem access.

For a comprehensive guide on all supported gnostic annotations, please see Gnostic Support.

### Example

Here is a small example of how you can use gnostic annotations to define top-level document information, which is a common use case for a `base` file.

```protobuf
syntax = "proto3";

package my.service.v1;

import "gnostic/openapi/v3/annotations.proto";

option (gnostic.openapi.v3.document) = {
  info: {
    title: "My Awesome API";
    version: "1.0.0";
    description: "This is a sample server for a pet store.";
    contact: {
      name: "API Support";
      url: "http://www.example.com/support";
      email: "support@example.com";
    }
    license: {
      name: "Apache 2.0";
      url: "http://www.apache.org/licenses/LICENSE-2.0.html";
    }
  }
};

service MyService {
  // ...
}
```

By using these annotations, you can achieve the same level of customization as with `base` and `override` files, while remaining compatible with the remote plugin execution environment. For more about gnostic annotations, see [the page about using gnostic with protoc-gen-connect-openapi](gnostic.md).

## Alternative: Using a local plugin

You can always use a local plugin if you absolutely need these options. See the [README](README.md) for more information.
