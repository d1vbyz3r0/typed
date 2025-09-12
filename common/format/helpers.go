package format

import (
	"reflect"
	"strconv"
	"strings"
)

func GetFloatTagValue(tag reflect.StructTag, name string) (float64, bool) {
	validate := tag.Get("validate")
	if validate == "" || validate == "-" {
		return 0, false
	}

	parts := strings.Split(validate, ",")
	for _, part := range parts {
		if strings.Contains(part, name) {
			lt := strings.Split(part, "=")
			val := lt[1]
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return 0, false
			}
			return v, true
		}
	}

	return 0, false
}
