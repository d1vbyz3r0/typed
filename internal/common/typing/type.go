package typing

import (
	"github.com/google/uuid"
	"reflect"
	"time"
)

// TODO: add context func call arguments
type Provider func(pkg string, funcName string) (reflect.Type, bool)

var providers = []Provider{
	strconvProvider,
	uuidProvider,
	dateTimeProvider,
}

var (
	intType     = reflect.TypeOf(int(0))
	int64Type   = reflect.TypeOf(int64(0))
	uintType    = reflect.TypeOf(uint(0))
	float64Type = reflect.TypeOf(float64(0))
	boolType    = reflect.TypeOf(false)
	uuidType    = reflect.TypeOf(uuid.UUID{})
	timeType    = reflect.TypeOf(time.Time{})
)

func RegisterTypeProvider(p Provider) {
	providers = append(providers, p)
}

func GetTypeFromUsageContext(pkg string, funcName string) (reflect.Type, bool) {
	for _, provider := range providers {
		t, ok := provider(pkg, funcName)
		if ok {
			return t, ok
		}
	}
	return nil, false
}

func strconvProvider(pkg string, name string) (reflect.Type, bool) {
	if pkg != "strconv" {
		return nil, false
	}

	switch name {
	case "Atoi":
		return intType, true

	case "ParseInt":
		return int64Type, true

	case "ParseUint":
		return uintType, true

	case "ParseFloat":
		return float64Type, true

	case "ParseBool":
		return boolType, true
	}

	return nil, false
}

func uuidProvider(pkg string, funcName string) (reflect.Type, bool) {
	if pkg != "uuid" {
		return nil, false
	}

	if funcName != "MustParse" && funcName != "Parse" {
		return nil, false
	}

	return uuidType, true
}

func dateTimeProvider(pkg string, funcName string) (reflect.Type, bool) {
	if pkg != "time" {
		return nil, false
	}

	if funcName != "Parse" {
		return nil, false
	}

	return timeType, true
}
