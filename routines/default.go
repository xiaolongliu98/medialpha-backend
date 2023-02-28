package routines

import (
	"context"
	"fmt"
	"log"
	"medialpha-backend/models"
	"medialpha-backend/services/unsafe"
	"medialpha-backend/utils"
	"time"
)

func WorkLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case task := <-models.TaskHandler.WaitTask():
			{
				if task == "ReloadAll" {
					//go handleReloadAll()
					handleReloadAll()
				} else if task == "SyncDir" {

				} else if task == "SyncDirRecursively" {
					handleSyncDirRecursively()
				} else if task == "SyncDirBasesRecursively" {
					handleSyncDirBasesRecursively()
				}
			}
		default:
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func handleReloadAll() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()

	t, err := models.TaskHandler.AcceptTask("ReloadAll")
	if utils.LogError(err) {
		return
	}
	defer models.TaskHandler.FinishTask()

	t.Start(6, "清空数据库视频")
	//models.CtxMap.Store("ReloadAll", "working:0/6")

	db := models.DB.Begin()
	err = unsafe.ClearVideoDB(db)
	if err != nil {
		db.Rollback()
		t.Abort("清空数据库视频失败", err)
		//models.CtxMap.Store("ReloadAll", "failed:"+err.Error())
		return
	}
	t.Step("加载视频数据")
	//models.CtxMap.Store("ReloadAll", "working:1/6")
	err = unsafe.LoadVideosIntoDB(db)
	if err != nil {
		db.Rollback()
		t.Abort("加载视频数据失败", err)
		//models.CtxMap.Store("ReloadAll", "failed:"+err.Error())
		return
	}
	t.Step("清空数据库目录")
	//models.CtxMap.Store("ReloadAll", "working:2/6")
	err = unsafe.ClearDirDB(db)
	if err != nil {
		db.Rollback()
		t.Abort("清空数据库目录失败", err)
		//models.CtxMap.Store("ReloadAll", "failed:"+err.Error())
		return
	}
	t.Step("加载目录数据")
	//models.CtxMap.Store("ReloadAll", "working:3/6")
	err = unsafe.LoadDirsIntoDB(db)
	if err != nil {
		db.Rollback()
		t.Abort("加载目录数据失败", err)
		//models.CtxMap.Store("ReloadAll", "failed:"+err.Error())
		return
	}
	t.Step("清空封面目录")
	//models.CtxMap.Store("ReloadAll", "working:4/6")
	err = unsafe.ClearCovers()
	utils.LogError(err)

	t.Step("生成视频封面")
	//models.CtxMap.Store("ReloadAll", "working:5/6")
	err = unsafe.GenerateCovers(db)
	utils.LogError(err)

	db.Commit()
	t.Success()
	//models.CtxMap.Store("ReloadAll", "success")
}

func handleSyncDirRecursively() {
	val, exists := models.TaskHandler.Params()["location"]
	if !exists {
		utils.LogError(fmt.Errorf("SyncDirRecursively Task param 'location'不存在"))
		return
	}
	location, ok := val.(string)
	if !ok {
		utils.LogError(fmt.Errorf("SyncDirRecursively Task param 'location' 类型错误"))
		return
	}
	_, _, _, _, err := unsafe.SyncDirRecursively(location, true)
	if utils.LogError(err) {
		return
	}
}

func handleSyncDirBasesRecursively() {
	_, _, _, _, err := unsafe.SyncDirBasesRecursively()
	if utils.LogError(err) {
		return
	}
}
