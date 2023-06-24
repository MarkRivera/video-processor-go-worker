package util

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"video_worker/types"

	amqp "github.com/rabbitmq/amqp091-go"
)

func CreatePath(rabbitTask types.RabbitTask) string {
	return "./" +
		"tmp/" +
		rabbitTask.Filename +
		"/" +
		strconv.FormatInt(int64(rabbitTask.ChunkNumber), 10) +
		"_" +
		rabbitTask.Filename[:len(rabbitTask.Filename)-4]
}

func CreateDirectory(path string) {
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

// JSON Util
func ParseMessage(msg amqp.Delivery) types.RabbitTask {
	var rabbitTask types.RabbitTask
	err := json.Unmarshal([]byte(msg.Body), &rabbitTask)

	if err != nil {
		log.Panicf("There was an error parsing the JSON! %s", err)
	}

	return rabbitTask
}

// Error Util

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
