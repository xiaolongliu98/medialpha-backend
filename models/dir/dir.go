package dir

import (
	"fmt"
	"medialpha-backend/constant"
	"medialpha-backend/utils"
	"os"
)

var Instance = &Dir{}

type Dir struct {
	ID         int64  `gorm:"primaryKey; column:id; type:int; not null"`
	Name       string `gorm:"type:string; "` //文件名称（包括后缀）
	ParentID   int64  `gorm:"type:int; "`
	Location   string `gorm:"type:string; index:idx_dir_location, unique;"` //所在目录
	NumFiles   int    `gorm:"type:int; not null;default:0"`
	NumSubDirs int    `gorm:"type:int; not null;default:0"`

	SubDirs   []*Dir `gorm:"-" json:"-"`
	ParentPtr *Dir   `gorm:"-" json:"-"`
}

func (*Dir) TableName() string {
	return TableName()
}

func TableName() string {
	return "dir"
}

func (loc *Dir) CheckAdd() error {
	if loc == nil {
		return utils.ErrorNil()
	}
	if utils.S(loc.Name).ContainsAny("/", "\\") {
		return fmt.Errorf("name名称非法")
	}
	return nil
}

func (loc *Dir) DirDiff() int {
	if loc == nil {
		return constant.Operator.NothingToDo()
	}
	fInfo, err := os.Stat(loc.Location)
	if err != nil || !fInfo.IsDir() {
		return constant.Operator.FileRemoved()
	}
	return constant.Operator.NothingToDo()
}
