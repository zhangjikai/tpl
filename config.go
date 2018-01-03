package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	StorePath string
}

func defaultConfig() (Config, error) {
	var config Config
	currentUser, err := user.Current()
	if err != nil {
		return config, err
	}
	config.StorePath = filepath.Join(currentUser.HomeDir, BASE_DIR, DEFAULT_STORAGE_FOLDER)
	return config, nil
}

func configFilePath() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(currentUser.HomeDir, BASE_DIR, CONFIG_FILE), nil
}

func initBaseDir() error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	baseDir := filepath.Join(currentUser.HomeDir, BASE_DIR)
	if !FileExists(baseDir) {
		os.MkdirAll(baseDir, 0755)
	}
	return nil
}

// LoadConfig loads config message from the given file.
func LoadConfig(file string) (Config, error) {
	var config Config

	if !FileExists(file) {
		return defaultConfig()
	}
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

// SaveConfig saves config message to the given file.
func SaveConfig(file string, config Config) error {
	configJson, _ := json.MarshalIndent(config, "", "    ")
	err := ioutil.WriteFile(file, configJson, 0755)
	if err != nil {
		return err
	}
	return nil
}
