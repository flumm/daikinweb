package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Units  map[string]string
	WebDir string
	Port   int
}

func LoadConfig(FileName string) *Config {
	var data = new(Config)

	f, err := os.Open(FileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open config, loading defaults...: ", err)
		goto defaults
	}

	err = json.NewDecoder(f).Decode(&data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse config, loading defaults...: ", err)
		goto defaults
	}

defaults:
	if data.WebDir == "" {
		data.WebDir = "./www/"
	}

	if data.Port == 0 {
		data.Port = 8080
	}
	return data
}
