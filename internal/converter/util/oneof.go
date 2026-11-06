package util

import (
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// OneofNote describes a oneof constraint in prose. Renderers don't display
// the composition keywords that enforce it, so the schema description is the
// only place a reader learns which fields exclude each other.
func OneofNote(fieldNames []string, required bool) string {
	quoted := make([]string, len(fieldNames))
	for i, name := range fieldNames {
		quoted[i] = "`" + name + "`"
	}
	list := strings.Join(quoted, ", ")
	if required {
		return "Exactly one of " + list + " must be set."
	}
	return "At most one of " + list + " may be set."
}

// AppendDescription adds a paragraph to the end of a description.
func AppendDescription(description, paragraph string) string {
	if description == "" {
		return paragraph
	}
	return description + "\n\n" + paragraph
}

// maxPairwiseOneofMembers is the largest oneof whose at-most-one rule is
// written as forbidden pairs. Pairs read naturally but grow with the square of
// the member count (4 members give 6 pairs), so larger oneofs use a form whose
// size grows linearly.
const maxPairwiseOneofMembers = 4

// OneofViolations returns schemas that match when more than one of the
// oneof's members is set. A plain protobuf oneof permits no member to be set;
// required oneofs add OneofNoneSet on top.
func OneofViolations(fieldNames []string) []*base.SchemaProxy {
	if len(fieldNames) < 2 {
		return nil
	}
	if len(fieldNames) > maxPairwiseOneofMembers {
		// Some member set, but not exactly one.
		return []*base.SchemaProxy{base.CreateSchemaProxy(&base.Schema{
			AllOf: []*base.SchemaProxy{
				base.CreateSchemaProxy(&base.Schema{AnyOf: presenceOf(fieldNames)}),
				base.CreateSchemaProxy(&base.Schema{
					Not: base.CreateSchemaProxy(&base.Schema{OneOf: presenceOf(fieldNames)}),
				}),
			},
		})}
	}
	var pairs []*base.SchemaProxy
	for i, first := range fieldNames {
		for _, second := range fieldNames[i+1:] {
			pairs = append(pairs, base.CreateSchemaProxy(&base.Schema{
				Required: []string{first, second},
			}))
		}
	}
	return pairs
}

// presenceOf builds one "this field is set" branch per name. oneOf over them
// matches when exactly one is set, anyOf when at least one is.
func presenceOf(fieldNames []string) []*base.SchemaProxy {
	branches := make([]*base.SchemaProxy, 0, len(fieldNames))
	for _, fieldName := range fieldNames {
		branches = append(branches, base.CreateSchemaProxy(&base.Schema{
			Required: []string{fieldName},
		}))
	}
	return branches
}

// OneofNoneSet returns a schema that matches when none of the oneof's members
// is set: the violation of a required oneof.
func OneofNoneSet(fieldNames []string) *base.SchemaProxy {
	return base.CreateSchemaProxy(&base.Schema{
		Not: base.CreateSchemaProxy(&base.Schema{AnyOf: presenceOf(fieldNames)}),
	})
}

// AddViolations forbids instances matching any of the violations by adding
// them to the schema's not. Oneof constraints live there, rather than in
// allOf/oneOf, because documentation renderers present oneOf branches as
// alternative shapes: presence-only branches render as empty or misleading
// variants, while not is ignored or shown collapsed.
func AddViolations(s *base.Schema, violations ...*base.SchemaProxy) {
	if len(violations) == 0 {
		return
	}
	var existing []*base.SchemaProxy
	if s.Not != nil {
		if sub := s.Not.Schema(); sub != nil && isPlainAnyOf(sub) {
			existing = sub.AnyOf
		} else {
			existing = []*base.SchemaProxy{s.Not}
		}
	}
	all := append(existing, violations...)
	if len(all) == 1 {
		s.Not = all[0]
		return
	}
	s.Not = base.CreateSchemaProxy(&base.Schema{AnyOf: all})
}

// isPlainAnyOf reports whether a schema is nothing but an anyOf, so further
// alternatives can be appended to it without changing its meaning.
func isPlainAnyOf(s *base.Schema) bool {
	return len(s.AnyOf) > 0 && len(s.AllOf) == 0 && len(s.OneOf) == 0 &&
		s.Not == nil && len(s.Required) == 0 && len(s.Type) == 0 &&
		(s.Properties == nil || s.Properties.Len() == 0)
}
