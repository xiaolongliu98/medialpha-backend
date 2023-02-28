package videos

import (
	"fmt"
	"log"
	"medialpha-backend/models"
	"medialpha-backend/models/dir"
	"medialpha-backend/models/video"
	SvcUtils "medialpha-backend/services/utils"
	"medialpha-backend/utils"
	"time"
)

func GetVideos(page, pageSize int) (map[string]any, error) {
	count, err := video.CountAll(models.DB)
	if err != nil {
		return nil, err
	}
	videos, err := video.GetPage(models.DB, page, pageSize)
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		"count":  count,
		"videos": utils.Structs2Maps(videos, utils.ID2StrFilter),
	}
	return result, nil
}

func SearchVideosByName(key string, page, pageSize int) (map[string]any, error) {
	videos, count, err := video.GetLikesNamePage(models.DB, key, page, pageSize)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"count":  count,
		"videos": utils.Structs2Maps(videos, utils.ID2StrFilter),
	}, nil
}
func GetVideoByID(id int64) (*video.Video, error) {
	v, err := video.GetByID(models.DB, id)
	if err != nil {
		return nil, err
	}
	v.LastViewTime = int(time.Now().UnixMilli())
	err = video.UpdateByID(models.DB, v, false, "LastViewTime")
	if err != nil {
		log.Println(err)
	}
	return v, nil
}
func GetVideosByVPath(path string, page, pageSize int) (map[string]any, error) {
	localPath, err := SvcUtils.ToLocalPath(path)
	if err != nil {
		return nil, err
	}
	loc, err := dir.GetByLocationSelect(models.DB, localPath, "id", "location", "num_files")

	if err != nil {
		return nil, fmt.Errorf("路径不存在")
	}

	videos, err := video.GetByLocationPage(models.DB, loc.Location, page, pageSize)
	if err != nil {
		return nil, err
	}

	data := map[string]any{
		"count":  loc.NumFiles,
		"videos": utils.Structs2Maps(videos, utils.ID2StrFilter),
	}
	return data, nil
}
