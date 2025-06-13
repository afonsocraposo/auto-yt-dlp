package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var ArchivesDir = "/app/archives"
var DownloadsDir = "/app/downloads"

type Subscription struct {
	URL           string `yaml:"url"`
	Destination   string `yaml:"destination"`
	MaxVideos      int    `yaml:"max_videos"`
	Filter        string `yaml:"filter"`
	Name          string `yaml:"name"`
}

type Config struct {
	Subscriptions []Subscription `yaml:"subscriptions"`
	ArchivesDir   string         `yaml:"archives_dir"`
}

func LoadConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG")
	if configPath == "" {
		configPath = "/app/config/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Println("Config file not found.")
		return &Config{Subscriptions: []Subscription{}}, nil // No config file found, return empty config
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)

	return &config, nil
}
