package binding

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
