package commons

import (
	"io"
	"medialpha-backend/models/video"
	"os"
)

func GetCoverBytes(name, location string) ([]byte, error) {
	v := video.Video{Name: name, Location: location}

	fd, err := os.Open(v.GetCoverFilename())
	defer fd.Close()
	data, err := io.ReadAll(fd)
	if err != nil {
		return nil, err
	}
	return data, err
}
