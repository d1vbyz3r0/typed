package generator

import "strings"

type PathParam struct {
	Name     string
	Required bool   // Always true for path params
	Type     string // Usually string
}

func extractPathParams(path string) []*PathParam {
	params := make([]*PathParam, 0)
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, &PathParam{
				Name:     strings.TrimPrefix(part, ":"),
				Required: true,
				Type:     "string",
			})
		}
	}
	return params
}
