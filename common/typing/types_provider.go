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
	IntType     = reflect.TypeFor[int]()
	Int64Type   = reflect.TypeFor[int64]()
	UintType    = reflect.TypeFor[uint]()
	Float64Type = reflect.TypeFor[float64]()
	BoolType    = reflect.TypeFor[bool]()
	UuidType    = reflect.TypeFor[uuid.UUID]()
	TimeType    = reflect.TypeFor[time.Time]()
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
