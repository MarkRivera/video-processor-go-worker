package video

import (
	"fmt"
	"path/filepath"

	moviego "github.com/mowshon/moviego"
)

type FullHDScale struct {
	next VideoHandler
}

func (f *FullHDScale) handle(loadedVideo moviego.Video, dirPath string, videoWidth int, videoHeight int) {
	absPath, absErr := filepath.Abs(dirPath + "/full-hd.mp4")
	if absErr != nil {
		fmt.Printf("There was an issue with determining the abs file path for video scaling in the Full HD Scaler! %v", absErr)

		panic("Ending Worker")
	}

	loadedVideo.Resize(1920, 1080).Output(absPath).Run()
	f.next.handle(loadedVideo, dirPath, videoWidth, videoHeight)
}

func (f *FullHDScale) setNext(next VideoHandler) {
	f.next = next
}
