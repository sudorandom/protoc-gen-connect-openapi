package util

import (
	"strings"
	"unicode"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// IsUnspecifiedEnumValue reports whether value is the "not set" sentinel
// recommended by the Protobuf style guide: number 0, named
// <ENUM_NAME>_UNSPECIFIED. Zero values with any other name are treated as real
// values.
func IsUnspecifiedEnumValue(value protoreflect.EnumValueDescriptor) bool {
	if value.Number() != 0 {
		return false
	}
	return string(value.Name()) == upperSnakeCase(string(value.Parent().Name()))+"_UNSPECIFIED"
}

// upperSnakeCase converts a CamelCase name to UPPER_SNAKE_CASE, keeping
// acronyms together, e.g. HTTPMethod becomes HTTP_METHOD.
func upperSnakeCase(name string) string {
	runes := []rune(name)
	var b strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextIsLower) {
				b.WriteRune('_')
			}
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}
