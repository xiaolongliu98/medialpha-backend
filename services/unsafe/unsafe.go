package unsafe

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"medialpha-backend/constant"
	"medialpha-backend/models"
	"medialpha-backend/models/config"
	"medialpha-backend/models/dir"
	"medialpha-backend/models/task"
	"medialpha-backend/models/video"
	SvcUtils "medialpha-backend/services/utils"
	"medialpha-backend/utils"
	"os"
	"reflect"
	"strings"
)

func readVideoAll() (map[string][]*video.VideoLite, error) {
	locations := config.Config.VideoLocations
	videoMap := map[string][]*video.VideoLite{}
	for _, loc := range locations {
		loc = strings.TrimRight(loc, "/\\ ")
		videos, err := readVideosFromDirRecursivelyHelper(loc)
		if err != nil {
			continue
		}

		videoMap[loc] = videos
	}
	return videoMap, nil
}

func readVideosFromDirRecursivelyHelper(location string) ([]*video.VideoLite, error) {
	entries, err := os.ReadDir(location)
	if err != nil {
		return nil, err
	}

	var videos []*video.VideoLite
	// DFS
	for _, e := range entries {
		if e.IsDir() {
			vs, err := readVideosFromDirRecursivelyHelper(location + "/" + e.Name())
			if err == nil {
				videos = append(videos, vs...)
			}
			continue
		}
		if !utils.IsVideoFile(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}

		v := &video.VideoLite{
			Location:   location,
			Name:       e.Name(),
			UpdateTime: int(info.ModTime().UnixMilli()),
		}
		videos = append(videos, v)
	}

	return videos, nil
}

func readDirAll() *dir.Dir {
	roots := config.Config.VideoLocations
	rootLoc := &dir.Dir{
		Name:       "",
		Location:   "",
		NumFiles:   0,
		NumSubDirs: 0,
		SubDirs:    nil,
		ParentPtr:  nil,
	}
	for _, root := range roots {
		loc, err := readDirsRecursivelyHelper(root)
		if err != nil {
			continue
		}
		loc.ParentPtr = rootLoc
		rootLoc.SubDirs = append(rootLoc.SubDirs, loc)
	}
	rootLoc.NumSubDirs = len(rootLoc.SubDirs)
	return rootLoc
}

func readDirsRecursivelyHelper(path string) (*dir.Dir, error) {
	// 处理parent信息
	fInfo, err := os.Stat(path)
	if err != nil {
		//log.Printf("[read dir error] %v\n", err)
		return nil, err
	}
	if !fInfo.IsDir() {
		return nil, fmt.Errorf("%v 不是一个目录", path)
	}
	//size := fInfo.Size()
	entries, err := os.ReadDir(path)
	if err != nil {
		//log.Printf("[read dir error] %v\n", err)
		return nil, err
	}

	loc := &dir.Dir{
		Name:       utils.PathBase(path),
		Location:   path,
		NumFiles:   0,
		NumSubDirs: 0,

		SubDirs:   nil,
		ParentPtr: nil,
	}

	for _, e := range entries {
		if e.IsDir() {
			subLocation, err := readDirsRecursivelyHelper(utils.AppendPath(path, e.Name()))

			if err == nil {
				subLocation.ParentPtr = loc
				loc.SubDirs = append(loc.SubDirs, subLocation)
			}
			continue
		}

		if utils.IsVideoFile(path + "/" + e.Name()) {
			loc.NumFiles++
		}
	}

	loc.NumSubDirs = len(loc.SubDirs)
	return loc, nil
}

func ClearVideoDB(db *gorm.DB) error {
	err := video.DelAll(db)
	return err
}

func ClearCovers() error {
	entries, err := os.ReadDir(config.Config.GetCoverLocationAbs())
	if err != nil {
		return err
	}
	//t := time.Now()
	counterAll := 0
	counter := 0
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") && !e.IsDir() {
			counterAll++
			if os.Remove(config.Config.GetCoverLocationAbs()+"/"+e.Name()) == nil {
				counter++
			}
		}
	}
	//diff := float64(time.Now().Sub(t).Milliseconds()) / 1000.
	//log.Printf("耗时:%.2f \n", diff)
	//log.Printf("总共发现了%v个封面，成功删除%v个视频的封面.\n", counterAll, counter)
	return nil
}

func LoadVideosIntoDB(db *gorm.DB) error {
	//t := time.Now()
	videoMap, err := readVideoAll()
	if err != nil {
		return err
	}

	counter := 0
	counterAll := 0
	for _, videos := range videoMap {
		for _, vl := range videos {
			counterAll++
			v, err := video.NewVideo(vl.Location, vl.Name)
			if err != nil {
				continue
			}
			err = v.ReadFromDisk()
			if err != nil {
				continue
			}
			err = video.Add(db, v)
			if err == nil {
				counter++
			} else {
				log.Println(err)
			}

			//v.GenerateCover() //耗时变为原来的2.4X
		}
	}
	//diff := float64(time.Now().Sub(t).Milliseconds()) / 1000.
	//log.Printf("耗时:%.2f \n", diff)
	//log.Printf("总共扫描到%v个视频，成功添加%v个视频.\n", counterAll, counter)
	return nil
}

func GenerateCovers(db *gorm.DB) error {
	//t := time.Now()
	counter := 0
	counterAll := 0

	total, err := video.CountAll(db)
	if err != nil {
		return err
	}
	pageSize := 10
	for page := 0; counterAll < total; page++ {
		videos, err := video.GetPage(db, page, pageSize)

		counterAll += len(videos)

		if err != nil {
			panic(err)
		}
		for _, v := range videos {
			err = v.GenerateCover()
			if err == nil {
				counter++
			}
		}
	}

	//diff := float64(time.Now().Sub(t).Milliseconds()) / 1000.
	//log.Printf("耗时:%.2f \n", diff)
	//log.Printf("总共处理了%v个视频，成功生成%v个视频的封面.\n", counterAll, counter)

	return nil
}

func ClearDirDB(db *gorm.DB) error {
	err := dir.DelAll(db)
	return err
}

func LoadDirsIntoDB(db *gorm.DB) error {
	//t := time.Now()
	parent := readDirAll()

	var Q []*dir.Dir
	Q = append(Q, parent)
	parent = nil

	for len(Q) != 0 {
		size := len(Q)
		for i := 0; i < size; i++ {
			front := Q[0]
			Q = Q[1:]

			if front.ParentPtr != nil {
				front.ParentID = front.ParentPtr.ID
			} else {
				front.ParentID = 0
			}
			err := dir.Add(db, front, false)
			if err != nil {
				continue
			}

			Q = append(Q, front.SubDirs...)
		}
	}
	//diff := float64(time.Now().Sub(t).Milliseconds()) / 1000.
	//log.Printf("耗时:%.2f \n", diff)
	//log.Println("==========================")

	return nil
}

// 删除location路径下所有的video以及其cover
func RemoveVideosAndCoversByPrefixLocation(db *gorm.DB, location string) (int, error) {
	if db == nil || reflect.ValueOf(db).IsNil() {
		db = models.DB
	}
	pageSize := 32

	total, err := video.CountByPrefixLocation(db, location)
	if err != nil {
		return 0, err
	}

	numRemoved := 0
	numPages := total/(1+pageSize) + 1
	for i := numPages - 1; i >= 0; i-- {
		videos, _, err := video.GetByPrefixLocation(db, location, i, pageSize)
		if err != nil {
			return numRemoved, err
		}
		for _, v := range videos {
			err := video.DelByID(db, v.ID)
			if utils.LogError(err) {
				continue
			}
			numRemoved++
			err = v.DelCoverFile()
			utils.LogError(err)
		}
	}

	return numRemoved, nil
}

// 删除location路径下所有的video、dir以及其cover
func RemoveVideosAndCoversAndSubDirsByPrefixLocation(db *gorm.DB, location string) error {
	if db == nil || reflect.ValueOf(db).IsNil() {
		db = models.DB
	}
	pageSize := 32

	_, err := dir.DelByPrefixLocation(db, location)
	if err != nil {
		return err
	}

	total, err := video.CountByPrefixLocation(db, location)
	if err != nil {
		return err
	}

	numPages := total/(1+pageSize) + 1
	for i := numPages - 1; i >= 0; i-- {
		videos, _, err := video.GetByPrefixLocation(db, location, i, pageSize)
		if err != nil {
			return err
		}
		for _, v := range videos {
			err := video.DelByID(db, v.ID)
			if utils.LogError(err) {
				continue
			}
			err = v.DelCoverFile()
			utils.LogError(err)
		}
	}

	return nil
}

func syncVideosByLocationFlat(
	location string,
	oldNumFiles, loadPageSize int,
	videoList []string) (int, int, int) {
	numPages := oldNumFiles/(1+loadPageSize) + 1
	validVideoNames := map[string]struct{}{}
	numVideoRemoved := 0
	numVideoAdded := 0
	var removeList []*video.Video

	// 检查现有数据库中是否存在被删除的
	for i := numPages - 1; i >= 0; i-- {
		videos, err := video.GetByLocationPage(models.DB, location, i, loadPageSize)
		if utils.LogError(err) {
			continue
		}
		for _, v := range videos {
			operator := v.FileDiff()
			switch operator {
			case constant.Operator.NothingToDo():
				validVideoNames[v.Name] = struct{}{}

			case constant.Operator.FileRemoved(), constant.Operator.FileChanged():
				// 有变化就先删除，当作新增处理
				removeList = append(removeList, v)
			}
		}
	}

	// 进行删除
	for _, v := range removeList {
		v.DelCoverFile()
		err := video.DelByID(models.DB, v.ID)
		if !utils.LogError(err) {
			numVideoRemoved++
		}
	}

	// 同步新增视频
	for _, videoName := range videoList {
		//if videoName == "86-不存在的战区 22.mp4" {
		//	a := 0
		//	fmt.Println(a)
		//}
		_, exists := validVideoNames[videoName]
		if exists {
			continue
		}
		v, err := video.NewVideo(location, videoName)
		if utils.LogError(err) {
			continue
		}
		err = v.ReadFromDisk()
		if utils.LogError(err) {
			continue
		}
		err = video.Add(models.DB, v)
		if !utils.LogError(err) {
			if exists, err := v.CoverExists(); !exists || err != nil {
				err2 := v.GenerateCover()
				utils.LogError(err)
				utils.LogError(err2)
			}
			numVideoAdded++
			validVideoNames[videoName] = struct{}{}
		}
	}

	return numVideoAdded, numVideoRemoved, len(validVideoNames)
}

func syncDirsByLocationFlat(
	location string,
	oldNumSubDirs, loadPageSize int,
	dirList []string,
	parentID int64) (int, int, int, map[string]struct{}) {

	numPages := oldNumSubDirs/(loadPageSize+1) + 1
	validDirNames := map[string]struct{}{}
	var removeList []*dir.Dir
	numDirAdded := 0
	numDirRemoved := 0
	numVideoRemoved := 0
	// 同步数据现有的目录，检查是否被删除
	for i := numPages - 1; i >= 0; i-- {
		subDirs, err := dir.GetByParentIDPage(models.DB, parentID, i, loadPageSize)
		if utils.LogError(err) {
			continue
		}
		for _, subDir := range subDirs {
			operator := subDir.DirDiff()
			switch operator {
			case constant.Operator.NothingToDo():
				validDirNames[subDir.Name] = struct{}{}

			case constant.Operator.FileRemoved(), constant.Operator.FileChanged():
				removeList = append(removeList, subDir)
			}
		}
	}

	for _, subDir := range removeList {
		numVideoRemovedSub := 0
		err := models.DB.Transaction(func(tx *gorm.DB) error {
			err := dir.DelByID(models.DB, subDir.ID)
			if err != nil {
				return err
			}
			numVideoRemovedSub, err = RemoveVideosAndCoversByPrefixLocation(tx, subDir.Location)
			return err
		})
		if !utils.LogError(err) {
			numDirRemoved++
			numVideoRemoved += numVideoRemovedSub
		}
	}

	// 同步新增目录
	for _, dirName := range dirList {
		if _, exists := validDirNames[dirName]; exists {
			continue
		}
		newDir := &dir.Dir{
			ID:         0,
			Name:       dirName,
			ParentID:   parentID,
			Location:   utils.AppendPath(location, dirName),
			NumFiles:   0,
			NumSubDirs: 0,
		}
		err := dir.Add(models.DB, newDir, true)
		if !utils.LogError(err) {
			validDirNames[dirName] = struct{}{}
			numDirAdded++
		}
	}

	return numDirAdded, numDirRemoved, numVideoRemoved, validDirNames
}

func syncDirRecursivelyHelper(location string) (int, int, int, int, error) {
	t := models.TaskHandler.CurrentTaskInfo()
	vPath, _ := SvcUtils.ToVirtualPath(location)
	t.Step("扫描" + vPath)

	dirList, videoList, err := utils.ReadDirForVideos(location, true, true)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	numDirAdded := 0
	numDirRemoved := 0
	numVideoAdded := 0
	numVideoRemoved := 0
	pageSize := 32

	// 读取父目录基本信息
	parentDir, err := dir.GetByLocationSelect(models.DB, location, "id", "NumFiles", "NumSubDirs")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	oldNumSubDirs, oldNumFiles := parentDir.NumSubDirs, parentDir.NumFiles
	// 同步所有子目录信息
	dAdds, dRemoves, vRemoves, validDirNames :=
		syncDirsByLocationFlat(location, oldNumSubDirs, pageSize, dirList, parentDir.ID)
	numDirAdded += dAdds
	numDirRemoved += dRemoves
	numVideoRemoved += vRemoves

	// 递归地进行(对有效的目录递归)
	for dirName, _ := range validDirNames {
		numDirAddedSub, numDirRemovedSub, numVideoAddedSub, numVideoRemovedSub, err :=
			syncDirRecursivelyHelper(utils.AppendPath(location, dirName))
		if utils.LogError(err) {
			continue
		}
		numDirAdded += numDirAddedSub
		numDirRemoved += numDirRemovedSub
		numVideoAdded += numVideoAddedSub
		numVideoRemoved += numVideoRemovedSub
	}

	// 同步现有视频文件
	adds, removes, numFiles := syncVideosByLocationFlat(location, oldNumFiles, pageSize, videoList)
	numVideoAdded += adds
	numVideoRemoved += removes

	// 同步父目录信息
	if oldNumSubDirs != len(validDirNames) || oldNumFiles != numFiles {
		parentDir.NumSubDirs = len(validDirNames)
		parentDir.NumFiles = numFiles
		err = dir.UpdateByID(models.DB, parentDir, false, "NumSubDirs", "NumFiles")
		utils.LogError(err)
	}

	return numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved, nil
}

// numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved, error
func SyncDirRecursively(location string, fromTask bool) (int, int, int, int, error) {
	var t *task.TaskInfo
	if fromTask {
		var err error
		t, err = models.TaskHandler.AcceptTask("SyncDirRecursively")
		if err != nil {
			return 0, 0, 0, 0, err
		}
		defer models.TaskHandler.FinishTask()
	}

	numDirs := utils.NumDirs(location)
	if fromTask {
		t.Start(numDirs, "开始扫描目标目录")
	}

	numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved, err := syncDirRecursivelyHelper(location)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	if fromTask {
		res := &map[string]any{
			"numDirAdded":     numDirAdded,
			"numDirRemoved":   numDirRemoved,
			"numVideoAdded":   numVideoAdded,
			"numVideoRemoved": numVideoRemoved,
		}
		t.Success(res)
	}
	return numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved, nil
}

// numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved
func SyncDirBasesRecursively() (int, int, int, int, error) {
	t, err := models.TaskHandler.AcceptTask("SyncDirBasesRecursively")
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer models.TaskHandler.FinishTask()
	numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved := 0, 0, 0, 0

	numDirs := 0
	for _, loc := range config.Config.VideoLocations {
		numDirs += utils.NumDirs(loc)
	}

	t.Start(numDirs, "开始同步")
	for _, loc := range config.Config.VideoLocations {
		numDirAddedSub, numDirRemovedSub, numVideoAddedSub, numVideoRemovedSub, err :=
			SyncDirRecursively(loc, false)
		if utils.LogError(err) {
			continue
		}
		numDirAdded += numDirAddedSub
		numDirRemoved += numDirRemovedSub
		numVideoAdded += numVideoAddedSub
		numVideoRemoved += numVideoRemovedSub
	}
	res := &map[string]any{
		"numDirAdded":     numDirAdded,
		"numDirRemoved":   numDirRemoved,
		"numVideoAdded":   numVideoAdded,
		"numVideoRemoved": numVideoRemoved,
	}
	t.Success(res)
	return numDirAdded, numDirRemoved, numVideoAdded, numVideoRemoved, nil
}
