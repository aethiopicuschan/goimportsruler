package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aethiopicuschan/goimportsruler/pkg/errs"
	"github.com/aethiopicuschan/goimportsruler/pkg/values"
	"gopkg.in/yaml.v3"
)

// Config represents the overall configuration for import rules.
type Config struct {
	configFileName string
	rules          []Rule
	excludes       []Package
}

// ConfigFileName returns the name of the configuration file.
func (c *Config) ConfigFileName() string {
	return c.configFileName
}

// Rules returns the list of import rules.
func (c *Config) Rules() []Rule {
	return c.rules
}

// Excludes returns the list of excludes.
func (c *Config) Excludes() []Package {
	return c.excludes
}

func (c *Config) toDTO() *config {
	rules := make([]rule, len(c.Rules()))
	for i, r := range c.Rules() {
		rules[i] = r.toDTO()
	}

	excludes := make([]pack, len(c.Excludes()))
	for i, e := range c.Excludes() {
		excludes[i] = e.toDTO()
	}

	return &config{
		Rules:    rules,
		Excludes: excludes,
	}
}

// ToJSON writes the configuration as JSON to the provided writer.
func (c *Config) ToJSON(w io.Writer) (err error) {
	dto := c.toDTO()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(dto)
	return
}

// ToYAML writes the configuration as YAML to the provided writer.
func (c *Config) ToYAML(w io.Writer) (err error) {
	dto := c.toDTO()
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	err = encoder.Encode(dto)
	return
}

type config struct {
	Rules    []rule `json:"rules" yaml:"rules"`
	Excludes []pack `json:"excludes" yaml:"excludes"`
}

func (c *config) toConfig() *Config {
	rules := make([]Rule, len(c.Rules))
	for i, r := range c.Rules {
		rules[i] = r.toRule()
	}

	excludes := make([]Package, len(c.Excludes))
	for i, ig := range c.Excludes {
		excludes[i] = ig.toPackage()
	}

	return &Config{
		rules:    rules,
		excludes: excludes,
	}
}

// LoadConfig loads the configuration file by walking up the directory tree starting at startPath.
// startPath can be a directory or a file path. The first matching config file is used.
func LoadConfig(startPath string) (cfg *Config, found string, err error) {
	if startPath == "" {
		startPath = "."
	}

	abs, err := filepath.Abs(startPath)
	if err != nil {
		return
	}

	fi, err := os.Stat(abs)
	if err != nil {
		return
	}

	dir := abs
	if !fi.IsDir() {
		dir = filepath.Dir(abs)
	}

	candidates := values.GetAllConfigFileNames()

	for {
		// Try all candidates in the current directory.
		for _, name := range candidates {
			p := filepath.Join(dir, name)
			if _, statErr := os.Stat(p); statErr == nil {
				found = p
				break
			}
		}
		if found != "" {
			break
		}

		// Walk up to the parent directory. Stop at filesystem root.
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if found == "" {
		err = errs.ErrConfigNotFound
		return
	}

	data, err := os.ReadFile(found)
	if err != nil {
		err = errs.ErrReadFailed
		return
	}

	var dto config

	switch ext := filepath.Ext(found); ext {
	case ".json":
		err = json.Unmarshal(data, &dto)
		if err != nil {
			err = fmt.Errorf("%w %s", errs.ErrInvalidFile, found)
			return
		}
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &dto)
		if err != nil {
			err = fmt.Errorf("%w %s", errs.ErrInvalidFile, found)
			return
		}
	default:
		err = fmt.Errorf("%w %s", errs.ErrInvalidFile, found)
		return
	}

	cfg = dto.toConfig()
	cfg.configFileName = found
	return
}

// ExampleConfig provides an example configuration.
func ExampleConfig() (c *Config) {
	internal := &config{
		Rules: []rule{
			{
				Name:        "Ban pkg to cmd",
				Description: "Disallow imports from pkg to cmd",
				Sources: []pack{
					"pkg/**",
				},
				Disallow: []pack{
					"cmd/**",
				},
				Excludes: []pack{
					".",
				},
			},
		},
		Excludes: []pack{
			"vendor/**",
		},
	}
	c = internal.toConfig()
	return
}
