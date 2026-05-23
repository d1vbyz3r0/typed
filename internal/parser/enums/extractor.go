package enums

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/d1vbyz3r0/typed/common/typing"
)

type descriptor struct {
	pkg  string
	name string
}

func Extract(pkg *types.Package, file *ast.File, info *types.Info) ([]*typing.Type, error) {
	enums := make(map[descriptor][]any)
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range valSpec.Names {
				c, ok := info.Defs[name].(*types.Const)
				if !ok {
					continue
				}

				named, ok := c.Type().(*types.Named)
				if !ok {
					continue
				}

				obj := named.Obj()
				if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != pkg.Path() {
					continue
				}

				value, err := constToAny(c.Val())
				if err != nil {
					return nil, fmt.Errorf("const %s: %w", name.Name, err)
				}

				k := descriptor{
					pkg:  obj.Pkg().Path(),
					name: obj.Name(),
				}
				enums[k] = append(enums[k], value)
			}
		}
	}

	res := make([]*typing.Type, 0, len(enums))
	for desc, vals := range enums {
		res = append(res, typing.Enum(
			typing.Named(desc.pkg, desc.name),
			vals,
		))
	}

	return res, nil
}

func constToAny(v constant.Value) (any, error) {
	switch v.Kind() {
	case constant.String:
		return constant.StringVal(v), nil
	case constant.Bool:
		return constant.BoolVal(v), nil
	case constant.Int:
		if i, ok := constant.Int64Val(v); ok {
			return i, nil
		}
		if u, ok := constant.Uint64Val(v); ok {
			return u, nil
		}
		return nil, fmt.Errorf("integer out of int64/uint64 range: %s", v)
	case constant.Float:
		f, ok := constant.Float64Val(v)
		if !ok {
			return nil, fmt.Errorf("float out of float64 range: %s", v)
		}
		return f, nil
	default:
		return nil, fmt.Errorf("unsupported const kind: %s", v.Kind())
	}
}
