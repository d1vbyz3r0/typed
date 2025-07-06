package typed

import (
	"reflect"
	"strings"
)

func ResolveReflectType(typeName string, registry map[string]any) (reflect.Type, bool) {
	switch typeName {
	case "string":
		return reflect.TypeOf(""), true

	case "int":
		return reflect.TypeOf(int(0)), true
	case "bool":
		return reflect.TypeOf(true), true
	case "float64":
		return reflect.TypeOf(float64(0)), true
	}

	if strings.HasPrefix(typeName, "[]") {
		elemName := typeName[2:]
		elemType, ok := ResolveReflectType(elemName, registry)
		if !ok {
			return nil, false
		}
		return reflect.SliceOf(elemType), true
	}

	if strings.HasPrefix(typeName, "map[string]") {
		valName := typeName[len("map[string]"):]
		valType, ok := ResolveReflectType(valName, registry)
		if !ok {
			return nil, false
		}
		return reflect.MapOf(reflect.TypeOf(""), valType), true
	}

	if val, ok := registry[typeName]; ok {
		return reflect.TypeOf(val).Elem(), true
	}

	return nil, false
}

func ExtractTypeName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) == 0 {
		return ""
	}

	var prefix string
	if strings.HasPrefix(fullPath, "[]") {
		prefix = "[]"
	} else if strings.HasPrefix(fullPath, "map[string]") {
		prefix = "map[string]"
	}

	return prefix + parts[len(parts)-1]
}
