package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Config defines a LedgerSQL sync configuration
type Config struct {
	DataDir  string `json:"data_dir"`
	DataFile string `json:"data_file"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// BindString returns the bind string for net/http
func (c Config) BindString() string {
	port := 8080
	if c.Port != 0 {
		port = c.Port
	}
	return fmt.Sprintf("%s:%d", c.Host, port)
}

// NewConfig loads a config from a given file or the default location
func NewConfig(file string) (Config, error) {
	var c Config
	var err error

	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return c, err
	}

	err = yaml.Unmarshal(contents, &c)
	return c, err
}
