package binding

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestHasTag(t *testing.T) {
	field := types.NewVar(token.NoPos, nil, "Field", types.Typ[types.String])
	t.Run("tag present and non-empty", func(t *testing.T) {
		s := types.NewStruct([]*types.Var{field}, []string{`form:"username"`})
		assert.True(t, HasTag(s, "form"))
	})

	t.Run("tag present but dash", func(t *testing.T) {
		s := types.NewStruct([]*types.Var{field}, []string{`form:"-"`})
		assert.False(t, HasTag(s, "form"))
	})

	t.Run("tag missing", func(t *testing.T) {
		s := types.NewStruct([]*types.Var{field}, []string{`json:"name"`})
		assert.False(t, HasTag(s, "form"))
	})
}

func TestHasTags(t *testing.T) {
	field1 := types.NewVar(token.NoPos, nil, "A", types.Typ[types.String])
	field2 := types.NewVar(token.NoPos, nil, "B", types.Typ[types.Int])

	t.Run("all fields contain at least one of the tags", func(t *testing.T) {
		s := types.NewStruct(
			[]*types.Var{field1, field2},
			[]string{`form:"x"`, `query:"y"`},
		)
		assert.True(t, HasTags(s, []string{"form", "query"}))
	})

	t.Run("empty tag and query tag", func(t *testing.T) {
		s := types.NewStruct(
			[]*types.Var{field1, field2},
			[]string{`form:"x"`, `query:"y"`},
		)
		assert.True(t, HasTags(s, []string{"form", "query"}))

	})

	t.Run("one field contains none of the tags", func(t *testing.T) {
		s := types.NewStruct(
			[]*types.Var{field1, field2},
			[]string{`form:"x"`, `json:"z"`}, // second field has neither form nor query
		)
		assert.False(t, HasTags(s, []string{"form", "query"}))
	})

	t.Run("tags are present but dash", func(t *testing.T) {
		s := types.NewStruct(
			[]*types.Var{field1, field2},
			[]string{`form:"-"`, `query:"-"`},
		)
		assert.False(t, HasTags(s, []string{"form", "query"}))
	})
}

func TestHasFiles_SingleFile(t *testing.T) {
	mimePkg := types.NewPackage("mime/multipart", "multipart")
	mimeType := types.NewNamed(types.NewTypeName(0, mimePkg, "FileHeader", nil), nil, nil)

	fields := []*types.Var{
		types.NewField(0, mimePkg, "File", mimeType, false),
	}

	tags := []string{
		`form:"file"`,
	}

	s := types.NewStruct(fields, tags)
	hasFiles := HasFiles(s)
	require.True(t, hasFiles)
}

func TestHasFiles_FileArray(t *testing.T) {
	mimePkg := types.NewPackage("mime/multipart", "multipart")
	mimeType := types.NewNamed(types.NewTypeName(0, mimePkg, "FileHeader", nil), nil, nil)
	mimeSlice := types.NewSlice(mimeType)

	fields := []*types.Var{
		types.NewField(0, mimePkg, "Files", mimeSlice, false),
	}

	tags := []string{
		`form:"files[]"`,
	}

	s := types.NewStruct(fields, tags)
	hasFiles := HasFiles(s)
	require.True(t, hasFiles)
}

func TestHasFiles_NoFiles(t *testing.T) {
	fields := []*types.Var{
		types.NewField(0, nil, "Name", types.Typ[types.String], false),
	}

	tags := []string{
		`form:"name"`,
	}

	s := types.NewStruct(fields, tags)
	hasFiles := HasFiles(s)
	require.False(t, hasFiles)
}

func TestHasAtLeastOneFieldWithoutBindingTag(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected bool
	}{
		{
			name: "all fields have json tags",
			src: `
				package test
				type Req struct {
					ID   int    ` + "`json:\"id\"`" + `
					Name string ` + "`json:\"name\"`" + `
				}
			`,
			expected: false,
		},
		{
			name: "one field without tag",
			src: `
				package test
				type Req struct {
					ID   int
					Name string ` + "`json:\"name\"`" + `
				}
			`,
			expected: true,
		},
		{
			name: "only validate tags",
			src: `
				package test
				type Req struct {
					ID   int    ` + "`validate:\"required\"`" + `
					Name string ` + "`validate:\"min=1\"`" + `
				}
			`,
			expected: true, // validate не считается биндинг-тэгом
		},
		{
			name: "mixed validate and json",
			src: `
				package test
				type Req struct {
					ID   int    ` + "`json:\"id\" validate:\"required\"`" + `
					Name string ` + "`validate:\"required\"`" + `
				}
			`,
			expected: true, // Name без биндинг-тэга
		},
		{
			name: "form tags only",
			src: `
				package test
				type Req struct {
					Name string ` + "`form:\"name\"`" + `
					Age  int    ` + "`form:\"age\"`" + `
				}
			`,
			expected: false,
		},
		{
			name: "path and header tags",
			src: `
				package test
				type Req struct {
					ID   int    ` + "`param:\"id\"`" + `
					User string ` + "`header:\"X-User\"`" + `
				}
			`,
			expected: false,
		},
		{
			name: "mixed path and untagged field",
			src: `
				package test
				type Req struct {
					ID   int    ` + "`path:\"id\"`" + `
					Name string
				}
			`,
			expected: true,
		},
	}

	var knownBindingTags = []string{
		"json",
		"xml",
		"form",
	}

	var skip = []string{
		"query",
		"header",
		"param",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.src, parser.ParseComments)
			require.NoError(t, err)

			conf := types.Config{Importer: nil}
			info := &types.Info{
				Types: make(map[ast.Expr]types.TypeAndValue),
				Defs:  make(map[*ast.Ident]types.Object),
				Uses:  make(map[*ast.Ident]types.Object),
			}

			pkg, err := conf.Check("test", fset, []*ast.File{file}, info)
			require.NoError(t, err)

			obj := pkg.Scope().Lookup("Req")
			require.NotNil(t, obj)

			typ, ok := obj.Type().Underlying().(*types.Struct)
			require.True(t, ok)

			got := HasAtLeastOneFieldWithoutBindingTag(typ, knownBindingTags, skip)
			require.Equal(t, tt.expected, got)
		})
	}
}
