package generator

import (
	"maps"
	"slices"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
)

func collectTypes(results []parser.Result) ([]*typing.Type, error) {
	_types := make(map[string]*typing.Type)
	processType := func(t *typing.Type) {
		if t == nil {
			return
		}

		k := typing.ToString(t, typing.DefaultNamer)
		if prev, ok := _types[k]; ok {
			// enums are special case, because they can be found both as Named type and as Enum type.
			// In that case we need to save Enum, not Named (Named describes enum value kind only)
			// TODO: filter that out in parser ?
			if prev.Kind() != typing.TypeKindEnum && t.Kind() == typing.TypeKindEnum {
				_types[k] = t
			}

			return
		}

		_types[k] = t
	}

	for _, res := range results {
		for _, h := range res.Handlers {
			req := h.Request
			if req != nil {
				processType(req.ModelType)
				err := typing.Traverse(req.ModelType, processType)
				if err != nil {
					return nil, err
				}
			}

			for _, responses := range h.Responses {
				for _, resp := range responses {
					processType(resp.ModelType)
					err := typing.Traverse(resp.ModelType, processType)
					if err != nil {
						return nil, err
					}
				}
			}
		}

		for _, model := range res.AdditionalModels {
			processType(model)
			err := typing.Traverse(model, processType)
			if err != nil {
				return nil, err
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

	return res, nil
}
