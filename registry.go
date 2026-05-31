package typed

import (
	"fmt"
	"iter"

	"github.com/d1vbyz3r0/typed/common/typing"
)

// T is descriptor of arbitrary type which holds instance of this type in Val and full type descriptor in Type
type T struct {
	Val         any
	Type        *typing.Type
	ImportAlias string // TODO: fill this shit
}

type Registry struct {
	items map[string]T
	enums map[string][]any
}

func NewRegistry(items ...T) (*Registry, error) {
	r := &Registry{
		items: make(map[string]T, len(items)),
		enums: make(map[string][]any),
	}

	for _, i := range items {
		if i.Val == nil {
			return nil, fmt.Errorf("val is nil")
		}
		if i.Type == nil {
			return nil, fmt.Errorf("type is nil")
		}

		r.items[i.Type.String()] = i
		if i.Type.Kind() == typing.TypeKindEnum {
			r.enums[i.Type.String()] = i.Type.EnumValues()
		}
	}

	return r, nil
}

func MustNewRegistry(items ...T) *Registry {
	r, err := NewRegistry(items...)
	if err != nil {
		panic(err)
	}
	return r
}

func (r *Registry) Lookup(pkg string, name string) (T, bool) {
	k := pkg + "." + name
	v, ok := r.items[k]
	return v, ok
}

func (r *Registry) LookupType(pkg string, name string) (*typing.Type, bool) {
	k := pkg + "." + name
	t, ok := r.items[k]
	return t.Type, ok
}

func (r *Registry) LookupValue(t *typing.Type) (any, bool) {
	k := t.String()
	res, ok := r.items[k]
	return res.Val, ok
}

func (r *Registry) LookupEnumValues(pkg string, name string) ([]any, bool) {
	k := pkg + "." + name
	vals, ok := r.enums[k]
	return vals, ok
}

// Values returns iterator over items
func (r *Registry) Values() iter.Seq[any] {
	return func(yield func(any) bool) {
		for _, v := range r.items {
			if !yield(v.Val) {
				return
			}
		}
	}
}
