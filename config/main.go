package config

import (
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Config defines a LedgerSQL sync configuration
type Config struct {
	DataDir  string `json:"data_dir"`
	DataFile string `json:"data_file"`
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
