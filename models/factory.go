package models

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"medialpha-backend/models/config"
	"medialpha-backend/models/dir"
	"medialpha-backend/models/task"
	"medialpha-backend/models/video"
	"medialpha-backend/utils"
	"os"
	"path/filepath"
	"sync"
)

var (
	once        sync.Once
	DB          *gorm.DB
	TaskHandler *task.TaskHandler

	Info  *log.Logger
	Debug *log.Logger
	Warn  *log.Logger
	Err   *log.Logger
)

func init() {
	once.Do(func() {
		os.MkdirAll(config.Config.GetCoverLocationAbs(), os.ModeDir)
		os.MkdirAll(filepath.Dir(config.Config.LogFilename), os.ModeDir)

		initDataBase()
		TaskHandler = task.NewTaskHandler()
		initLog()
	})
}

func initDataBase() {
	db, err := gorm.Open(sqlite.Open(utils.GetProjectRoot()+"/medialpha.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	if config.Config.SqlDebug {
		DB = db.Debug()
	} else {
		DB = db
	}

	err = db.AutoMigrate(video.Instance)
	utils.LogError(err)
	err = db.AutoMigrate(dir.Instance)
	utils.LogError(err)
}

func initLog() {
	logFile, err := os.OpenFile(config.Config.LogFilename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err.Error())
	}

	Info = log.New(io.Discard, "[INFO]", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(io.Discard, "[DEBUG]", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(io.Discard, "[WARN]", log.Ldate|log.Ltime|log.Lshortfile)
	Err = log.New(io.Discard, "[ERROR]", log.Ldate|log.Ltime|log.Lshortfile)
	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// debug > info > warn > err > nothing
	switch config.Config.LogLevel {
	case "debug":
		Info.SetOutput(logFile)
		Debug.SetOutput(logFile)
		Warn.SetOutput(logFile)
		Err.SetOutput(logFile)
	case "warn":
		Warn.SetOutput(logFile)
		Err.SetOutput(logFile)
	case "err":
		Err.SetOutput(logFile)
	case "info":
		Info.SetOutput(logFile)
		Warn.SetOutput(logFile)
		Err.SetOutput(logFile)
	default:
		log.Println("[warn] 请注意！您正在使用无日志输出")
	}
}
