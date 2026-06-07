package typed

import (
	"reflect"
	"strings"

	"github.com/d1vbyz3r0/typed/logging"
)

// FieldNameGenerator renames properties in schema, according to echo's `form` and `xml` bind tags.
// Note, that if both `form` and `xml` tags are presented for struct field, form tag will be used
// so use same naming, or don't use one model for both cases
func FieldNameGenerator(field reflect.StructField, defaultName string) string {
	tag := field.Tag
	if v, ok := tag.Lookup("form"); ok && v != "-" {
		return v
	}

	if v, ok := tag.Lookup("xml"); ok && v != "-" {
		parts := strings.Split(v, ",")
		if len(parts) < 1 {
			logging.Warn("found invalid xml tag", "value", tag)
			return defaultName
		}
		return parts[0]
	}

	return defaultName
}
