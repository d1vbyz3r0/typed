package typed

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/logging"
	"github.com/getkin/kin-openapi/openapi3gen"
)

func NewTypeNameGenerator(registry *Registry) openapi3gen.TypeNameGenerator {
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

		pkg := t.PkgPath()
		name = t.Name()
		_type, ok := registry.Lookup(pkg, name)
		if !ok {
			// TODO: need a workaround for nested types, now conflicts can occur
			logging.Debug("type not found in registry, using reflect string representation", "pkg", pkg, "name", name, "reflect_type", t)
			return t.String()
		}

		if _type.Type.Kind() == typing.TypeKindBasic {
			return t.String()
		}

		return _type.ImportAlias + "." + t.Name()
	}
}
