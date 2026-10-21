package main

import (
	"encoding/json"
	"io"
	"os"
)

type Config struct {
	Entries         []ConfigEntry `json:"entries"`
	TimeoutSeconds  int32         `json:"timeoutSec"`
	IntervalSeconds int32         `json:"intervalSec"`
}

type ConfigEntry struct {
	Url string `json:"url"`
}

func read_config(path string) (*Config, error) {
	// open file
	jsonFile, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	// read bytes from file
	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		return nil, err
	}

	var config Config

	// parse json
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
