package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"medialpha-backend/server/middlewares"
	"medialpha-backend/server/models"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var enableCros = false

func RegisterHandlerFunc(method, pattern string, handlerFunc http.HandlerFunc) {
	log.Printf("[注册Handler] %v [%v]\n", pattern, method)
	handleFunc := func(w http.ResponseWriter, req *http.Request) {
		if method != req.Method {
			http.NotFound(w, req)
			return
		}
		t := time.Now()
		handlerFunc(w, req)
		d := time.Now().Sub(t)
		log.Printf("[已匹配]time: %.2fms, pattern: %v\n", float64(d.Microseconds())/1000., pattern)
	}
	if enableCros {
		handleFunc = middlewares.CrosMiddleware(handleFunc)
	}

	http.HandleFunc(pattern, handleFunc)
}

func Start(ctx context.Context, serviceName, ip string, port int) {
	ctx, cancelFunc := context.WithCancel(ctx)

	var server http.Server
	server.Addr = ip + ":" + strconv.Itoa(port)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err.Error())
		}
		cancelFunc()
	}()

	go func() {

		log.Printf("%v已启动，按下任意键停止服务器.\n", serviceName)
		var s string
		fmt.Scanln(&s)
		server.Shutdown(ctx)
		cancelFunc()
	}()

	<-ctx.Done()
}

/*
*
JSON Config File For Example:

-----------------------------

[

		{
		  "baseURI": "/v1",
	      "struct": "VideoController",
		  "handlers": [
		    {"handlerFunc": "GetVideos" ,"method": "GET", "uri": "/videos"}
		  ]
		}

]
*/
func RegisterHandlersFromJSON(jsonFilename string, handlerStructs ...any) error {
	configs, err := models.FromJson(jsonFilename)
	if err != nil {
		return err
	}
	registerConfigs(configs, handlerStructs...)
	return nil
}

func registerConfigs(configs models.HandlerConfigs, handlerStructs ...any) {
	m := map[string][]map[string]string{}

	for _, config := range configs {
		for _, handler := range config.Handlers {
			pattern := config.BaseURI + "/" + strings.TrimLeft(handler.Uri, "/ ")
			handlerInfo := map[string]string{
				"FuncName": handler.HandlerFunc,
				"Method":   handler.Method,
				"Pattern":  pattern,
			}
			m[config.Struct] = append(m[config.Struct], handlerInfo)
		}
	}

	for _, handlerStruct := range handlerStructs {
		val := reflect.ValueOf(handlerStruct)

		structName := val.String()[len("<*"):]
		structName = structName[:len(structName)-len(" Value>")]
		if lastIdx := strings.LastIndex(structName, "."); lastIdx != -1 {
			structName = structName[lastIdx+1:]
		}
		handlers, exists := m[structName]
		if !exists {
			log.Printf("Struct[%v]未找到\n", structName)
			continue
		}
		for _, handlerInfo := range handlers {
			method := val.MethodByName(handlerInfo["FuncName"])
			handlerFunc, ok := method.Interface().(func(http.ResponseWriter, *http.Request))
			if !ok {
				log.Printf("HandleFunc[%v]错误\n", handlerInfo["FuncName"])
				continue
			}
			RegisterHandlerFunc(handlerInfo["Method"], handlerInfo["Pattern"], handlerFunc)
		}
	}
}

func handlerScan(location string) (models.HandlerConfigs, error) {
	entries, err := os.ReadDir(location)
	if err != nil {
		return nil, err
	}

	var subdirs []string

	/**
	state 状态机
	-1 - Package包错误
	0 - 该行：监视base struct与handler func
	1 - 该行：是base struct
	2 - 该行：是handler func
	*/
	stateCode := 0
	stateArgs := map[string]string{
		"BaseURI":           "",
		"StructName":        "",
		"HandlerFuncURI":    "",
		"HandlerFuncMethod": "GET",
	}

	var handlers []*models.Handler

	for _, e := range entries {
		if e.IsDir() {
			subdirs = append(subdirs, e.Name())
			continue
		}
		if stateCode == -1 {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".go") {
			continue
		}

		// for current package go files
		data, err := os.ReadFile(location + "/" + e.Name())
		if err != nil {
			log.Printf("sacn go package error: %v\n", err)
			stateCode = -1
			continue
		}

		buf := bytes.NewBuffer(data)

		var line string
		for {
			line, err = buf.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.Trim(line, "\r\n ")

			if stateCode == 1 {
				// 该行：是base struct
				// 解析struct
				if strings.Contains(line, "type ") && strings.Contains(line, " struct") {
					stateArgs["StructName"] = line[len("type "):strings.LastIndex(line, " struct")]
				}
				stateCode = 0

			} else if stateCode == 2 {
				// 该行：是handler func
				// 解析func
				// func (VideoHandlerSet) GetVersion(w http.ResponseWriter, req *http.Request) {
				if strings.Contains(line, "http.ResponseWriter") &&
					strings.Contains(line, "*http.Request") &&
					strings.HasPrefix(line, "func (") {

					idx := strings.Index(line, ") ")
					line = line[idx+len(") "):]
					idx = strings.Index(line, "(")
					line = line[:idx]
					handler := &models.Handler{
						Method:      stateArgs["HandlerFuncMethod"],
						Uri:         stateArgs["HandlerFuncURI"],
						HandlerFunc: line,
					}
					handlers = append(handlers, handler)
				}
				stateCode = 0
			} else if strings.Contains(line, "@base") {
				// 解析base注解
				// '// @base /v1'
				if idx := strings.LastIndex(line, " "); !strings.HasSuffix(line, "@base") && idx != -1 {
					//fmt.Printf("'%v', '%v'\n", line, line[idx+1:])
					stateArgs["BaseURI"] = line[idx+1:]
				}
				stateCode = 1
			} else if strings.Contains(line, "@router") {
				// 解析router注解
				// '// @router /videos GET'
				split := strings.Split(line, " ")
				if len(split) == 4 {
					stateArgs["HandlerFuncURI"] = split[2]
					stateArgs["HandlerFuncMethod"] = strings.ToUpper(split[3])
					stateCode = 2
				} else {
					log.Printf("parse @router error: [%v]\n", line)
					stateCode = 0
				}
			}

		}

		if err != io.EOF {
			log.Printf("sacn go package error: %v\n", err)
			stateCode = -1
			continue
		}
	}

	var configs []*models.HandlerConfig

	if stateArgs["StructName"] != "" && len(handlers) != 0 {
		config := &models.HandlerConfig{
			BaseURI:  stateArgs["BaseURI"],
			Struct:   stateArgs["StructName"],
			Handlers: handlers,
		}
		configs = append(configs, config)
	} else if stateArgs["StructName"] == "" && len(handlers) != 0 {
		log.Printf("scan error: not found struct(%v)", location)
	}

	for _, name := range subdirs {
		subConfig, err := handlerScan(location + "/" + name)
		if err != nil {
			log.Printf("scan dir error: %v", err)
			continue
		}

		configs = append(configs, subConfig...)
	}

	return configs, nil
}

func AutoScanHandlers(controllerLocation string, jsonOutLocation string, handlerStructs ...any) error {
	controllerLocation = strings.TrimRight(controllerLocation, "/\\ ")
	configs, err := handlerScan(controllerLocation)
	if err != nil {
		configs, err = models.FromJson(jsonOutLocation + "/controllers.json")
		if err != nil {
			return err
		}
	} else {
		err := configs.SaveJson(jsonOutLocation + "/controllers.json")
		if err != nil {
			return err
		}
	}
	registerConfigs(configs, handlerStructs...)
	return nil
}

func EnableCros() {
	enableCros = true
}
