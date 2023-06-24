package main

import (
	"log"

	"video_worker/rabbit"
	"video_worker/util"
)

func main() {
	util.CreateDirectory("./tmp")
	conn := rabbit.CreateConnection()
	defer conn.Close()

	ch := rabbit.CreateChannel(conn)
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"video_queue", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)

	util.FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	util.FailOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			util.ProcessMsg(d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
