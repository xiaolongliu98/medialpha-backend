package commons

import (
	"gorm.io/gorm"
	"medialpha-backend/models"
	"medialpha-backend/models/config"
	"medialpha-backend/models/dir"
	"medialpha-backend/models/video"
	SvcCommons "medialpha-backend/services/commons"
	SvcUnsafe "medialpha-backend/services/unsafe"
	SvcUtils "medialpha-backend/services/utils"
	"medialpha-backend/utils"
	"net/http"
	"os"
)

// @base /v1/common
type CommonHandlerSet struct{}

// GetVersion 获取
// @router /version GET
func (CommonHandlerSet) GetVersion(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK("Medialpha v1.0"))
}

// ReloadAll 删除当前所有视频、目录信息，并且重新加载
// @router /reload/all GET
func (CommonHandlerSet) ReloadAll(w http.ResponseWriter, req *http.Request) {
	err := models.TaskHandler.SubmitTask("ReloadAll")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(""))
}

// @query name, location required
// @router /cover GET
func (CommonHandlerSet) GetCover(w http.ResponseWriter, req *http.Request) {
	// RDovVmlkZW9zL+Wui+a1qS3nur/mgKfku6PmlbA=
	// RDovVmlkZW9zL Wui a1qS3nur/mgKfku6PmlbA=

	name, err := utils.GetQuery(req, "name")
	location, err := utils.GetQuery(req, "location")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	//if name == "34.mp4" {
	//	a := 0
	//	fmt.Println(a)
	//}

	name = utils.Base64Decode(name, false, true)
	location = utils.Base64Decode(location, false, true)

	data, err := SvcCommons.GetCoverBytes(name, location)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// @query id(video id) required
// @router /cover/reload GET
func (CommonHandlerSet) ReloadCover(w http.ResponseWriter, req *http.Request) {
	id, err := utils.GetQueryInt64(req, "id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	v, err := video.GetByID(models.DB, id)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	err = v.GenerateCover()
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(""))
}

// @query clear[*yes, no] str optional
// @router /task GET
func (CommonHandlerSet) CurrentTask(w http.ResponseWriter, req *http.Request) {
	clear, _ := utils.GetQuery(req, "clear")
	clear = utils.WithDefault(clear, "", "yes")

	t := models.TaskHandler.CurrentTaskInfo().ToTaskInfoResp()
	if !t.Running && clear == "yes" {
		models.TaskHandler.ClearTask()
	}
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(t))
}

// @query path[Abs Virtual(Unix) Path] required
// @router /syncdir/recursively GET
func (CommonHandlerSet) SyncDirRecursively(w http.ResponseWriter, req *http.Request) {
	path, err := utils.GetQuery(req, "path")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	path = utils.Base64Decode(path, false, true)

	path = utils.FormatUnixPathAbs(path)
	if path == "/" {
		err = models.TaskHandler.SubmitTask("SyncDirBasesRecursively")
	} else {
		path, err = SvcUtils.ToLocalPath(path)
		if err != nil {
			w.WriteHeader(http.StatusOK)
			w.Write(utils.WrapperError(err.Error()))
			return
		}
		params := &map[string]any{
			"location": path,
		}
		err = models.TaskHandler.SubmitTask("SyncDirRecursively", params)
	}

	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(""))
}

// @router /get/base GET
func (CommonHandlerSet) GetBaseDir(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)

	var data []*map[string]string
	for _, loc := range config.Config.VideoLocations {
		m := &map[string]string{
			"path": loc,
			"name": config.Config.PathMapper[loc][1:],
		}

		data = append(data, m)
	}

	w.Write(utils.WrapperOK(utils.SafeSlice(data)))
}

// @query path[Abs Local Path] required
// @router /add/base GET
func (CommonHandlerSet) AddBaseDir(w http.ResponseWriter, req *http.Request) {
	path, err := utils.GetQuery(req, "path")
	if err != nil || path == "" {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	path = utils.Base64Decode(path, false, true)
	path = utils.FormatPathAbs(path)

	if fInfo, err := os.Stat(path); err != nil || !fInfo.IsDir() {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	cfg := config.Config.Copy()
	cfg.VideoLocations = append(cfg.VideoLocations, path)
	if err := cfg.FormatAndCheck(); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	d := &dir.Dir{
		Name:       utils.PathBase(path),
		Location:   path,
		NumFiles:   0,
		NumSubDirs: 0,
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		err := dir.Add(tx, d, true)
		if err != nil {
			return err
		}

		root, err := dir.CreateRootAndGet(tx)
		if err != nil {
			return err
		}
		root.NumSubDirs++
		err = dir.UpdateByID(tx, root, false, "NumSubDirs")
		if err != nil {
			return err
		}

		err = cfg.Save()
		return err
	})

	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	config.Config = cfg

	w.WriteHeader(http.StatusOK)
	var data []*map[string]string
	for _, loc := range config.Config.VideoLocations {
		m := &map[string]string{
			"path": loc,
			"name": config.Config.PathMapper[loc][1:],
		}
		data = append(data, m)
	}
	w.Write(utils.WrapperOK(utils.SafeSlice(data)))
}

// @query path[Abs Local Path] required
// @router /del/base GET
func (CommonHandlerSet) DelBaseDir(w http.ResponseWriter, req *http.Request) {
	path, err := utils.GetQuery(req, "path")
	if err != nil || path == "" {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	path = utils.Base64Decode(path, false, true)

	idx := 0
	for i, loc := range config.Config.VideoLocations {
		if loc == path {
			idx = i
			break
		}
	}
	if idx >= len(config.Config.VideoLocations) {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError("不存在该目录"))
		return
	}

	cfg := config.Config.Copy()
	cfg.VideoLocations = utils.StrSliceDel(cfg.VideoLocations, idx)

	if err := cfg.FormatAndCheck(); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	//config.Config = config
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	err = models.DB.Transaction(func(tx *gorm.DB) error {
		err := SvcUnsafe.RemoveVideosAndCoversAndSubDirsByPrefixLocation(tx, path)
		if err != nil {
			return err
		}

		root, err := dir.CreateRootAndGet(tx)
		if err != nil {
			return err
		}
		root.NumSubDirs--
		err = dir.UpdateByID(tx, root, false, "NumSubDirs")
		if err != nil {
			return err
		}

		err = cfg.Save()
		return err
	})
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	config.Config = cfg
	w.WriteHeader(http.StatusOK)
	var data []*map[string]string
	for _, loc := range config.Config.VideoLocations {
		m := &map[string]string{
			"path": loc,
			"name": config.Config.PathMapper[loc][1:],
		}
		data = append(data, m)
	}
	w.Write(utils.WrapperOK(utils.SafeSlice(data)))
}
