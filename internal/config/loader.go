package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port   int    `yaml:"port"`
		APIKey string `yaml:"api_key"`
	} `yaml:"server"`
	Commands map[string]string `yaml:"commands"`

	// Credenciales por defecto para Telnet (desde env; no van en YAML)
	DefaultSwitchUser     string `yaml:"-"`
	DefaultSwitchPassword string `yaml:"-"`
}

// LoadConfig lee el archivo YAML y lo mapea a la estructura
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	applySwitchCredentialsFromEnv(&cfg)
	return &cfg, nil
}

func applySwitchCredentialsFromEnv(cfg *Config) {
	if u := os.Getenv("NOC_SWITCH_USER"); u != "" {
		cfg.DefaultSwitchUser = u
	} else {
		cfg.DefaultSwitchUser = "allied"
	}
	if p := os.Getenv("NOC_SWITCH_PASSWORD"); p != "" {
		cfg.DefaultSwitchPassword = p
	} else {
		cfg.DefaultSwitchPassword = "4ll13d"
	}
}
