package util

import "strings"

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
