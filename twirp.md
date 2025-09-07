# Twirp Support

`protoc-gen-connect-openapi` provides support for [Twirp](https://twitchtv.github.io/twirp/docs/intro.html) services.

To enable Twirp support, you need to explicitly enable the `twirp` feature.

Here's an example `buf.gen.yaml` configuration:

```yaml
version: v2
plugins:
  - remote: buf.build/community/sudorandom-connect-openapi:v0.19.1
    out: gen
    opt:
      - features=twirp
```

When the `twirp` feature is enabled, `protoc-gen-connect-openapi` will generate OpenAPI specifications for your Twirp services.
