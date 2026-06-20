package typed

import (
	"slices"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/testdata/dto"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	tests := []struct {
		name    string
		items   []T
		wantErr string
		check   func(t *testing.T, r *Registry)
	}{
		{
			name: "lookup registered named type",
			items: []T{
				{
					Val:         new(dto.User),
					Type:        typing.Named("github.com/acme/dto", "User"),
					ImportAlias: "dto",
				},
			},
			check: func(t *testing.T, r *Registry) {
				item, ok := r.Lookup("github.com/acme/dto", "User")
				require.True(t, ok)
				require.Equal(t, "dto", item.ImportAlias)
				require.Equal(t, typing.Named("github.com/acme/dto", "User").String(), item.Type.String())

				val, ok := r.LookupValue(typing.Named("github.com/acme/dto", "User"))
				require.True(t, ok)
				require.IsType(t, new(dto.User), val)
			},
		},
		{
			name: "lookup missing type",
			items: []T{
				{
					Val:  new(dto.User),
					Type: typing.Named("github.com/acme/dto", "User"),
				},
			},
			check: func(t *testing.T, r *Registry) {
				item, ok := r.Lookup("github.com/acme/dto", "Missing")
				require.False(t, ok)
				require.Zero(t, item)

				val, ok := r.LookupValue(typing.Named("github.com/acme/dto", "Missing"))
				require.False(t, ok)
				require.Nil(t, val)
			},
		},
		{
			name: "lookup slice type value",
			items: []T{
				{
					Val:  new([]dto.User),
					Type: typing.Slice(typing.Named("github.com/acme/dto", "User")),
				},
			},
			check: func(t *testing.T, r *Registry) {
				val, ok := r.LookupValue(
					typing.Slice(typing.Named("github.com/acme/dto", "User")),
				)
				require.True(t, ok)
				require.IsType(t, new([]dto.User), val)
			},
		},
		{
			name: "lookup map type value",
			items: []T{
				{
					Val: new(map[string]dto.User),
					Type: typing.Map(
						typing.Basic("string"),
						typing.Named("github.com/acme/dto", "User"),
					),
				},
			},
			check: func(t *testing.T, r *Registry) {
				val, ok := r.LookupValue(typing.Map(
					typing.Basic("string"),
					typing.Named("github.com/acme/dto", "User"),
				))
				require.True(t, ok)
				require.IsType(t, new(map[string]dto.User), val)
			},
		},
		{
			name: "lookup enum values",
			items: []T{
				{
					Val: new(dto.Status),
					Type: typing.Enum(
						typing.Named("github.com/acme/dto", "Status"),
						[]any{"active", "inactive"},
					),
				},
			},
			check: func(t *testing.T, r *Registry) {
				vals, ok := r.LookupEnumValues("github.com/acme/dto", "Status")
				require.True(t, ok)
				require.ElementsMatch(t, []any{"active", "inactive"}, vals)
			},
		},
		{
			name: "lookup enum values for non enum returns false",
			items: []T{
				{
					Val:  new(dto.User),
					Type: typing.Named("github.com/acme/dto", "User"),
				},
			},
			check: func(t *testing.T, r *Registry) {
				vals, ok := r.LookupEnumValues("github.com/acme/dto", "User")
				require.False(t, ok)
				require.Nil(t, vals)
			},
		},
		{
			name: "values are sorted by type string",
			items: []T{
				{
					Val:  new(dto.User),
					Type: typing.Named("github.com/acme/dto", "User"),
				},
				{
					Val:  new(dto.Form),
					Type: typing.Named("github.com/acme/dto", "Form"),
				},
				{
					Val:  new(string),
					Type: typing.Basic("string"),
				},
			},
			check: func(t *testing.T, r *Registry) {
				got := slices.Collect(r.Values())
				require.Len(t, got, 3)
				require.IsType(t, new(dto.Form), got[0])
				require.IsType(t, new(dto.User), got[1])
				require.IsType(t, new(string), got[2])
			},
		},
		{
			name: "nil val returns error",
			items: []T{
				{
					Val:  nil,
					Type: typing.Named("github.com/acme/dto", "User"),
				},
			},
			wantErr: "val is nil",
		},
		{
			name: "nil type returns error",
			items: []T{
				{
					Val:  new(dto.User),
					Type: nil,
				},
			},
			wantErr: "type is nil",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRegistry(tc.items...)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, r)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, r)

			if tc.check != nil {
				tc.check(t, r)
			}
		})
	}
}
