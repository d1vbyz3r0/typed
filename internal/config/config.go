package generator

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

type Config struct {
	Input  InputConfig  `yaml:"input"`
	Output OutputConfig `yaml:"output"`
	Debug  bool         `yaml:"debug"`
}

func (c *Config) Validate() error {
	if err := c.Input.Validate(); err != nil {
		return fmt.Errorf("invalid input config: %w", err)
	}

	if err := c.Output.Validate(); err != nil {
		return fmt.Errorf("invalid output config: %w", err)
	}

	return nil
}

type InputConfig struct {
	ApiPrefix          *string          `yaml:"api-prefix,omitempty"`
	Title              string           `yaml:"title"`
	Version            string           `yaml:"version"`
	ServerUrl          string           `yaml:"server-url"`
	RoutesProviderCtor string           `yaml:"routes-provider-ctor"`
	RoutesProviderPkg  string           `yaml:"routes-provider-pkg"`
	Handlers           []HandlersConfig `yaml:"handlers"`
	Models             []ModelsConfig   `yaml:"models"`
}

func (c *InputConfig) Validate() error {
	if c.RoutesProviderCtor == "" {
		return errors.New("routes-provider-ctor is required")
	}

	if c.RoutesProviderPkg == "" {
		return errors.New("routes-provider-pkg is required")
	}

	for i, h := range c.Handlers {
		if err := h.Validate(); err != nil {
			return fmt.Errorf("validate handler[%d] config: %w", i, err)
		}
	}

	for i, m := range c.Models {
		if err := m.Validate(); err != nil {
			return fmt.Errorf("validate model[%d] config: %w", i, err)
		}
	}

	return nil
}

type ModelsConfig struct {
	Path      string  `yaml:"path"`
	Recursive bool    `yaml:"recursive"`
	Filter    *string `yaml:"filter,omitempty"`
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
