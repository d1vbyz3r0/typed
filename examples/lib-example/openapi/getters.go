package openapi

import (
	"github.com/d1vbyz3r0/typed"
	"github.com/getkin/kin-openapi/openapi3"
)

// Getters are written manually if required

func GetRegistry() *typed.Registry {
	return registry
}

func GetSpec() *openapi3.T {
	return spec
}
