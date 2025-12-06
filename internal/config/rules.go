package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Service string `yaml:"service"`
	Level   string `yaml:"level"`
	Issue   string `yaml:"issue"`
}

type RulesConfig struct {
	Rules map[string]Rule `yaml:"rules"`
}

func LoadRules() RulesConfig {
	data, err := os.ReadFile("rules.yaml")
	if err != nil {
		log.Fatalf("Failed to read rules.yaml: %v", err)
	}

	var rules RulesConfig
	if err := yaml.Unmarshal(data, &rules); err != nil {
		log.Fatalf("Failed to parse rules.yaml: %v", err)
	}

	return rules
}

// Get safely retrieves a rule by key. If missing, allows fallback or fatal exit.
// Currently logging fatal to ensure configuration consistency.
func (r RulesConfig) Get(key string) Rule {
	rule, ok := r.Rules[key]
	if !ok {
		log.Fatalf("Rule not found for key: %s", key)
	}
	return rule
}
