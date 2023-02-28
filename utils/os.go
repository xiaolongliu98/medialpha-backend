package utils

import (
	"encoding/base64"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func GetProjectRoot() string {
	dir := getCurrentAbPath()
	if strings.HasSuffix(dir, "utils") {
		// run 模式
		dir = filepath.Dir(dir)
	}
	return dir
}

// 最终方案-全兼容
func getCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	tmpDir, _ := filepath.EvalSymlinks(os.TempDir())
	if strings.Contains(dir, tmpDir) {
		return getCurrentAbPathByCaller()
	}
	return dir
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		abPath = path.Dir(filename)
	}
	return abPath
}

func FormatPathAbs(path string) string {
	sysType := runtime.GOOS // "linux" , "windows"
	if sysType != "windows" {
		return FormatUnixPathAbs(path)
	}
	return FormatWindowsPathAbs(path)
}

func FormatUnixPathAbs(path string) string {
	if path == "" {
		return "/"
	}
	path = strings.Trim(path, " ")
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	splits := strings.Split(path, "/")
	// "a", ""
	if len(splits) == 1 {
		return "/" + path
	}

	// "/a/b/c/", "/a/", "/a/b", "a/b/c", "a/"
	path = ""
	for _, each := range splits {
		if each == "" {
			continue
		}

		path += "/" + each
	}
	if path == "" || path == "/" {
		return "/"
	}
	return path
}

func FormatWindowsPathAbs(path string) string {
	// For Example: "D:/", "", ""
	if path == "" {
		return ""
	}

	path = strings.Trim(path, " ")
	path = strings.ReplaceAll(path, "\\", "/")
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	splits := strings.Split(path, "/")
	if len(splits) == 1 {
		return path
	}

	path = ""
	for _, each := range splits {
		if each == "" {
			continue
		}

		path += "/" + each
	}
	if path == "" || path == "/" {
		return ""
	}

	return path[1:]
}

func PathBase(path string) string {
	path = FormatPathAbs(path)
	idx := strings.LastIndex(path, "/")
	if idx == -1 {
		return path
	}
	return path[idx+1:]
}

func AppendPath(path, adder string) string {
	if path == "" || path == "/" {
		if runtime.GOOS == "windows" {
			return adder
		}
		return "/" + adder
	}
	if strings.HasSuffix(path, "/") {
		return path + adder
	}
	for strings.HasPrefix(adder, "/") {
		adder = adder[1:]
	}
	return path + "/" + adder
}

func LogError(err error) bool {
	if err != nil {
		log.Println(err)
		return true
	}
	return false
}

func Base64Encode(src string, replaceBackslash bool) string {
	res := base64.StdEncoding.EncodeToString([]byte(src))
	if replaceBackslash {

		return strings.ReplaceAll(res, "/", "-")
	}
	return res
}

func Base64Decode(src string, replaceBackslash bool, space2plus bool) string {
	if replaceBackslash {
		src = strings.ReplaceAll(src, "-", "/")
	}
	if space2plus {
		src = strings.ReplaceAll(src, " ", "+")
	}
	resBytes, _ := base64.StdEncoding.DecodeString(src)
	return string(resBytes)
}

// 请确保传入的2个path format过
func PathPreContains(src1, src2 string) bool {
	src1Split := strings.Split(src1, "/")
	src2Split := strings.Split(src2, "/")

	i := len(src1Split)
	if len(src2Split) < i {
		i = len(src2Split)
	}

	for j := 0; j < i; j++ {
		if src1Split[j] != src2Split[j] {
			return false
		}
	}

	return true
}

func NumDirs(location string) int {
	es, err := os.ReadDir(location)
	if err != nil {
		return 0
	}
	count := 1
	for _, e := range es {
		if e.IsDir() {
			count += NumDirs(AppendPath(location, e.Name()))
		}
	}
	return count
}

// return dirList, videoList, error
func ReadDirForVideos(location string, handleDir, handleVideo bool) ([]string, []string, error) {
	entries, err := os.ReadDir(location)
	if err != nil {
		return nil, nil, err
	}
	var videoList []string
	var dirList []string

	for _, e := range entries {
		if e.IsDir() {
			if handleDir {
				dirList = append(dirList, e.Name())
			}
			continue
		}

		if !handleVideo {
			continue
		}

		if !IsVideoFile(location + "/" + e.Name()) {
			continue
		}
		videoList = append(videoList, e.Name())
	}
	return dirList, videoList, nil
}
