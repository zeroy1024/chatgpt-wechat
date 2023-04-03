package config

import (
	"encoding/json"
	"os"
	"wechatAPI/chatgpt"
	"wechatAPI/database"
)

var (
	cfg  *Config
	path = "config.json"
)

type Config struct {
	Host               string                     `json:"host"`
	Port               int                        `json:"port"`
	UnofficialProxyAPI chatgpt.UnofficialProxyAPI `json:"unofficial_proxy_api"`
	Database           database.Database          `json:"database"`
	WechatToken        string                     `json:"wechat_token"`
}

func InitConfig() error {
	cfg = new(Config)
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, cfg)
	if err != nil {
		return err
	}

	return nil
}

func GetConfig() *Config {
	return cfg
}
