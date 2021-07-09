package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/zeu5/visualizations/log"
)

type Config struct {
	DBURI      string           `json:"db_uri"`
	ServerAddr string           `json:"addr"`
	Log        log.LoggerConfig `json:"log"`
}

func ParseConfig(path string) (*Config, error) {

	defaultConfig := &Config{
		DBURI:      "mongodb://localhost:27017",
		ServerAddr: "localhost:8080",
		Log:        log.DefaultLoggerConfig,
	}
	if path == "" {
		return defaultConfig, nil
	}
	fileContents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %s", err)
	}

	err = json.Unmarshal(fileContents, defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %s", err)
	}
	return defaultConfig, nil
}
