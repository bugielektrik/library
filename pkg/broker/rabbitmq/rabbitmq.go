package rabbitmq

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitMQClient *RabbitMQ

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMqConnection() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	RabbitMQClient = &RabbitMQ{
		Conn:    conn,
		Channel: ch,
	}
}
func CloseRabbitMQ() {
	if err := RabbitMQClient.Channel.Close(); err != nil {
		panic(err)
	}
}
