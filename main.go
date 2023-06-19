package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func createDirectory(path string) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Directory doesn't exist, so create it
			err = os.MkdirAll(path, 0755)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
			fmt.Println("Directory created:", path)
		} else {
			fmt.Println("Error checking directory:", err)
			return
		}
	}
}

func main() {
	createDirectory("./tmp")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"video_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			processMsg(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

type RabbitTask struct {
	ChunkName   string `json:"chunkName"`
	ChunkNumber int    `json:"chunkNumber"`
	Filename    string `json:"filename"`
	Data        string `json:"data"`
	TotalChunks int    `json:"totalChunks"`
}

func processMsg(msg amqp.Delivery) {
	var rabbitTask RabbitTask
	err := json.Unmarshal([]byte(msg.Body), &rabbitTask)

	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
		return
	}

	dirPath := "./tmp/" + rabbitTask.Filename

	// Collect Chunks and Create Master File
	collectAndConcatMasterFile(rabbitTask)

	// Use FFMPEG to Process Video Files
	_, err = os.Stat(dirPath + "/master.mp4")
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

func collectAndConcatMasterFile(rabbitTask RabbitTask) {
	dirPath := "./tmp/" + rabbitTask.Filename
	createDirectory(dirPath)

	path := createPath(rabbitTask)
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

func ffmpegProcess(rabbitTask RabbitTask) {
	fmt.Println("Processing with FFMPEG")
}

func createPath(rabbitTask RabbitTask) string {
	return "./" +
		"tmp/" +
		rabbitTask.Filename +
		"/" +
		strconv.FormatInt(int64(rabbitTask.ChunkNumber), 10) +
		"_" +
		rabbitTask.Filename[:len(rabbitTask.Filename)-4]
}

func createMasterFile(rabbitTask RabbitTask, files []fs.DirEntry, dirPath string) {
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
