package utils

import (
	"fmt"
	"medialpha-backend/models/config"
	"strings"
)

func ToLocalPath(path string) (string, error) {
	if path == "/" || path == "" {
		return "", nil
	}
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	// '/Videos'(virtual) -> 'D:/c/d/Videos'(local base)
	// For example: '/Videos/a/b' -> 'D:/c/d/Videos/a/b'
	// 1 - 提取base '/Videos'
	prefix := path
	if strings.Count(path, "/") > 1 {
		prefix = prefix[:1+strings.Index(path[1:], "/")]
	}

	localPath, exists := config.Config.PathMapper[prefix]
	if !exists {
		return "", fmt.Errorf("未找到该地址")
	}
	path = path[len(prefix):]
	return localPath + path, nil
	// 2 - 检查Bases中是否存在以此为后缀的
	//targetBase := ""
	//for _, base := range bases {
	//	if strings.HasSuffix(base, prefix) {
	//		targetBase = base
	//		break
	//	}
	//}
	//
	//if targetBase == "" {
	//	return "", fmt.Errorf("地址非法")
	//}
	//// 3 - 将path前缀替换为target base
	//path = path[len(prefix):]
	//return targetBase + path, nil
}

func ToVirtualPath(path string) (string, error) {
	if path == "" {
		return "/", nil
	}
	// '/Videos'(virtual) <- 'D:/c/d/Videos'(local base)
	// For example: '/Videos/a/b' <- 'D:/c/d/Videos/a/b'
	for k, v := range config.Config.PathMapper {
		if strings.HasPrefix(path, k) {
			return v + path[len(k):], nil
		}
	}

	return "", fmt.Errorf("未找到该地址")
}
