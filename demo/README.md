# Tooling demos

This directory runs the plugin's output through downstream tooling: documentation sites and type-code generators. It exists to answer "what does the output actually look like in real tools?" — including tools with imperfect JSON Schema 2020-12 support, where a failure is itself useful information.

Everything here is isolated from the plugin itself: a separate `mise.toml` (node, uv) and `package.json`, nothing added to the root Go module or toolchain. The `go` and `buf` tools come from the repository root's mise config.

## Usage

From the repository root, pass a proto file:

```shell
just demo internal/converter/testdata/standard/helloworld.proto
just demo internal/converter/testdata/jsonschema/jsonschema.proto
```

This builds `protoc-gen-connect-openapi` from the working tree, generates both the OpenAPI spec and the JSON Schema bundle for that proto (resolving imports via the nearest `buf.yaml`), runs every tool against them, and serves the results at `http://localhost:4173` (pass a second argument to change the port). An already-generated `*.openapi.yaml`/`*.openapi.json` or `*.jsonschema.json` file is also accepted and demoed directly.

The first run installs the toolchain and npm dependencies. Each invocation builds `demo/out/<name>/` with an index page linking that demo's artifacts; the root index links every demo built so far.

Recipes can also be run from this directory: `just build <input>`, `just serve`, `just clean`.

## What gets built

For a `.proto` input, both sections below. For a spec input, the matching one.

From the OpenAPI spec (`out/<name>/openapi/`):

| Artifact | Tool |
|---|---|
| `redoc.html` | [Redoc](https://redocly.com/docs/cli/) static build |
| `scalar.html` | [Scalar](https://scalar.com/) (loads from CDN) |
| `swagger-ui.html` | [Swagger UI](https://swagger.io/tools/swagger-ui/) (loads from CDN) |
| `elements.html` | [Stoplight Elements](https://stoplight.io/open-source/elements) (loads from CDN) |
| `types.ts` | [openapi-typescript](https://openapi-ts.dev/) |

From the JSON Schema bundle (`out/<name>/jsonschema/`):

| Artifact | Tool |
|---|---|
| `schema-docs.html` | [json-schema-for-humans](https://github.com/coveooss/json-schema-for-humans) static docs |
| `types.ts` | [json-schema-to-typescript](https://github.com/bcherny/json-schema-to-typescript) |
| `types.go` | [quicktype](https://quicktype.io/) |
| `wrapped.schema.json` | The bundle wrapped in a root `anyOf` over every `$defs` entry, since type generators need a root type to start from |

Tools that fail on a given input are reported and skipped rather than aborting the build — an incompatibility (for example, partial draft 2020-12 support in a type generator) is often the interesting result.

## Notes

- The CDN-based documentation pages (Scalar, Swagger UI, Elements) need network access when opened; Redoc and json-schema-for-humans builds are fully static.
- `out/` and `node_modules/` are untracked; `just clean` removes all built demos.
- To add a tool: extend `scripts/build_demo.sh` (a `run_step` + `add_artifact` pair) and, for npm tools, add the dependency to `package.json`.
