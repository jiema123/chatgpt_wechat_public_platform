package config

import (
	"encoding/json"
	"fmt"
	"myapp/pkg/logger"
	"os"
	"sync"
	"time"
)

type Configuration struct {
	AppId     string `json:"app_id"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token"`
	AesKey    string `json:"aes_key"`
	// gpt apikey
	ApiKey string `json:"api_key"`
	// 会话超时时间
	SessionTimeout time.Duration `json:"session_timeout"`
	// GPT请求最大字符数
	MaxTokens uint `json:"max_tokens"`
	// GPT模型
	Model string `json:"model"`
	// 清空会话口令
	SessionClearToken string `json:"session_clear_token"`
	// 规则 system user assistant
	Role string `json:"role"`
	// 画图画口令
	ImageStartKey string `json:"image_start_key"`
	//0-10 的数字
	ImageN int `json:"image_n"`
	//256x256, 512x512, or 1024x1024.
	ImageSize string `json:"image_size"`
}

var config *Configuration
var once sync.Once

// LoadConfig 加载配置
func LoadConfig() *Configuration {
	once.Do(func() {
		// 给配置赋默认值
		config = &Configuration{
			SessionTimeout:    60,
			MaxTokens:         512,
			Model:             "gpt-3.5-turbo-0301",
			Role:              "assistant",
			SessionClearToken: "下个问题",
			ImageStartKey:     "画图画",
			ImageN:            10,
			ImageSize:         "256x256",
		}

		// 判断配置文件是否存在，存在直接JSON读取
		_, err := os.Stat("config.json")
		if err == nil {
			f, err := os.Open("config.json")
			if err != nil {
				logger.Danger(fmt.Sprintf("open config error: %v", err))
				return
			}
			defer f.Close()
			encoder := json.NewDecoder(f)
			err = encoder.Decode(config)
			if err != nil {
				logger.Danger(fmt.Sprintf("decode config error: %v", err))
				return
			}
		}
	})
	if config.ApiKey == "" {
		logger.Danger("config error: api key required")
	}

	return config
}
