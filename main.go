package main

import (
	"context"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"medialpha-backend/controllers/commons"
	"medialpha-backend/controllers/dirs"
	"medialpha-backend/controllers/videos"
	"medialpha-backend/models/config"
	"medialpha-backend/routines"
	"medialpha-backend/server"
	"medialpha-backend/utils"
	"net/http"
)

func main() {
	//
	fmt.Printf("欢迎使用Medialpha v1.0\n - 请访问：%v\n", config.Config.GetURL())
	server.EnableCros()
	// 注册前端页面
	http.Handle("/", http.FileServer(http.Dir(utils.GetProjectRoot()+"/dist")))

	err := server.AutoScanHandlers(
		utils.GetProjectRoot()+"/controllers",
		utils.GetProjectRoot()+"/configs",
		&videos.VideoHandlerSet{},
		&commons.CommonHandlerSet{},
		&dirs.LocationHandlerSet{})

	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	go routines.WorkLoop(ctx)
	server.Start(ctx, "Medialpha", config.Config.IP, config.Config.Port)

}
