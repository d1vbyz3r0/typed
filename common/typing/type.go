package typing

import (
	"reflect"
	"time"

	"github.com/google/uuid"
)

// TODO: add context func call arguments
type Provider func(pkg string, funcName string) (reflect.Type, bool)

var providers = []Provider{
	strconvProvider,
	uuidProvider,
	dateTimeProvider,
}

var (
	IntType     = reflect.TypeOf(int(0))
	Int64Type   = reflect.TypeOf(int64(0))
	UintType    = reflect.TypeOf(uint(0))
	Float64Type = reflect.TypeOf(float64(0))
	BoolType    = reflect.TypeOf(false)
	UuidType    = reflect.TypeOf(uuid.UUID{})
	TimeType    = reflect.TypeOf(time.Time{})
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
		return IntType, true

	case "ParseInt":
		return Int64Type, true

	case "ParseUint":
		return UintType, true

	case "ParseFloat":
		return Float64Type, true

	case "ParseBool":
		return BoolType, true
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

	return UuidType, true
}

func dateTimeProvider(pkg string, funcName string) (reflect.Type, bool) {
	if pkg != "time" {
		return nil, false
	}

	if funcName != "Parse" {
		return nil, false
	}

	return TimeType, true
}
