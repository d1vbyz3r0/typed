package typed

import (
	"fmt"
	"iter"
	"maps"
	"slices"

	"github.com/d1vbyz3r0/typed/common/typing"
)

// T describes a Go type registered for schema generation.
// Val holds a sample value, Type holds the full type descriptor,
// and ImportAlias optionally stores the generated import alias.
type T struct {
	Val         any
	Type        *typing.Type
	ImportAlias string
}

type Registry struct {
	items map[string]T
}

func NewRegistry(items ...T) (*Registry, error) {
	r := &Registry{
		items: make(map[string]T, len(items)),
	}

	for _, item := range items {
		if item.Val == nil {
			return nil, fmt.Errorf("val is nil")
		}
		if item.Type == nil {
			return nil, fmt.Errorf("type is nil")
		}

		r.items[item.Type.String()] = item
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

// Lookup returns type descriptor for provided pkg and type name if it was registered
func (r *Registry) Lookup(pkg string, name string) (T, bool) {
	k := makeTypeKey(pkg, name)
	v, ok := r.items[k]
	return v, ok
}

// LookupValue returns instance for provided type descriptor if it was registered
func (r *Registry) LookupValue(t *typing.Type) (any, bool) {
	k := t.String()
	res, ok := r.items[k]
	return res.Val, ok
}

// LookupEnumValues returns possible enum values for requested type, if enum was registered
func (r *Registry) LookupEnumValues(pkg string, name string) ([]any, bool) {
	item, ok := r.Lookup(pkg, name)
	if !ok || item.Type.Kind() != typing.TypeKindEnum {
		return nil, false
	}
	return item.Type.EnumValues(), true
}

// Values returns iterator over registry items.
// Items are sorted by type descriptor (typing.Type) string representation
func (r *Registry) Values() iter.Seq[any] {
	items := slices.Collect(maps.Values(r.items))
	slices.SortFunc(items, func(a, b T) int {
		if a.Type.String() < b.Type.String() {
			return -1
		} else if a.Type.String() > b.Type.String() {
			return 1
		}
		return 0
	})

	return func(yield func(any) bool) {
		for _, v := range items {
			if !yield(v.Val) {
				return
			}
		}
	}
}

func makeTypeKey(pkg string, name string) string {
	if pkg == "" {
		// if pkg is empty - type probably is basic
		return name
	}
	return pkg + "." + name
}
