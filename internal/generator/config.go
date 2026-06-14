package generator

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

type Config struct {
	GenerateLib     bool         `yaml:"generate-lib"`
	LibPkg          string       `yaml:"lib-pkg"`
	ProcessingHooks []string     `yaml:"processing-hooks"`
	Input           InputConfig  `yaml:"input"`
	Output          OutputConfig `yaml:"output"`
	Debug           bool         `yaml:"debug"`
	Concurrency     int          `yaml:"concurrency"`
}

func (c Config) Validate() error {
	if c.Input.RoutesProviderCtor == "" && !c.GenerateLib {
		return errors.New("routes-provider-ctor is required")
	}

	if c.Input.RoutesProviderPkg == "" && !c.GenerateLib {
		return errors.New("routes-provider-pkg is required")
	}

	for i, h := range c.Input.Handlers {
		if err := h.Validate(); err != nil {
			return fmt.Errorf("validate handler(n=%d,path=%s) config: %w", i, h.Path, err)
		}
	}

	for i, m := range c.Input.Models {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("validate model(n=%d,path=%s) config: %w", i, m.Path, err)
		}
	}

	if err := c.Output.Validate(); err != nil {
		return fmt.Errorf("invalid output config: %w", err)
	}

	if c.GenerateLib && c.LibPkg == "" {
		return errors.New("lib-pkg is required when generate_lib is true")
	}

	return nil
}

type InputConfig struct {
	ApiPrefix          *string          `yaml:"api-prefix,omitempty"`
	Title              string           `yaml:"title"`
	Version            string           `yaml:"version"`
	Servers            []Server         `yaml:"servers"`
	RoutesProviderCtor string           `yaml:"routes-provider-ctor"`
	RoutesProviderPkg  string           `yaml:"routes-provider-pkg"`
	Handlers           []HandlersConfig `yaml:"handlers"`
	Models             []ModelsConfig   `yaml:"models"`
}

type Server struct {
	Url string `yaml:"url"`
}

type ModelsConfig struct {
	Path          string        `yaml:"path"`
	Recursive     bool          `yaml:"recursive"`
	IncludeModels []ModelFilter `yaml:"include"`
	ExcludeModels []ModelFilter `yaml:"exclude"`
}

func (c ModelsConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	for i, filter := range c.IncludeModels {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("validate include filter[%d]: %w", i, err)
		}
	}

	for i, filter := range c.ExcludeModels {
		if err := filter.Validate(); err != nil {
			return fmt.Errorf("validate exclude filter[%d]: %w", i, err)
		}
	}

	return nil
}

type ModelFilter struct {
	// Path matches package path relative to the configured models.path
	Path string `yaml:"path"`
	// ImportPath is a full import path regex
	ImportPath string `yaml:"import-path"`
	// Pkg matches Go package name
	Pkg string `yaml:"pkg"`
	// Name matches model type name
	Name string `yaml:"name"`
}

func (f ModelFilter) Validate() error {
	if f.Path == "" && f.ImportPath == "" && f.Pkg == "" && f.Name == "" {
		return errors.New("empty model filter")
	}
	return nil
}

type CompiledModelFilter struct {
	Path       *regexp.Regexp
	ImportPath *regexp.Regexp
	Pkg        *regexp.Regexp
	Name       *regexp.Regexp
}

type HandlersConfig struct {
	Path      string `yaml:"path"`
	Recursive bool   `yaml:"recursive"`
}

func (c HandlersConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	return nil
}

type OutputConfig struct {
	Path     string `yaml:"path"`
	SpecPath string `yaml:"spec-path"`
}

func (c OutputConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	if c.SpecPath == "" {
		return errors.New("spec-path is required")
	}

	return nil
}
