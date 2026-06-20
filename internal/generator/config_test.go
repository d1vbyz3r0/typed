package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidateOutputPackage(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "main package requires routes provider",
			cfg: Config{
				Output: OutputConfig{
					Path:     "gen/spec.go",
					SpecPath: "gen/spec.yaml",
				},
			},
			wantErr: "routes-provider-ctor is required",
		},
		{
			name: "main package requires spec path",
			cfg: Config{
				Input: InputConfig{
					RoutesProviderCtor: "NewServer",
					RoutesProviderPkg:  "example.com/project/server",
				},
				Output: OutputConfig{Path: "gen/spec.go"},
			},
			wantErr: "invalid output config: spec-path is required",
		},
		{
			name: "non-main package does not require main inputs",
			cfg: Config{
				Output: OutputConfig{
					Path:        "gen/spec.go",
					PackageName: "spec",
				},
			},
		},
		{
			name: "invalid package name",
			cfg: Config{
				Output: OutputConfig{
					Path:        "gen/spec.go",
					PackageName: "not-valid",
				},
			},
			wantErr: `invalid output config: invalid package name "not-valid"`,
		},
		{
			name: "package name cannot be keyword",
			cfg: Config{
				Output: OutputConfig{
					Path:        "gen/spec.go",
					PackageName: "type",
				},
			},
			wantErr: `invalid output config: invalid package name "type"`,
		},
		{
			name: "package name cannot be blank identifier",
			cfg: Config{
				Output: OutputConfig{
					Path:        "gen/spec.go",
					PackageName: "_",
				},
			},
			wantErr: `invalid output config: invalid package name "_"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.EqualError(t, err, tt.wantErr)
		})
	}
}

func TestOutputConfigPackageDefaultsToMain(t *testing.T) {
	output := OutputConfig{}

	require.Equal(t, "main", output.Package())
	require.True(t, output.IsMain())
}
