package main

import (
	"encoding/json"
	"os"
)

const DEFAULT_CONFIG_FILE string = "subway.json"

type Conf struct {
	Root           string `json:"root"`
	Port           int    `json:"port"`
	TorPath        string `json:"torPath"`
	PortForwarding string `json:"portForwarding"`
}

func DefaultConfig() *Conf {
	return &Conf{
		Root:           "./",
		Port:           8864,
		TorPath:        "tor",
		PortForwarding: "",
	}
}

func LoadConfig(filePath string) (*Conf, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	defer file.Close()

	var config Conf
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Conf, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

func GenerateDefaultConfig(filePath string) error {
	return SaveConfig(DefaultConfig(), filePath)
}
