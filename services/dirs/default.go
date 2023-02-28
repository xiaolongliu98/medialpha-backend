package dirs

import (
	"fmt"
	"medialpha-backend/models"
	"medialpha-backend/models/config"
	"medialpha-backend/models/dir"
	SvcUtils "medialpha-backend/services/utils"
	"medialpha-backend/utils"
)

func GetDirsByVPath(path string, page, pageSize int) (map[string]any, error) {
	localPath, err := SvcUtils.ToLocalPath(path)
	if err != nil {
		return nil, err
	}
	d, err := dir.GetByLocationSelect(models.DB, localPath, "id", "num_sub_dirs")
	if err != nil {
		return nil, fmt.Errorf("路径不存在")
	}
	dirs, err := dir.GetByParentIDPage(models.DB, d.ID, page, pageSize)
	if err != nil {
		return nil, err
	}

	for _, each := range dirs {
		virtualPath, err := SvcUtils.ToVirtualPath(each.Location)
		if err != nil {
			return nil, err
		}
		each.Location = virtualPath
		each.Name = utils.PathBase(virtualPath)
	}

	data := map[string]any{
		"count": d.NumSubDirs,
		"dirs":  utils.Structs2Maps(dirs, utils.ID2StrFilter),
	}

	if path == "/" {
		data["count"] = len(config.Config.VideoLocations)
	}

	return data, nil
}
func SearchDirsByName(key string, page, pageSize int) (map[string]any, error) {
	dirs, count, err := dir.GetLikesNamePage(models.DB, key, page, pageSize)
	if err != nil {
		return nil, err
	}

	// 转换为Virtual Locations
	for _, each := range dirs {
		virtualPath, err := SvcUtils.ToVirtualPath(each.Location)
		if err != nil {
			return nil, err
		}
		each.Location = virtualPath
		each.Name = utils.PathBase(virtualPath)
	}

	return map[string]any{
		"count": count,
		"dirs":  utils.Structs2Maps(dirs, utils.ID2StrFilter),
	}, nil
}
