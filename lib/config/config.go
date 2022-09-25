package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Database holds data necessary for database configuration
type Database struct {
	Timeout int  `yaml:"timeout_seconds,omitempty"`
	SSLMode bool `yaml:"sslmode,omitempty"`
}

// Server holds data necessary for server configuration
type Server struct {
	Port         string `yaml:"port,omitempty"`
	Debug        bool   `yaml:"debug,omitempty"`
	ReadTimeout  int    `yaml:"read_timeout_seconds,omitempty"`
	WriteTimeout int    `yaml:"write_timeout_seconds,omitempty"`
}

type Configuration struct {
	Server   *Server   `yaml:"server,omitempty"`
	DB       *Database `yaml:"database,omitempty"`
	DataDump *DataDump `yaml:"data_dump"`
}

type DataDump struct {
	FileName string `yaml:"file_name"`
}

// Load returns Configuration struct
func Load(path string) (*Configuration, error) {
	bytes, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("error reading config file, %w", err)
	}

	var cfg = new(Configuration)

	if err := yaml.Unmarshal(bytes, cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %w", err)
	}

	return cfg, nil
}
