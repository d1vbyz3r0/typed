package typed

import (
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

func extractOpTag(path string, prefix string) (string, error) {
	// /api/v1/tasks -> tasks
	p, found := strings.CutPrefix(path, prefix)
	if !found {
		return "", fmt.Errorf("bad api prefix '%s' for path: '%s'", prefix, path)
	}

	p, _ = strings.CutPrefix(p, "/")
	p = strings.Title(p)

	parts := strings.Split(p, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("bad path or prefix provided, nothing to extract after prefix cutoff: %s", parts)
	}

	return parts[0], nil
}

func GenerateRefs(g *openapi3gen.Generator, schemas openapi3.Schemas, registry *Registry) error {
	for instance := range registry.Values() {
		_, err := g.NewSchemaRefForValue(instance, schemas)
		if err != nil {
			return fmt.Errorf("new schema ref for value: %w", err)
		}
	}
	return nil
}
