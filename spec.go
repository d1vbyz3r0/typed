package typed

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

type SpecFormat string

const (
	UndefinedFormat = SpecFormat("")
	YamlFormat      = SpecFormat("yaml")
	JsonFormat      = SpecFormat("json")
)

func getSpecFormat(path string) SpecFormat {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return YamlFormat

	case ".json":
		return JsonFormat

	default:
		return UndefinedFormat
	}
}

func SaveSpec(spec *openapi3.T, outPath string) error {
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	format := getSpecFormat(outPath)
	switch format {
	case YamlFormat:
		enc := yaml.NewEncoder(f)
		enc.SetIndent(2)
		err := enc.Encode(spec)
		if err != nil {
			return fmt.Errorf("encode spec: %w", err)
		}

	case JsonFormat:
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		err := enc.Encode(spec)
		if err != nil {
			return fmt.Errorf("encode spec: %w", err)
		}

	case UndefinedFormat:
		return fmt.Errorf("can't define spec format basing on path, check extension: %s", outPath)
	}

	return nil
}
