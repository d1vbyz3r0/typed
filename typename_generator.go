package typed

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// TypeNameGenerator is an openapi3gen.TypeNameGenerator to use when generating schema
var TypeNameGenerator = NewTypeNameGenerator()

// NewTypeNameGenerator creates default typename generator
func NewTypeNameGenerator() openapi3gen.TypeNameGenerator {
	anonStructsCount := 1
	anonStructsIndex := make(map[string]string)
	anonymousStructRegex := regexp.MustCompile(`^struct\s*{.*}$`)

	return func(t reflect.Type) string {
		t = typing.DerefReflectPtr(t)
		name := t.String()
		if anonymousStructRegex.MatchString(name) {
			if anonName, ok := anonStructsIndex[name]; ok {
				return anonName
			}

			newName := fmt.Sprintf("AnonStruct%d", anonStructsCount)
			anonStructsCount++

			anonStructsIndex[name] = newName
			return newName
		}

		return name
	}
}
