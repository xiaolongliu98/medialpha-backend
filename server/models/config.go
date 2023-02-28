package models

import (
	"encoding/json"
	"os"
)

type Handler struct {
	Method      string `json:"method"`
	Uri         string `json:"uri"`
	HandlerFunc string `json:"handlerFunc"`
}

type HandlerConfig struct {
	BaseURI  string     `json:"baseURI"`
	Struct   string     `json:"struct"`
	Handlers []*Handler `json:"handlers"`
}

func FromJson(filename string) (HandlerConfigs, error) {
	// 读取配置文件
	jsonBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var configs []*HandlerConfig

	err = json.Unmarshal(jsonBytes, &configs)
	//for _, config := range configs {
	//	sort.Slice(config.Handlers, func(i, j int) bool {
	//		if config.Handlers[i].Struct == config.Handlers[j].Struct {
	//			return config.Handlers[i].HandlerFunc < config.Handlers[j].HandlerFunc
	//		}
	//		return config.Handlers[i].Struct < config.Handlers[j].Struct
	//	})
	//}
	return configs, err
}

type HandlerConfigs []*HandlerConfig

func (h HandlerConfigs) SaveJson(filename string) error {
	jsonBytes, _ := json.Marshal(&h)
	err := os.WriteFile(filename, jsonBytes, os.ModePerm)
	return err
}
