package converter

import (
	"fmt"
	"log/slog"
	"strings"

	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	libjson "github.com/pb33f/libopenapi/json"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/pb33f/libopenapi/utils"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/visibility"
)

// jsonSchemaDialect is the dialect of the standalone JSON Schema output. The
// generated schemas target OpenAPI 3.1, whose schema objects are a superset of
// draft 2020-12; the few OpenAPI-specific keywords that can appear (e.g.
// gnostic-provided 'example') are ignored by JSON Schema validators.
const jsonSchemaDialect = "https://json-schema.org/draft/2020-12/schema"

const openAPISchemaRefPrefix = "#/components/schemas/"

// appendJSONSchemasToSpec collects the message and enum schemas from fd into
// spec.Components.Schemas. Unlike appendToSpec, it never generates HTTP path
// items, so protocol-specific schemas (connect errors, headers, etc.) are
// never added.
func appendJSONSchemasToSpec(opts options.Options, spec *v3.Document, fd protoreflect.FileDescriptor) {
	// Only collect types from the root if TrimUnusedTypes is off
	if !opts.TrimUnusedTypes {
		enums := fd.Enums()
		for i := 0; i < enums.Len(); i++ {
			enum := enums.Get(i)
			if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(enum), opts.AllowedVisibilities) {
				continue
			}
			AddEnumToSchema(opts, enum, spec)
		}

		messages := fd.Messages()
		for i := 0; i < messages.Len(); i++ {
			message := messages.Get(i)
			if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(message), opts.AllowedVisibilities) {
				continue
			}
			AddMessageSchemas(opts, message, spec)
		}
	}

	// Types referenced by service methods can live in imported files, so they
	// are collected even when root collection already ran. With TrimUnusedTypes
	// they are the only types collected.
	services := fd.Services()
	for i := 0; i < services.Len(); i++ {
		service := services.Get(i)
		if !opts.HasService(service.FullName()) {
			continue
		}
		if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(service), opts.AllowedVisibilities) {
			opts.Logger.Debug("Filtering service due to visibility", slog.String("service", string(service.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
			continue
		}
		methods := service.Methods()
		for j := 0; j < methods.Len(); j++ {
			method := methods.Get(j)
			if visibility.ShouldBeFiltered(visibility.GetVisibilityRule(method), opts.AllowedVisibilities) {
				opts.Logger.Debug("Filtering method due to visibility", slog.String("method", string(method.FullName())), slog.Any("restriction_selectors", opts.AllowedVisibilities))
				continue
			}
			AddMessageSchemas(opts, method.Input(), spec)
			AddMessageSchemas(opts, method.Output(), spec)
		}
	}

	spec.Components.Schemas = orderedmap.SortAlpha(spec.Components.Schemas)
}

// renderJSONSchema renders the document's component schemas as a standalone
// JSON Schema (draft 2020-12) document with every type under $defs, rewriting
// internal references from '#/components/schemas/' to '#/$defs/'.
func renderJSONSchema(spec *v3.Document) (string, error) {
	defs := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for pair := spec.Components.Schemas.First(); pair != nil; pair = pair.Next() {
		rendered, err := pair.Value().MarshalYAML()
		if err != nil {
			return "", fmt.Errorf("rendering schema %q: %w", pair.Key(), err)
		}
		node, ok := rendered.(*yaml.Node)
		if !ok {
			return "", fmt.Errorf("rendering schema %q: unexpected node type %T", pair.Key(), rendered)
		}
		rewriteSchemaRefs(node)
		defs.Content = append(defs.Content, utils.CreateStringNode(pair.Key()), node)
	}

	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content,
		utils.CreateStringNode("$schema"), utils.CreateStringNode(jsonSchemaDialect),
		utils.CreateStringNode("$defs"), defs,
	)

	b, err := libjson.YAMLNodeToJSON(root, "  ")
	if err != nil {
		return "", fmt.Errorf("rendering JSON schema document: %w", err)
	}
	return string(b), nil
}

// rewriteSchemaRefs walks a rendered schema and rewrites every $ref that
// points into '#/components/schemas/' to the '#/$defs/' equivalent.
func rewriteSchemaRefs(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.MappingNode {
		for i := 1; i < len(node.Content); i += 2 {
			key, value := node.Content[i-1], node.Content[i]
			if key.Value == "$ref" && value.Kind == yaml.ScalarNode && strings.HasPrefix(value.Value, openAPISchemaRefPrefix) {
				value.Value = "#/$defs/" + strings.TrimPrefix(value.Value, openAPISchemaRefPrefix)
			}
		}
	}
	for _, child := range node.Content {
		rewriteSchemaRefs(child)
	}
}
