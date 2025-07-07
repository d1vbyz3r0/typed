package typed

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/getkin/kin-openapi/openapi3gen"
	"reflect"
	"regexp"
	"sync/atomic"
)

// TypeNameGenerator is an openapi3gen.TypeNameGenerator to use when generating schema
var TypeNameGenerator = NewTypeNameGenerator()

// NewTypeNameGenerator creates default typename generator
func NewTypeNameGenerator() openapi3gen.TypeNameGenerator {
	var (
		anonymousStructRegex = regexp.MustCompile(`^struct\s*{.*}$`)
		anonStructsCount     atomic.Uint64
	)

	return func(t reflect.Type) string {
		t = typing.DerefReflectPtr(t)
		name := t.String()
		if anonymousStructRegex.MatchString(name) {
			name = fmt.Sprintf("AnonymousStruct%d", anonStructsCount.Load())
			anonStructsCount.Add(1)
		}

		return name
	}
}
