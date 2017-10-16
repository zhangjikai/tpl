package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

type Config struct {
	StorePath string
}

func defaultConfig() Config {
	var config Config
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err.Error())
	}
	config.StorePath = filepath.Join(currentUser.HomeDir, DEFAULT_STORAGE_FOLDER)
	return config
}

// LoadConfig loads config message from the given file.
func LoadConfig(file string) Config {
	var config Config

	if !FileExists(file) {
		return defaultConfig()
	}
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// SaveConfig saves config message to the given file.
func SaveConfig(file string, config Config) {
	configJson, _ := json.MarshalIndent(config, "", "    ")
	err := ioutil.WriteFile(file, configJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
