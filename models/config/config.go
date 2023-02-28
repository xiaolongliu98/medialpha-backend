package config

import (
	"encoding/json"
	"fmt"
	"medialpha-backend/utils"
	"os"
	"sort"
	"strings"
)

type Configuration struct {
	VideoLocations      []string
	IP                  string `json:"ip,omitempty"`
	Port                int
	VideoCoversLocation string
	SqlDebug            bool
	LogFilename         string
	LogLevel            string

	PathMapper map[string]string `json:"-"`
}

func NewDefaultConfig() *Configuration {
	config := &Configuration{
		VideoLocations:      []string{},
		IP:                  "localhost",
		Port:                8081,
		VideoCoversLocation: "./covers",
		SqlDebug:            false,
		LogFilename:         "./logs/medialpha.log",
		LogLevel:            "info",
	}
	return config
}

func FromJsonFile(filename string) (*Configuration, error) {
	// 读取配置文件
	jsonBytes, err := os.ReadFile(filename)
	config := NewDefaultConfig()
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(jsonBytes, config)
	if err != nil {
		return config, err
	}

	if err := config.FormatAndCheck(); err != nil {
		return nil, err
	}

	return config, nil
}

func (config *Configuration) FormatAndCheck() error {
	for i := 0; i < len(config.VideoLocations); i++ {
		config.VideoLocations[i] = utils.FormatPathAbs(config.VideoLocations[i])
	}

	sort.Slice(config.VideoLocations, func(i, j int) bool {
		return config.VideoLocations[i] < config.VideoLocations[j]
	})

	m := map[string]int{}
	config.PathMapper = map[string]string{}

	for i := 0; i < len(config.VideoLocations); i++ {
		base := "/" + utils.PathBase(config.VideoLocations[i])
		count, exists := m[base]
		if exists {
			m[base]++
			base += fmt.Sprintf(" (%v)", count)
		}
		m[base]++
		config.PathMapper[base] = config.VideoLocations[i]
		config.PathMapper[config.VideoLocations[i]] = base
	}

	return config.CheckLegal()
}

func (config *Configuration) CheckLegal() error {
	if config == nil {
		return utils.ErrorNil()
	}
	// 检查视频目录是否重复 和 包含 以及 base重复
	m1 := map[string]struct{}{} // -> 检查目录重复
	for i := 0; i < len(config.VideoLocations); i++ {
		location := config.VideoLocations[i]
		if location == "" {
			return fmt.Errorf("不能为空")
		}

		_, exists := m1[location]
		if exists {
			return fmt.Errorf("目录重复")
		}
		m1[location] = struct{}{}

		if i == 0 {
			continue
		}
		preLocation := config.VideoLocations[i-1]
		if utils.PathPreContains(location, preLocation) {
			return fmt.Errorf("目录存在包含关系！ '%v' and '%v'", location, preLocation)
		}

	}

	return nil
}

func (config *Configuration) Save() error {
	jsonBytes, _ := json.Marshal(config)
	return os.WriteFile(utils.GetProjectRoot()+"/configs/config.json", jsonBytes, os.ModePerm)
}

func (config *Configuration) Copy() *Configuration {
	var copyLocs []string
	for _, loc := range config.VideoLocations {
		copyLocs = append(copyLocs, loc)
	}
	return &Configuration{
		VideoLocations:      copyLocs,
		IP:                  config.IP,
		Port:                config.Port,
		VideoCoversLocation: config.VideoCoversLocation,
		SqlDebug:            config.SqlDebug,
		LogFilename:         config.LogFilename,
		LogLevel:            config.LogLevel,
	}
}

func (config *Configuration) GetCoverLocationAbs() string {
	coverLocSpilt := strings.Split(config.VideoCoversLocation, "/")
	coverLoc := ""
	for i := 0; i < len(coverLocSpilt); i++ {
		if coverLocSpilt[i] == "." {
			coverLoc += utils.GetProjectRoot() + "/"
		} else {
			coverLoc += coverLocSpilt[i] + "/"
		}
	}
	coverLoc = utils.FormatPathAbs(coverLoc)
	return coverLoc
}

func (config *Configuration) GetURL() string {
	return fmt.Sprintf("http://%v:%v", config.IP, config.Port)
}
