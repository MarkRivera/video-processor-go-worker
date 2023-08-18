package video

import (
	"fmt"
	"path/filepath"

	moviego "github.com/mowshon/moviego"
)

type HDScale struct {
	next VideoHandler
}

func (f *HDScale) handle(loadedVideo moviego.Video, dirPath string, videoWidth int, videoHeight int) {
	absPath, absErr := filepath.Abs(dirPath + "/hd.mp4")
	if absErr != nil {
		fmt.Printf("There was an issue with determining the abs file path for video scaling in the HD scaler! %v", absErr)

		panic("Ending Worker")
	}

	loadedVideo.Resize(1280, 720).Output(absPath).Run()
	f.next.handle(loadedVideo, dirPath, videoWidth, videoHeight)
}

func (f *HDScale) setNext(next VideoHandler) {
	f.next = next
}
