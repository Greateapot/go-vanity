package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

type Rule struct {
	Match  string     `yaml:"match"`
	Import ImportRule `yaml:"import"`
	Source *Source    `yaml:"source,omitempty"`
}

type ImportRule struct {
	VCS          string  `yaml:"vcs"`
	Repo         string  `yaml:"repo"`
	Subdirectory *string `yaml:"subdirectory,omitempty"`
}

type Source struct {
	Web  string `yaml:"web"`
	File string `yaml:"file"`
	Line string `yaml:"line"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config

	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)

	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.Version != "1.0" {
		return errors.New("config: unknown version")
	}

	if len(c.Rules) == 0 {
		return errors.New("config: no rules defined")
	}

	for i, rule := range c.Rules {
		if err := rule.validate(); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}

	return nil
}

func (r *Rule) validate() error {
	if r.Match == "" {
		return errors.New("match is required")
	}

	if err := validateMatch(r.Match); err != nil {
		return fmt.Errorf("match: %w", err)
	}

	if err := r.Import.validate(); err != nil {
		return fmt.Errorf("import: %w", err)
	}

	if r.Source != nil {
		if err := r.Source.validate(); err != nil {
			return fmt.Errorf("source: %w", err)
		}
	}

	return nil
}

func (r *ImportRule) validate() error {
	switch r.VCS {
	case "git", "hg", "svn", "bzr", "fossil", "mod":
	default:
		return fmt.Errorf("unsupported vcs %q", r.VCS)
	}

	if r.Repo == "" {
		return errors.New("repo is required")
	}

	u, err := url.Parse(r.Repo)
	if err != nil {
		return fmt.Errorf("invalid repo URL: %w", err)
	}

	if u.Scheme == "" || u.Host == "" {
		return errors.New("repo must be an absolute URL")
	}

	if r.Subdirectory != nil {
		if err := validateSubdirectory(*r.Subdirectory); err != nil {
			return fmt.Errorf("subdirectory: %w", err)
		}
	}

	return nil
}

func (s *Source) validate() error {
	if s.Web == "" {
		return errors.New("web is required")
	}

	if s.File == "" {
		return errors.New("file is required")
	}

	if s.Line == "" {
		return errors.New("line is required")
	}

	return nil
}

func validateMatch(match string) error {
	if match == "" {
		return errors.New("must not be empty")
	}

	if strings.ContainsAny(match, " \t\r\n") {
		return errors.New("must not contain whitespace")
	}

	if strings.Count(match, "*") > 1 {
		return errors.New("must contain at most one '*'")
	}

	if strings.Contains(match, "*") && !strings.HasSuffix(match, "/*") {
		return errors.New("'*' is only allowed at the end")
	}

	if !strings.Contains(match, "/") {
		return errors.New("must contain a path")
	}

	return nil
}

func validateSubdirectory(subdirectory string) error {
	if subdirectory == "" {
		return errors.New("must not be empty")
	}

	if strings.HasPrefix(subdirectory, "/") {
		return errors.New("must be a relative path")
	}

	parts := strings.Split(subdirectory, "/")

	for _, part := range parts {
		if part == "" {
			return errors.New("must not contain empty path components")
		}

		if part == "." || part == ".." {
			return errors.New("must not contain '.' or '..'")
		}
	}

	return nil
}
