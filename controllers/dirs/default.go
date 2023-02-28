package dirs

import (
	"medialpha-backend/constant"
	SvcDirs "medialpha-backend/services/dirs"
	"medialpha-backend/utils"
	"net/http"
)

// @base /v1
type LocationHandlerSet struct{}

// @query page, path optional
// @router /dir/list GET
func (h LocationHandlerSet) ListDirs(w http.ResponseWriter, req *http.Request) {
	page, _ := utils.GetQueryInt(req, "page")
	path, _ := utils.GetQuery(req, "path")
	path = utils.Base64Decode(path, false, true)

	path = utils.FormatUnixPathAbs(path)

	data, err := SvcDirs.GetDirsByVPath(path, page, constant.Location.PageSize())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(data))
}

// @query page optional, key required
// @router /dir/search GET
func (h LocationHandlerSet) SearchDirs(w http.ResponseWriter, req *http.Request) {
	page, _ := utils.GetQueryInt(req, "page")
	key, _ := utils.GetQuery(req, "key")
	key = utils.Base64Decode(key, false, true)

	data, err := SvcDirs.SearchDirsByName(key, page, constant.Location.PageSize())
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(utils.WrapperError(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(utils.WrapperOK(data))
}
