package video

import (
	"fmt"

	"github.com/mowshon/moviego"
	"github.com/tidwall/gjson"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

type VideoLoader struct {
	AbsPath     string
	VideoWidth  int
	VideoHeight int
	LoadedVideo moviego.Video
}

func NewVideoLoader(absPath string) (*VideoLoader, error) {
	loader := &VideoLoader{AbsPath: absPath}

	videoProbe, probeErr := ffmpeg_go.Probe(absPath)
	if probeErr != nil {
		return nil, fmt.Errorf("there was an issue probing the video: %v", probeErr)
	}

	loader.VideoWidth = int(gjson.Get(videoProbe, "streams.0.width").Int())
	loader.VideoHeight = int(gjson.Get(videoProbe, "streams.0.height").Int())

	loadedVideo, loadErr := moviego.Load(loader.AbsPath)
	if loadErr != nil {
		return nil, fmt.Errorf("had a problem loading the video: %v", loadErr)
	}

	loader.LoadedVideo = loadedVideo
	return loader, nil
}
