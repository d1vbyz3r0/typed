package generator

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
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

func (c *Config) Validate() error {
	if c.Input.RoutesProviderCtor == "" && !c.GenerateLib {
		return errors.New("routes-provider-ctor is required")
	}

	if c.Input.RoutesProviderPkg == "" && !c.GenerateLib {
		return errors.New("routes-provider-pkg is required")
	}

	for i, h := range c.Input.Handlers {
		if err := h.Validate(); err != nil {
			return fmt.Errorf("validate handler[%d] config: %w", i, err)
		}
	}

	for i, m := range c.Input.Models {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("validate model[%d] config: %w", i, err)
		}
	}

	if err := c.Output.Validate(); err != nil {
		return fmt.Errorf("invalid output config: %w", err)
	}

	if c.GenerateLib && c.LibPkg == "" {
		return errors.New("lib_pkg is required when generate_lib is true")
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
	Path          string   `yaml:"path"`
	Recursive     bool     `yaml:"recursive"`
	IncludeModels []string `yaml:"include-models"`
	ExcludeModels []string `yaml:"exclude-models"`
}

func (c *ModelsConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	return nil
}

type HandlersConfig struct {
	Path      string `yaml:"path"`
	Recursive bool   `yaml:"recursive"`
}

func (c *HandlersConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	return nil
}

type OutputConfig struct {
	Path     string `yaml:"path"`
	SpecPath string `yaml:"spec-path"`
}

func (c *OutputConfig) Validate() error {
	if c.Path == "" {
		return errors.New("path is required")
	}

	if c.SpecPath == "" {
		return errors.New("spec-path is required")
	}

	return nil
}
