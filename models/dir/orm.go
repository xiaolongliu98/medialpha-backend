package dir

import (
	"fmt"
	"gorm.io/gorm"
	"medialpha-backend/utils"
	"strings"
)

func DelByID(db *gorm.DB, id int64) error {
	return db.Delete(Instance, id).Error
}

func DelAll(db *gorm.DB) error {
	res := db.Where("1 = 1").Delete(Instance)
	return res.Error
}

// 未设置parentID的情况下默认设置为root的ID
func Add(db *gorm.DB, dir *Dir, autoSetParent bool) error {
	err := dir.CheckAdd()
	if err != nil {
		return err
	}

	if dir.ParentID == 0 && autoSetParent {
		// 默认添加到/下
		//var root *Dir
		root, err := CreateRootAndGet(db)
		if err != nil {
			return err
		}
		dir.ParentID = root.ID
	}

	dir.ID = utils.GetSnowFlakeId()
	res := db.Create(dir)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func GetByID(db *gorm.DB, id int64) (*Dir, error) {
	v := &Dir{ID: id}
	res := db.First(v)
	//db.Find(v)
	if res.Error != nil {
		return nil, res.Error
	}
	return v, nil
}

func GetByLocationSelect(db *gorm.DB, location string, cols ...string) (*Dir, error) {
	if !utils.S("id").In(cols...) {
		cols = append(cols, "id")
	}
	d := &Dir{}
	res := db.Select(cols).Where("location = ?", location).Find(d)
	if res.Error != nil || d.ID == 0 {
		return nil, fmt.Errorf("路径不存在")
	}
	return d, nil
}

func GetByParentIDPage(db *gorm.DB, parentID int64, page, pageSize int) ([]*Dir, error) {
	var d []*Dir
	res := db.
		Where("parent_id = ?", parentID).
		Order("name").
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&d)

	if res.Error != nil {
		return nil, res.Error
	}
	return d, nil
}

// []*Dir, total int, error
func GetLikesNamePage(db *gorm.DB, key string, page, pageSize int) ([]*Dir, int, error) {
	var dirs []*Dir
	var total int64
	dbTemp := db.Model(Instance).Where("name LIKE ?", "%"+key+"%")
	res := dbTemp.Count(&total)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	res = dbTemp.
		Order("name").
		Offset(page * pageSize).
		Limit(pageSize).
		Find(&dirs)
	if res.Error != nil {
		return nil, 0, res.Error
	}

	return dirs, int(total), nil
}

func UpdateByID(db *gorm.DB, d *Dir, updateAll bool, cols ...string) error {
	if updateAll {
		return db.Save(d).Error
	}
	if len(cols) == 0 {
		return fmt.Errorf("未指定更新字段")
	}
	res := db.Model(d).Select(cols).Updates(d)
	return res.Error
}

func DelByPrefixLocation(db *gorm.DB, preLocation string) (int, error) {
	res := db.Where("location LIKE ?", preLocation+"%").Delete(Instance)
	rows := res.RowsAffected
	return int(rows), res.Error
}

func CreateRootAndGet(db *gorm.DB) (*Dir, error) {
	d := &Dir{}
	res := db.Where("parent_id = 0").First(d)
	if res.Error == nil {
		return d, nil
	}
	if !strings.Contains(res.Error.Error(), "not found") {
		return nil, res.Error
	}
	// create
	d.ID = utils.GetSnowFlakeId()
	d.Name = ""
	d.ParentID = 0
	d.Location = ""
	res = db.Create(d)
	return d, res.Error
}
