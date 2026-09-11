package main

import (
	"errors"
	"fmt"
	"strings"
)

type Matcher struct {
	rules []Rule
}

func NewMatcher(config *Config) (*Matcher, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	rules := make([]Rule, len(config.Rules))
	copy(rules, config.Rules)

	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Match == rules[j].Match {
				return nil, fmt.Errorf(
					"duplicate match pattern %q",
					rules[i].Match,
				)
			}
		}
	}

	return &Matcher{
		rules: rules,
	}, nil
}

func (m *Matcher) Match(importPath string) (*Rule, bool) {
	var matched *Rule
	var matchedLength int

	for i := range m.rules {
		rule := &m.rules[i]

		if !matchPattern(rule.Match, importPath) {
			continue
		}

		length := specificity(rule.Match)

		if matched == nil || length > matchedLength {
			matched = rule
			matchedLength = length
		}
	}

	return matched, matched != nil
}

func matchPattern(pattern, value string) bool {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(value, prefix)
	}

	return pattern == value
}

func specificity(pattern string) int {
	return len(strings.TrimSuffix(pattern, "*"))
}
