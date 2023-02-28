package videos

import (
	"fmt"
	"medialpha-backend/constant"
	"medialpha-backend/models"
	"medialpha-backend/models/video"
	SvcVideos "medialpha-backend/services/videos"
	"medialpha-backend/utils"
	"net/http"
	"os"
	"time"
)

// @base /v1
type VideoHandlerSet struct{}

// @query page optional
// @router /videos GET
func (VideoHandlerSet) GetVideos(w http.ResponseWriter, req *http.Request) {
	page, _ := utils.GetQueryInt(req, "page")
	data, err := SvcVideos.GetVideos(page, constant.Video.PageSize())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(data))
}

// @query id required
// @router /video/stream GET
func (VideoHandlerSet) ServeVideoStream(w http.ResponseWriter, req *http.Request) {
	id, err := utils.GetQueryInt64(req, "id")
	//filename, err := utils.GetBody(req, "filename")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.LogError(err)
		return
	}

	v, err := video.GetByID(models.DB, id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		utils.LogError(err)
		return
	}
	if exists, _ := v.FileExists(); !exists {
		v.DelCoverFile()
		video.DelByID(models.DB, v.ID)

		w.WriteHeader(http.StatusInternalServerError)
		utils.LogError(fmt.Errorf("视频文件不存在"))
		return
	}

	filename := v.GetFilename()
	fd, err := os.Open(filename)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.LogError(err)
		return
	}
	defer fd.Close()
	http.ServeContent(w, req, utils.PathBase(filename), time.Now(), fd)
}

// @query id required
// @router /video GET
func (VideoHandlerSet) GetVideo(w http.ResponseWriter, req *http.Request) {
	id, err := utils.GetQueryInt64(req, "id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	v, err := SvcVideos.GetVideoByID(id)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(utils.Struct2Map(v, utils.ID2StrFilter)))
}

// @query key required
// @router /video/search GET
func (VideoHandlerSet) SearchVideos(w http.ResponseWriter, req *http.Request) {
	key, err := utils.GetQuery(req, "key")
	page, _ := utils.GetQueryInt(req, "page")
	key = utils.Base64Decode(key, false, true)

	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	data, err := SvcVideos.SearchVideosByName(key, page, constant.Video.PageSize())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(data))
}

// @query page optional, path required
// @router /dir/videos GET
func (VideoHandlerSet) GetVideosByVPath(w http.ResponseWriter, req *http.Request) {
	page, _ := utils.GetQueryInt(req, "page")
	path, _ := utils.GetQuery(req, "path")
	path = utils.Base64Decode(path, false, true)
	path = utils.FormatUnixPathAbs(path)
	if path == "/" {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperParamIllegal())
		return
	}

	data, err := SvcVideos.GetVideosByVPath(path, page, constant.Video.PageSize())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(data))
}
