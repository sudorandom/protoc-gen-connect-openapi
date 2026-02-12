package visibility

import (
	"strings"

	api_visibility "google.golang.org/genproto/googleapis/api/visibility"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// GetVisibilityRule extracts the visibility rule from a descriptor's options.
// It checks for field, message, method, enum, and enum value visibility extensions.
func GetVisibilityRule(desc protoreflect.Descriptor) *api_visibility.VisibilityRule {
	var options proto.Message
	var extension interface{}
	switch d := desc.(type) {
	case protoreflect.FieldDescriptor:
		options = d.Options().(*descriptorpb.FieldOptions)
		extension = api_visibility.E_FieldVisibility
	case protoreflect.MessageDescriptor:
		options = d.Options().(*descriptorpb.MessageOptions)
		extension = api_visibility.E_MessageVisibility
	case protoreflect.MethodDescriptor:
		options = d.Options().(*descriptorpb.MethodOptions)
		extension = api_visibility.E_MethodVisibility
	case protoreflect.ServiceDescriptor:
		options = d.Options().(*descriptorpb.ServiceOptions)
		extension = api_visibility.E_ApiVisibility
	case protoreflect.EnumDescriptor:
		options = d.Options().(*descriptorpb.EnumOptions)
		extension = api_visibility.E_EnumVisibility
	case protoreflect.EnumValueDescriptor:
		options = d.Options().(*descriptorpb.EnumValueOptions)
		extension = api_visibility.E_ValueVisibility
	default:
		return nil
	}

	if options == nil {
		return nil
	}

	xt, ok := extension.(protoreflect.ExtensionType)
	if !ok {
		return nil
	}

	if !proto.HasExtension(options, xt) {
		return nil
	}

	ext := proto.GetExtension(options, xt)
	if vis, ok := ext.(*api_visibility.VisibilityRule); ok {
		return vis
	}

	return nil
}

// ShouldBeFiltered checks if a given visibility rule's restriction is present in the
// list of enabled restriction selectors. If the rule's restriction is NOT in the list
// of selectors, the element should be filtered.
//
// The restriction field may contain a comma-separated list of labels (e.g.
// "INTERNAL,EXTERNAL"), meaning the element is visible to any of those audiences.
// The element is included if at least one of its labels appears in the selectors.
func ShouldBeFiltered(rule *api_visibility.VisibilityRule, restrictionSelectors map[string]bool) bool {
	if rule == nil {
		return false // No rule, so not filtered (always include elements without visibility rules)
	}
	if len(restrictionSelectors) == 0 {
		return true // Has a rule but no selectors specified, so filter it out
	}
	for _, label := range strings.Split(rule.Restriction, ",") {
		if _, ok := restrictionSelectors[strings.TrimSpace(label)]; ok {
			return false // At least one label matches, so it should NOT be filtered
		}
	}
	return true // No match found, so it should be filtered
}

// ExtractVisibilityRestriction returns the string value of the visibility restriction.
// If the descriptor has no visibility rules, it returns an empty string.
func ExtractVisibilityRestriction(desc protoreflect.Descriptor) string {
	rule := GetVisibilityRule(desc)
	if rule != nil {
		return rule.Restriction
	}
	return ""
}
