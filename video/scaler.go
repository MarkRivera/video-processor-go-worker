package video

import (
	moviego "github.com/mowshon/moviego"
)

type VideoHandler interface {
	handle(loadedVideo moviego.Video, dirPath string, videoWidth int, videoHeight int) // This creates the videos
	setNext(next VideoHandler) // Tells us which is the next step in the chain
}

func SelectStarterScaler(videoWidth, videoHeight int) VideoHandler {
	sd := &SDScale{}

	hd := &HDScale{}
	hd.setNext(sd)

	fullHD := &FullHDScale{}
	fullHD.setNext(hd)

	switch {
	case videoWidth >= 1920 && videoHeight >= 1080:
		return fullHD

	case videoWidth >= 1280 && videoHeight >= 720:
		return hd

	case videoWidth >= 640 && videoHeight >= 480:
		return sd

	default:
		return nil
	}
}
