package video

import (
	"fmt"
	"gorm.io/gorm"
	"medialpha-backend/utils"
)

func Add(db *gorm.DB, video *Video) error {
	err := video.CheckAdd()
	if err != nil {
		return err
	}

	video.ID = utils.GetSnowFlakeId()
	res := db.Create(video)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func GetByID(db *gorm.DB, id int64) (*Video, error) {
	v := &Video{ID: id}
	res := db.First(v)
	//db.Find(v)
	if res.Error != nil {
		return nil, res.Error
	}
	return v, nil
}

func CountAll(db *gorm.DB) (int, error) {
	var count int64
	res := db.Table(TableName()).Count(&count)

	if res.Error != nil {
		return 0, res.Error
	}
	return int(count), nil
}

func DelAll(db *gorm.DB) error {
	res := db.Where("1 = 1").Delete(Instance)
	return res.Error
}

func DelByID(db *gorm.DB, id int64) error {
	return db.Delete(Instance, id).Error
}

func GetPage(db *gorm.DB, page, pageSize int) ([]*Video, error) {
	var videos []*Video
	res := db.
		Order("last_view_time DESC, name").
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}
	return videos, nil
}

// []*Video, total int, error
func GetLikesNamePage(db *gorm.DB, key string, page, pageSize int) ([]*Video, int, error) {
	var videos []*Video
	var total int64
	dbTemp := db.Model(Instance).Where("name LIKE ?", "%"+key+"%")
	res := dbTemp.Count(&total)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	res = dbTemp.
		Order("last_view_time DESC, name").
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&videos)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	return videos, int(total), nil
}

func GetByLocationPage(db *gorm.DB, location string, page, pageSize int) ([]*Video, error) {
	var videos []*Video
	res := db.
		Where("location = ?", location).
		Order("name").
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&videos)

	if res.Error != nil {
		return nil, res.Error
	}
	return videos, nil
}

func UpdateByID(db *gorm.DB, v *Video, updateAll bool, cols ...string) error {
	if updateAll {
		return db.Save(v).Error
	}
	if len(cols) == 0 {
		return fmt.Errorf("未指定更新字段")
	}
	res := db.Model(v).Select(cols).Updates(v)
	return res.Error
}

func DelByPrefixLocation(db *gorm.DB, prefixLocation string) (int, error) {
	res := db.Model(Instance).Where("location LIKE ?", prefixLocation+"%").Delete(Instance)

	rows := res.RowsAffected
	return int(rows), res.Error
}

func GetByPrefixLocation(db *gorm.DB, prefixLocation string, page, pageSize int) ([]*Video, int, error) {
	var videos []*Video
	var count int64

	dbTemp := db.Model(Instance).Where("location LIKE ?", prefixLocation+"%")
	res := dbTemp.Count(&count)
	if res.Error != nil {
		return nil, 0, res.Error
	}
	res = dbTemp.
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&videos)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	return videos, int(count), nil
}

func CountByPrefixLocation(db *gorm.DB, prefixLocation string) (int, error) {
	var count int64
	res := db.Model(Instance).Where("location LIKE ?", prefixLocation+"%").Count(&count)
	if res.Error != nil {
		return 0, res.Error
	}
	return int(count), nil
}
