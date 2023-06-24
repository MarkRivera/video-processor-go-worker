package rabbit

import (
	"video_worker/util"

	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateConnection() *amqp.Connection {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	util.FailOnError(err, "Failed to connect to RabbitMQ")

	return conn
}

func CreateChannel(conn *amqp.Connection) *amqp.Channel {
	ch, err := conn.Channel()
	util.FailOnError(err, "Failed to open a channel")

	return ch
}
