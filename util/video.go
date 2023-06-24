package util

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"video_worker/types"

	"github.com/mowshon/moviego"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/tidwall/gjson"
	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func ProcessMsg(msg amqp.Delivery) {
	rabbitTask := ParseMessage(msg)
	dirPath := "./tmp/" + rabbitTask.Filename

	// If Master File Exists, return early
	_, statErr := os.Stat(dirPath + "/master.mp4")
	if statErr == nil {
		fmt.Println("Master File Exists! Stopping Worker Execution.")
		return
	}

	// Collect Chunks and Create Master File
	collectAndCreateMasterFile(rabbitTask)

	// Use FFMPEG to Process Master File
	_, err := os.Stat(dirPath + "/master.mp4")
	if err == nil {
		ffmpegProcess(rabbitTask)
	} else if os.IsNotExist(err) {
		fmt.Println("File Doesn't Exist, skipping now")
	} else {
		fmt.Println("Error Checking File Existence: ", err)
	}

	// Store in DB
	// Tell User processing is done
}

func collectAndCreateMasterFile(rabbitTask types.RabbitTask) {
	dirPath := "./tmp/" + rabbitTask.Filename
	CreateDirectory(dirPath)

	path := CreatePath(rabbitTask)
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("There was an Error creating the file!", err)
		return
	}
	defer file.Close()

	// Decode and Store
	decodedBytes, err := base64.StdEncoding.DecodeString(rabbitTask.Data)
	if err != nil {
		fmt.Println("There was an error decoding file!", err)
		return
	}

	_, err = file.Write(decodedBytes)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	file.Close()

	// Check if all chunks are present, if so, begin concatination
	dir, err := os.Open(dirPath)
	if err != nil {
		fmt.Println("There was an error reading the Directory!", err)
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		fmt.Println("There was an error reading the File Length!", err)
		return
	}

	fileCount := 0
	for _, file := range fileInfos {
		if file.Mode().IsRegular() {
			fileCount++
		}
	}

	if fileCount == rabbitTask.TotalChunks {
		files, err := os.ReadDir(dirPath)
		if err != nil {
			fmt.Println("There was an error reading the directory!", err)
		}

		createMasterFile(rabbitTask, files, dirPath)
	}
}

func ffmpegProcess(rabbitTask types.RabbitTask) {
	dirPath := "./tmp/" + rabbitTask.Filename

	absPath, absErr := filepath.Abs(dirPath + "/master.mp4")
	if absErr != nil {
		fmt.Println("There was an issue with determining the abs file path!", absErr)
	}

	videoProbe, _ := ffmpeg_go.Probe(absPath)
	videoWidth := gjson.Get(videoProbe, "streams.0.width").Int()
	videoHeight := gjson.Get(videoProbe, "streams.0.height").Int()

	// Prepare to load
	loadedVideo, loadErr := moviego.Load(absPath)
	if loadErr != nil {
		fmt.Println("Had a problem loading the video!", loadErr)
	}

	if videoWidth > 1920 && videoHeight > 1080 {
		scaleVideo("1080", loadedVideo, dirPath)
		scaleVideo("720", loadedVideo, dirPath)
		scaleVideo("480", loadedVideo, dirPath)

		return
	}

	if videoWidth > 1280 && videoHeight > 720 {
		scaleVideo("720", loadedVideo, dirPath)
		scaleVideo("480", loadedVideo, dirPath)

		return
	}
}

// Video Processing Util

func scaleVideo(targetResolution string, loadedVideo moviego.Video, dirPath string) {
	if targetResolution == "1080" {
		absPath, absErr := filepath.Abs(dirPath + "/full-hd.mp4")
		if absErr != nil {
			fmt.Println("There was an issue with determining the abs file path for video Scaling!", absErr)
		}
		loadedVideo.Resize(1920, 1080).Output(absPath).Run()
	}

	if targetResolution == "720" {
		absPath, absErr := filepath.Abs(dirPath + "/hd.mp4")
		if absErr != nil {
			fmt.Println("There was an issue with determining the abs file path for video Scaling!", absErr)
		}
		loadedVideo.Resize(1280, 720).Output(absPath).Run()
	}

	if targetResolution == "480" {
		absPath, absErr := filepath.Abs(dirPath + "/sd.mp4")
		if absErr != nil {
			fmt.Println("There was an issue with determining the abs file path for video Scaling!", absErr)
		}
		loadedVideo.Resize(640, 480).Output(absPath).Run()
	}
}

func createMasterFile(rabbitTask types.RabbitTask, files []fs.DirEntry, dirPath string) {
	output, err := os.Create("./tmp/" + rabbitTask.Filename + "/master.mp4")
	if err != nil {
		fmt.Println("There was an error creating a new file for the chunks", err)
	}

	defer output.Close()

	for _, file := range files {
		fileInfo, err := file.Info()

		if err != nil {
			fmt.Println("There was an error getting file info!")
			return
		}

		path := filepath.Join(dirPath, fileInfo.Name())
		file, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("There was an error reading the file!", err)
		}

		_, err = output.Write(file)
		if err != nil {
			fmt.Println("There was an error writing to the new file!", err)
		}

		deleteChunks(path, "master.mp4")
	}
}

func deleteChunks(path string, keepFilename string) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		// Check if the file name matches the file to keep
		if info.Name() != keepFilename {
			err := os.Remove(path) // Delete the file
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println("There was an issue deleting the chunks!", err)
	}
}
