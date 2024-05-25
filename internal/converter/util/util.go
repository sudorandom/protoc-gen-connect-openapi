package util

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"github.com/sudorandom/protoc-gen-connect-openapi/internal/converter/options"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

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

func MethodToRequestBody(opts options.Options, method protoreflect.MethodDescriptor, isStreaming bool) *v3.RequestBody {
	inputName := string(method.Input().FullName())
	return &v3.RequestBody{
		Content:  MakeMediaTypes(opts, "#/components/schemas/"+FormatTypeRef(inputName), true, isStreaming),
		Required: BoolPtr(true),
	}
}

// MakeMediaTypes generates media types with references to the bodies
func MakeMediaTypes(opts options.Options, ref string, isRequest, isStreaming bool) *orderedmap.Map[string, *v3.MediaType] {
	mediaTypes := orderedmap.New[string, *v3.MediaType]()
	for _, protocol := range options.Protocols {
		isNotAStreamingMethod := isStreaming != protocol.IsStreaming
		isStreamingDisabled := isStreaming && !opts.WithStreaming
		if isNotAStreamingMethod || isStreamingDisabled {
			continue
		}

		_, shouldUse := opts.ContentTypes[protocol.Name]
		if !(isStreaming || shouldUse) {
			continue
		}

		var description string
		if isRequest {
			description = protocol.RequestDesc
		} else {
			description = protocol.ResponseDesc
		}

		// Since this protocol has a description, wrap it
		if description != "" {
			props := orderedmap.New[string, *base.SchemaProxy]()
			props.Set("protobufBinaryContents", base.CreateSchemaProxyRef(ref))
			schema := &base.Schema{
				Type:        []string{"object"},
				Format:      "binary",
				Properties:  props,
				Description: description,
			}
			mediaTypes.Set(protocol.ContentType, &v3.MediaType{
				Schema: base.CreateSchemaProxy(schema),
			})
		} else {
			mediaTypes.Set(protocol.ContentType, &v3.MediaType{
				Schema: base.CreateSchemaProxyRef(ref),
			})
		}
	}
	return mediaTypes
}
