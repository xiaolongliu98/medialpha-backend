package video

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"medialpha-backend/constant"
	"medialpha-backend/models/config"
	"medialpha-backend/utils"
	"os"
	"strconv"
	"strings"
)

var Instance = &Video{}

type Video struct {
	ID           int64   `gorm:"primaryKey; column: id; type: int; not null"`
	Location     string  `gorm:"type:string; not null; index:idx_video_filename, unique"` //所在目录
	Name         string  `gorm:"type:string; not null; index:idx_video_filename, unique"` //文件名称（包括后缀）
	Size         int     `gorm:"type:int; not null;"`                                     //文件大小
	UpdateTime   int     `gorm:"type:int; not null;"`                                     //修改时间
	Duration     int     `gorm:"type:int; not null;"`                                     // 时常 秒 seconds
	Width        int     `gorm:"type:int; not null;"`                                     //分辨率
	Height       int     `gorm:"type:int; not null;"`                                     //分辨率
	FrameRate    float64 `gorm:"type:int; "`                                              //帧率(Frame Rate)
	VideoSuffix  string  `gorm:"type:string; not null;"`                                  //封装格式(MP4、AVI、MKV、FLV、WMA)
	BitRate      int     `gorm:"type:int; "`                                              //码率(Bit Rate)
	CodeType     string  `gorm:"type:string; "`                                           //编码格式(H264, H265, H261, H263)
	LastViewTime int     `gorm:"type:int; not null; default:0"`                           //上次观看时间 默认未0
}

func (*Video) TableName() string {
	return TableName()
}

func TableName() string {
	return "video"
}

func NewVideo(location, name string) (*Video, error) {
	location = strings.TrimRight(location, "/\\ ")
	target := location + "/" + name

	if info, err := os.Stat(target); !(err == nil && !info.IsDir()) {
		return nil, fmt.Errorf("目标视频文件不存在")
	}

	if !utils.IsVideoFile(name) {
		return nil, fmt.Errorf("目标视频文件格式不正确")
	}

	v := &Video{Location: location, Name: name}
	return v, nil
}

func (v *Video) ReadFromDisk() error {
	if v == nil {
		return fmt.Errorf("空指针异常")
	}
	target := v.GetFilename()
	infos, err := utils.ReadVideoStreamsInfo(target)
	if err != nil {
		return err
	}

	maxDuration := 0.
	totalBitRate := 0
	// duration handler
	for _, info := range infos {
		infoMap := *info
		dStr := utils.WithDefault(infoMap["duration"], "", "0")
		dStr = utils.WithDefault(dStr, "N/A", "0")
		d, err := strconv.ParseFloat(dStr, 64)
		if err == nil {
			maxDuration = math.Max(maxDuration, d)
		}

		bitRateStr := utils.WithDefault(infoMap["bit_rate"], "", "0")
		bitRateStr = utils.WithDefault(bitRateStr, "N/A", "0")
		bitRate64, err := strconv.ParseInt(bitRateStr, 10, 64)
		if err == nil {
			totalBitRate += int(bitRate64)
		}
	}
	v.Duration = int(maxDuration)
	v.BitRate = totalBitRate

	// videos handler
	for _, info := range infos {
		infoMap := *info
		streamType := infoMap["codec_type"]
		if !strings.HasPrefix(streamType, "video") {
			continue
		}
		// codeType
		v.CodeType = infoMap["codec_name"]
		// w,h
		widthStr := utils.WithDefault(infoMap["width"], "N/A", "0")
		widthStr = utils.WithDefault(widthStr, "", "0")
		width64, err1 := strconv.ParseInt(widthStr, 10, 64)
		if err1 == nil && v.Width == 0 {
			v.Width = int(width64)
		}

		heightStr := utils.WithDefault(infoMap["height"], "N/A", "0")
		heightStr = utils.WithDefault(heightStr, "", "0")
		height64, err2 := strconv.ParseInt(heightStr, 10, 64)
		if err2 == nil && v.Height == 0 {
			v.Height = int(height64)
		}

		frameRateStr := utils.WithDefault(infoMap["r_frame_rate"], "N/A", "0/1")
		frameRateStr = utils.WithDefault(frameRateStr, "", "0/1")

		split := strings.Split(frameRateStr, "/")
		a, err3 := strconv.ParseFloat(split[0], 64)
		b, err4 := strconv.ParseFloat(split[1], 64)
		if err3 == nil && err4 == nil && b != 0. && v.FrameRate == 0. {
			a, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", a/b), 64)
			v.FrameRate = a
		}

		err := utils.ErrorsOf(err1, err2, err3, err4)
		if err != nil {
			break
		}
	}

	fInfo, err := os.Stat(target)
	if err != err {
		return err
	}

	v.UpdateTime = int(fInfo.ModTime().UnixMilli())
	v.Size = int(fInfo.Size())
	if dotPos := strings.LastIndex(v.Name, "."); dotPos == -1 {
		return fmt.Errorf("目标视频文件格式不正确")
	} else {
		v.VideoSuffix = v.Name[dotPos:]
	}
	return nil
}

func (v *Video) Print() {
	if v == nil {
		return
	}
	bytes, _ := json.Marshal(v)
	var m map[string]any
	json.Unmarshal(bytes, &m)
	fmt.Println(m)
}

func (v *Video) GetFilename() string {
	return v.Location + "/" + v.Name
}

func (v *Video) FileExists() (bool, error) {
	if v == nil {
		return false, utils.ErrorNil()
	}
	info, err := os.Stat(v.GetFilename())
	return err == nil && !info.IsDir(), err
}

func (v *Video) CheckAdd() error {
	if v == nil {
		return utils.ErrorNil()
	}

	if v.Size == 0 {
		return fmt.Errorf("size为空")
	}
	if len(v.Name) == 0 {
		return fmt.Errorf("name为空")
	}
	if !utils.IsVideoFile(v.Name) {
		return fmt.Errorf("视频文件格式不符")
	}

	if utils.S(v.VideoSuffix).NotIn(constant.Video.SuffixList()...) {
		return fmt.Errorf("视频文件格式不符")
	}

	if v.UpdateTime <= 0 {
		return fmt.Errorf("修改时间未设置")
	}

	if v.Width <= 0 || v.Height <= 0 {
		return fmt.Errorf("视频尺寸未设置")
	}

	if v.Duration <= 0 {
		return fmt.Errorf("视频时长未设置")
	}

	if len(v.Location) <= 0 {
		return fmt.Errorf("视频目录未设置")
	}

	return nil
}

func (v *Video) GenerateCover() error {
	if v == nil {
		return utils.ErrorNil()
	}
	return utils.GenerateFrame(v.GetFilename(), v.GetCoverFilename(), v.Duration)
}

func (v *Video) CoverExists() (bool, error) {
	if v == nil {
		return false, utils.ErrorNil()
	}
	info, err := os.Stat(v.GetCoverFilename())
	return err == nil && !info.IsDir(), err
}

func (v *Video) DelCoverFile() error {
	if v == nil {
		return utils.ErrorNil()
	}
	err := os.Remove(v.GetCoverFilename())
	return err
}

func (v *Video) GetCoverFilename() string {
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(v.GetFilename())))
	//md5: 	b0804ec967f48520697662a204f5fe72
	target := fmt.Sprintf("%v/%v_%v.jpg", config.Config.GetCoverLocationAbs(), v.Name, md5Str)
	return target
}

func (v *Video) FileDiff() int {
	if v == nil {
		return constant.Operator.NothingToDo()
	}

	fInfo, err := os.Stat(v.GetFilename())
	if err != nil || fInfo.IsDir() {
		return constant.Operator.FileRemoved()
	}
	updateTime := int(fInfo.ModTime().UnixMilli())
	if updateTime != v.UpdateTime {
		return constant.Operator.FileChanged()
	}
	return constant.Operator.NothingToDo()
}

type VideoLite struct {
	Location   string
	Name       string
	UpdateTime int
}
