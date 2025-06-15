package config

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	Sources map[string]SourceStreamCfg `json:"sources"`
	Outputs map[string]OutputStreamCfg `json:"outputs"`
}

// the command + args to run that we'll read the stdout of as an input source.
type SourceStreamCfg struct {
	Cmd  string   `json:"cmd"`
	Args []string `json:"args,omitempty"`
}

type OutputStreamCfg struct {
	OutputType OutputType     `json:"type"`
	Config     map[string]any `json:"config,omitempty"`
}

func LoadConfigFrom(path string) (Configuration, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Configuration{}, err
	}

	var cfg Configuration
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return Configuration{}, err
	}

	return cfg, nil
}
