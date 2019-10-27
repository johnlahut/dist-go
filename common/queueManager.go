package common

import (
	"github.com/streadway/amqp"
)

var conn *amqp.Connection
var ch *amqp.Channel
var q amqp.Queue

// Closeq closes the queue
func Closeq() {
	ch.Close()
	conn.Close()
}

// Connq connects to the queue and channel
func Connq() {
	var err error
	conn, err = amqp.Dial(LocalQueue)
	FailOnError(err, "failed to connect to queue")

	ch, err = conn.Channel()
	FailOnError(err, "failed to connect to channel")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	FailOnError(err, "failed to set qos")

	delcareq("testing")
}

func delcareq(name string) {
	var err error
	q, err = ch.QueueDeclare(
		name,  // names
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	FailOnError(err, "failed to declare queue")
}

// Commitq commits the given message to the queue
func Commitq(msg []byte) {
	err := ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         msg,
		})
	FailOnError(err, "failed to send message to queue")
}

// Readq will start reading messages from the q
func Readq() {

}

// Getq returns the current queue
func Getq() amqp.Queue {
	return q
}

// GetCh returns the current channel
func GetCh() *amqp.Channel {
	return ch
}

func main() {

}
