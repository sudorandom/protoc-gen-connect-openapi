package util

import (
	"path"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

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

func FormatComments(loc protoreflect.SourceLocation) string {
	var builder strings.Builder
	if loc.LeadingComments != "" {
		builder.WriteString(strings.TrimSpace(loc.LeadingComments))
		builder.WriteString(" ")
	}
	if loc.TrailingComments != "" {
		builder.WriteString(strings.TrimSpace(loc.TrailingComments))
		builder.WriteString(" ")
	}
	return strings.TrimSpace(builder.String())
}

func FormatOperationComments(loc protoreflect.SourceLocation) (summary string, description string) {
	var leadingComments = strings.TrimSpace(loc.LeadingComments)
	var trailingComments = strings.TrimSpace(loc.TrailingComments)

	if leadingComments == "" && trailingComments == "" {
		return "", ""
	}

	// Split leading comments by double newline to separate blocks
	blocks := strings.Split(leadingComments, "\n\n")

	// The first block is the summary
	summary = strings.ReplaceAll(strings.TrimSpace(blocks[0]), "\n", " ")

	// The rest of the blocks form the description
	if len(blocks) > 1 {
		description = strings.Join(blocks[1:], "\n\n")
		description = strings.TrimSpace(description)
	} else {
		// If there's only one block, it serves as both summary and description
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
	for _, s := range strs {
		if str == s {
			return strs
		}
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
