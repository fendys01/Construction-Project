package rabbit

import (
	"log"

	"github.com/streadway/amqp"
)

func handleError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

// Connect ...
func Connect(host string) (*amqp.Connection, *amqp.Channel, error) {
	var (
		err         error
		conn        *amqp.Connection
		amqpChannel *amqp.Channel
	)

	conn, err = amqp.Dial(host)
	handleError(err, "Can't connect to AMQP")

	amqpChannel, err = conn.Channel()
	handleError(err, "Can't create a amqpChannel")

	return conn, amqpChannel, err
}
