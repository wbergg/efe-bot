package config

import (
	"encoding/json"
	"os"
)

type TelegramConfig struct {
	TgAPIKey  string `json:"tgAPIkey"`
	TgChannel string `json:"tgChannel"`
}

type SystembolagetAPI struct {
	Url                       string `json:"url"`
	Ocp_apim_subscription_key string `json:"ocp_apim_subscription_key"`
}

type Config struct {
	Telegram TelegramConfig   `json:"Telegram"`
	SBAPI    SystembolagetAPI `json:"SBAPI"`
}

var Loaded Config

func LoadConfig(filepath string) (Config, error) {
	var c Config

	data, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}

	Loaded = c

	return c, nil
}
