package generator

import (
	"maps"
	"slices"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
)

func collectTypes(results []parser.Result) []*typing.Type {
	_types := make(map[string]*typing.Type)
	for _, res := range results {
		for _, h := range res.Handlers {
			req := h.Request
			if req != nil && req.ModelType != nil {
				k := typing.ToString(req.ModelType, typing.DefaultNamer)
				_types[k] = req.ModelType
			}

			for _, responses := range h.Responses {
				for _, resp := range responses {
					if resp.ModelType == nil {
						continue
					}
					k := typing.ToString(resp.ModelType, typing.DefaultNamer)
					_types[k] = resp.ModelType
				}
			}
		}

		for _, model := range res.AdditionalModels {
			k := typing.ToString(model, typing.DefaultNamer)
			if prev, ok := _types[k]; ok {
				// enums are special case, because they can be found both as Named type and as Enum type.
				// In that case we need to save Enum, not Named (Named describes enum value kind only)
				// TODO: filter that out in parser ?
				if prev.Kind() != typing.TypeKindEnum && model.Kind() == typing.TypeKindEnum {
					_types[k] = model
				}
			} else {
				_types[k] = model
			}
		}
	}

	res := slices.Collect(maps.Values(_types))
	slices.SortFunc(res, func(a, b *typing.Type) int {
		if a.String() < b.String() {
			return -1
		} else if a.String() > b.String() {
			return 1
		}
		return 0
	})
	return res
}
