package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Server represents http server configuration options.
type Server struct {
	Port                  int           `yaml:"port"`
	ReadTimeoutSecs       time.Duration `yaml:"read_timeout_secs"`
	ReadHeaderTimeoutSecs time.Duration `yaml:"read_header_timeout_secs"`
	IdleTimeoutSecs       time.Duration `yaml:"idle_timeout_secs"`
	WriteTimeoutSecs      time.Duration `yaml:"write_timeout_secs"`
	ShutdownGraceSecs     time.Duration `yaml:"shutdown_grace_secs"`
}

// Database represents postgres database configuration options.
type Database struct {
	Address             string        `yaml:"address"`
	MaxIdleConns        int           `yaml:"max_idle_conns"`
	MaxOpenConns        int           `yaml:"max_open_conns"`
	ConnMaxLifetimeSecs time.Duration `yaml:"conn_max_lifetime_secs"`
}

// Sessions represents session store configuration options.
type Sessions struct {
	Redis             string        `yaml:"redis"`
	SessionExpiryMins time.Duration `yaml:"session_expiry_mins"`
	UserExpiryMins    time.Duration `yaml:"user_expiry_mins"`
}

// Config represents the server configuration options.
type Config struct {
	Server   Server   `yaml:"server"`
	Database Database `yaml:"database"`
	Sessions Sessions `yaml:"sessions"`
}

// Load attempts to load the app configuration from the file located at the provided path.
func Load(path string) (*Config, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
