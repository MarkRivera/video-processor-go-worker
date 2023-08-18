package video

import (
	"fmt"
	"path/filepath"

	moviego "github.com/mowshon/moviego"
)

type SDScale struct {
	next VideoHandler
}

func (f *SDScale) handle(loadedVideo moviego.Video, dirPath string, videoWidth int, videoHeight int) {
	absPath, absErr := filepath.Abs(dirPath + "/sd.mp4")
	if absErr != nil {
		fmt.Printf("There was an issue with determining the abs file path for video scaling in the SD Scaler! %v", absErr)

		panic("Ending Worker")
	}

	loadedVideo.Resize(640, 480).Output(absPath).Run()
}

func (f *SDScale) setNext(next VideoHandler) {
	f.next = next
}
