package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Tasks    []Task       `yaml:"tasks"`
	Telegram TelegramConf `yaml:"telegram"`
}

type Task struct {
	ID                string           `yaml:"id"`
	Description       string           `yaml:"description"`
	Command           string           `yaml:"command"`
	FrequensySec      int              `yaml:"frequency_sec"`
	Notify            NotificationConf `yaml:"notify"`
	CWD               string           `yaml:"cwdir"`
	FailConfidence    int              `yaml:"fail_confidence"`
	SuccessConfidence int              `yaml:"success_confidence"`
}

type NotificationConf struct {
	Telegram bool `yaml:"telegram"`
	InfluxDB bool `yaml:"influxdb"`
}

type TelegramConf struct {
	Enabled           bool     `yaml:"enabled"`
	Token             string   `yaml:"token"`
	Recipients        []string `yaml:"recipients"`
	FailConfidence    int      `yaml:"fail_confidence"`
	SuccessConfidence int      `yaml:"success_confidence"`
}

func parseConfig(path string) (*Config, error) {
	config := Config{}
	yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := yaml.UnmarshalStrict(yamlData, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
