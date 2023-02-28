package config

import (
	"medialpha-backend/utils"
	"sync"
)

var (
	once   sync.Once
	Config *Configuration
)

func init() {
	once.Do(func() {
		loadConfig()
	})
}

func loadConfig() {
	// 读取配置文件
	c, err := FromJsonFile(utils.GetProjectRoot() + "/configs/config.json")
	utils.LogError(err)
	Config = c
}
