package video

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"video_worker/types"
	"video_worker/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

func ProcessMsg(msg amqp.Delivery) { // Possible Refactor with Template Method Pattern
	rabbitTask := util.ParseMessage(msg)
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
		fmt.Println("Begin Resolution Generation")
		ffmpegProcess(rabbitTask)
	} else if os.IsNotExist(err) {
		fmt.Println("File Doesn't Exist, skipping now")
	} else {
		fmt.Println("Error Checking File Existence: ", err)
	}

	// Contact Webhook to Notify of Completion

	// Tell User processing is done
}

func collectAndCreateMasterFile(rabbitTask types.RabbitTask) {
	dirPath := "./tmp/" + rabbitTask.Filename
	util.CreateDirectory(dirPath)

	path := util.CreatePath(rabbitTask)
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

	videoLoader, err := NewVideoLoader(absPath)
	if err != nil {
		fmt.Println("There was an issue creating the video loader!", err)
		return
	}

	scaler := SelectStarterScaler(videoLoader.VideoWidth, videoLoader.VideoHeight)

	scaler.handle(videoLoader.LoadedVideo, dirPath, videoLoader.VideoWidth, videoLoader.VideoHeight)

	fmt.Println("Finished Processing Video!")
}

// Video Processing Util

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
