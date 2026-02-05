package util

import (
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	yamlv3 "go.yaml.in/yaml/v3"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func ConvertNodeV3toV4(n *yamlv3.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	newNode := &yaml.Node{
		Kind:        yaml.Kind(n.Kind),
		Style:       yaml.Style(n.Style),
		Tag:         n.Tag,
		Value:       n.Value,
		Anchor:      n.Anchor,
		Alias:       ConvertNodeV3toV4(n.Alias),
		Content:     make([]*yaml.Node, len(n.Content)),
		HeadComment: n.HeadComment,
		LineComment: n.LineComment,
		FootComment: n.FootComment,
		Line:        n.Line,
		Column:      n.Column,
	}
	for i, c := range n.Content {
		newNode.Content[i] = ConvertNodeV3toV4(c)
	}
	return newNode
}

func AppendComponents(spec *v3.Document, components *v3.Components) {
	for pair := components.Schemas.First(); pair != nil; pair = pair.Next() {
		spec.Components.Schemas.Set(pair.Key(), pair.Value())
	}
	for pair := components.Responses.First(); pair != nil; pair = pair.Next() {
		spec.Components.Responses.Set(pair.Key(), pair.Value())
	}
	for pair := components.Parameters.First(); pair != nil; pair = pair.Next() {
		spec.Components.Parameters.Set(pair.Key(), pair.Value())
	}
	for pair := components.Examples.First(); pair != nil; pair = pair.Next() {
		spec.Components.Examples.Set(pair.Key(), pair.Value())
	}
	for pair := components.RequestBodies.First(); pair != nil; pair = pair.Next() {
		spec.Components.RequestBodies.Set(pair.Key(), pair.Value())
	}
	for pair := components.Headers.First(); pair != nil; pair = pair.Next() {
		spec.Components.Headers.Set(pair.Key(), pair.Value())
	}
	for pair := components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
		spec.Components.SecuritySchemes.Set(pair.Key(), pair.Value())
	}
	for pair := components.Links.First(); pair != nil; pair = pair.Next() {
		spec.Components.Links.Set(pair.Key(), pair.Value())
	}
	for pair := components.Callbacks.First(); pair != nil; pair = pair.Next() {
		spec.Components.Callbacks.Set(pair.Key(), pair.Value())
	}
}

func TypeFieldDescription(opts options.Options, tt protoreflect.FieldDescriptor) string {
	b := strings.Builder{}
	b.WriteString(FormatComments(tt.ParentFile().SourceLocations().ByDescriptor(tt)))
	if opts.WithProtoAnnotations {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("(proto ")
		switch tt.Kind() {
		case protoreflect.MessageKind:
			b.WriteString(string(tt.Message().FullName()))
		case protoreflect.EnumKind:
			b.WriteString(string(tt.Enum().FullName()))
		default:
			b.WriteString(tt.Kind().String())
		}
		b.WriteByte(')')
	}
	return b.String()
}

var internalCommentsRegex = regexp.MustCompile(`(?s)\(--.*--\)`)

func filterInternalComments(comments string) string {
	if comments == "" {
		return ""
	}
	filtered := strings.TrimSpace(internalCommentsRegex.ReplaceAllString(comments, ""))
	return filtered
}

func FormatComments(loc protoreflect.SourceLocation) string {
	var builder strings.Builder
	leadingComments := filterInternalComments(loc.LeadingComments)
	if leadingComments != "" {
		builder.WriteString(strings.TrimSpace(leadingComments))
		builder.WriteString(" ")
	}
	trailingComments := filterInternalComments(loc.TrailingComments)
	if trailingComments != "" {
		builder.WriteString(strings.TrimSpace(trailingComments))
		builder.WriteString(" ")
	}
	return strings.TrimSpace(builder.String())
}

func FormatOperationComments(loc protoreflect.SourceLocation) (summary string, description string) {
	var leadingComments = strings.TrimSpace(filterInternalComments(loc.LeadingComments))
	var trailingComments = strings.TrimSpace(filterInternalComments(loc.TrailingComments))

	if leadingComments == "" && trailingComments == "" {
		return "", ""
	}

	// Split leading comments by double newline to separate blocks
	blocks := strings.Split(leadingComments, "\n\n")

	if len(blocks) > 1 {
		// If there are multiple blocks, the first block is the summary, and the rest is the description
		summary = strings.TrimSpace(blocks[0])
		description = strings.Join(blocks[1:], "\n\n")
		description = strings.TrimSpace(description)
	} else {
		// If there's only one block, it serves as the description, and the summary is empty
		summary = ""
		description = strings.TrimSpace(blocks[0])
	}

	// Append trailing comments to the description
	if trailingComments != "" {
		if description != "" {
			description += "\n\n" // Add a blank line if description already exists
		}
		description += trailingComments
	}

	return summary, description
}

func BoolPtr(b bool) *bool {
	return &b
}

func FormatTypeRef(t string) string {
	return strings.TrimPrefix(t, ".")
}

func IsMethodDeprecated(md protoreflect.MethodDescriptor) *bool {
	options, ok := md.Options().(*descriptorpb.MethodOptions)
	if !ok || options == nil {
		return nil
	}
	if options.Deprecated == nil {
		return nil
	}
	return options.Deprecated
}

func IsFieldDeprecated(fd protoreflect.FieldDescriptor) *bool {
	options, ok := fd.Options().(*descriptorpb.FieldOptions)
	if !ok || options == nil {
		return nil
	}
	if options.Deprecated == nil {
		return nil
	}
	return options.Deprecated
}

func MethodToRequestBody(opts options.Options, method protoreflect.MethodDescriptor, s *base.SchemaProxy, isStreaming bool) *v3.RequestBody {
	return &v3.RequestBody{
		Content:  MakeMediaTypes(opts, s, true, isStreaming),
		Required: BoolPtr(true),
	}
}

// MakeMediaTypes generates media types with references to the bodies
func MakeMediaTypes(opts options.Options, s *base.SchemaProxy, isRequest, isStreaming bool) *orderedmap.Map[string, *v3.MediaType] {
	mediaTypes := orderedmap.New[string, *v3.MediaType]()
	for _, protocol := range options.Protocols {
		isStreamingDisabled := isStreaming && !opts.WithStreaming
		if isStreaming != protocol.IsStreaming || isStreamingDisabled {
			continue
		}

		_, shouldUse := opts.ContentTypes[protocol.Name]
		if !isStreaming && !shouldUse {
			continue
		}

		mediaTypes.Set(protocol.ContentType, &v3.MediaType{Schema: s})
	}
	return mediaTypes
}

func MakeFieldName(opts options.Options, fd protoreflect.FieldDescriptor) string {
	if opts.WithProtoNames {
		return string(fd.Name())
	}
	return fd.JSONName()
}

func MakePath(opts options.Options, main string) string {
	return path.Join(opts.PathPrefix, main)
}

func AppendStringDedupe(strs []string, str string) []string {
	if slices.Contains(strs, str) {
		return strs
	}
	return append(strs, str)
}

// Singular returns the singular form of a given plural noun. .
func Singular(plural string) string {

	if strings.HasSuffix(plural, "ves") {
		return strings.TrimSuffix(plural, "ves") + "f"
	}
	if strings.HasSuffix(plural, "ies") {
		return strings.TrimSuffix(plural, "ies") + "y"
	}
	if strings.HasSuffix(plural, "s") {
		return strings.TrimSuffix(plural, "s")
	}
	return plural
}

// ResolveSchemaRef takes a reference string and determines if it is a fully
// qualified protobuf message name. If so, it converts it to a valid OpenAPI
// schema reference. Otherwise, it returns the original string.
func ResolveSchemaRef(ref string) string {
	if strings.HasPrefix(ref, ".") {
		// This is a fully qualified proto name. We need to look up the
		// message and convert it to a schema reference.
		messageName := strings.TrimPrefix(ref, ".")
		return "#/components/schemas/" + FormatTypeRef(messageName)
	}
	return ref
}

// MergeOrAppendParameter merges or appends a parameter to a list of parameters.
func MergeOrAppendParameter(existingParams []*v3.Parameter, newParam *v3.Parameter) []*v3.Parameter {
	found := false
	for _, p := range existingParams {
		if p.Name != newParam.Name || p.In != newParam.In {
			continue
		}
		found = true
		if p.Description == "" && newParam.Description != "" {
			p.Description = newParam.Description
		}
		// If p.Required is nil (not set) and newParam.Required is set, then use newParam.Required.
		// This preserves an explicitly set false in p.Required.
		if p.Required == nil && newParam.Required != nil {
			p.Required = newParam.Required
		}
		if p.Schema == nil && newParam.Schema != nil {
			p.Schema = newParam.Schema
		} else if p.Schema != nil && newParam.Schema != nil {
			// Merge schema properties
			if p.Schema.Schema().Title == "" {
				p.Schema.Schema().Title = newParam.Schema.Schema().Title
			}
			if p.Schema.Schema().Description == "" {
				p.Schema.Schema().Description = newParam.Schema.Schema().Description
			}
			if len(p.Schema.Schema().Type) == 0 {
				p.Schema.Schema().Type = newParam.Schema.Schema().Type
			}
			if p.Schema.Schema().Format == "" {
				p.Schema.Schema().Format = newParam.Schema.Schema().Format
			}
			if len(p.Schema.Schema().Enum) == 0 {
				p.Schema.Schema().Enum = newParam.Schema.Schema().Enum
			}
			if p.Schema.Schema().Default == nil {
				p.Schema.Schema().Default = newParam.Schema.Schema().Default
			}
			if p.Schema.Schema().Items == nil {
				p.Schema.Schema().Items = newParam.Schema.Schema().Items
			}
		}
		// If p.Explode is nil (not set) and newParam.Explode is set, then use newParam.Explode.
		// This preserves an explicitly set false in p.Explode.
		if p.Explode == nil {
			p.Explode = newParam.Explode
		}
		// Assuming Deprecated, AllowEmptyValue, AllowReserved are bool (non-pointer) based on compiler errors
		// This means "empty/nil" is false. We update if current is false.
		if !p.Deprecated { // If p.Deprecated is false
			p.Deprecated = newParam.Deprecated // Set it from newParam
		}
		if !p.AllowEmptyValue { // If p.AllowEmptyValue is false
			p.AllowEmptyValue = newParam.AllowEmptyValue // Set it from newParam
		}
		if p.Style == "" {
			p.Style = newParam.Style
		}
		if !p.AllowReserved { // If p.AllowReserved is false
			p.AllowReserved = newParam.AllowReserved // Set it from newParam
		}
	}
	if !found {
		existingParams = append(existingParams, newParam)
	}
	return existingParams
}
