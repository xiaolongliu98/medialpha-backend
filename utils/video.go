package utils

import (
	"bytes"
	"fmt"
	"medialpha-backend/constant"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func GenerateFrame(videoFile, imageFile string, duration int) error {
	duration /= 2
	info, err := os.Stat(imageFile)
	// err != nil: no such file or dir
	if err == nil && !info.IsDir() {
		err = os.Remove(imageFile)
	}
	hour, minute, second := duration/3600, (duration%3600)/60, duration%3600%60
	args := fmt.Sprintf("ffmpeg -ss %d:%d:%d -i '%s' -frames:v 1 -s 384x216 '%s'", hour, minute, second, videoFile, imageFile)
	cmd := exec.Command("powershell", args)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return err
	}
	//fmt.Println(buf.String())
	return nil
}

// return x seconds
func ReadVideoDuration(fileName string) (int, error) {
	args := fmt.Sprintf("ffprobe -i '%s'  -show_entries format=duration -v quiet -of csv='p=0'", fileName)
	cmd := exec.Command("powershell", args)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	dStr := strings.Trim(buf.String(), "\r\n")
	d, err := strconv.ParseFloat(dStr, 64)
	if err != nil {
		return 0, err
	}
	return int(d), err
}

/*
*

	Default K-V as follows:

------------------------------------------------------

	"index":        "", // for videos, audio
	"duration":     "", // for videos, audio
	"width":        "", // for videos
	"height":       "", // for videos
	"codec_name":   "", // for videos, audio
	"r_frame_rate": "", // for videos
	"bit_rate":     "", // for videos, audio
	"nb_frames":    "", // for videos
	"codec_type":   "", // for videos, audio, subtitle
	"sample_rate":  "", // for audio
*/
func ReadVideoStreamsInfo(fileName string) ([]*map[string]string, error) {
	args := fmt.Sprintf("ffprobe -i '%s'  -show_streams", fileName)
	cmd := exec.Command("powershell", args)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	s := buf.String()
	//fmt.Println(s)
	ss := strings.Split(s, "[STREAM]")[1:]

	var streamInfos []*map[string]string

	for _, each := range ss {
		info := map[string]string{
			"index":        "", // for videos, audio
			"duration":     "", // for videos, audio
			"width":        "", // for videos
			"height":       "", // for videos
			"codec_name":   "", // for videos, audio
			"r_frame_rate": "", // for videos
			"bit_rate":     "", // for videos, audio
			"nb_frames":    "", // for videos
			"codec_type":   "", // for videos, audio, subtitle
			"sample_rate":  "", // for audio
		}
		lines := strings.Split(each, "\n")
		for j, line := range lines {
			if j == len(lines)-2 {
				break
			}
			line = strings.Trim(line, "\r\n ")
			if len(line) == 0 {
				continue
			}
			res := strings.Split(line, "=")
			if len(res) != 2 {
				continue
			}
			key, value := res[0], res[1]
			_, exists := info[key]
			if exists {
				info[key] = value
			}
		}

		streamInfos = append(streamInfos, &info)
	}

	return streamInfos, nil
}

func IsVideoFile(filename string) bool {
	return S(filename).EndsWithAny(constant.Video.SuffixList()...)
}
