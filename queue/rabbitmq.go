package queue

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
)

// CIQueue is a name of rabbitMQ queue with CI jobs
const CIQueue = "ciJobs"

// RMQQueue is used for operations with rabbitMQ queues
type RMQQueue struct {
	rabbitConnection *amqp.Connection
	channel          *amqp.Channel
	CIJobsQueue      amqp.Queue
}

func declareQueue(ch *amqp.Channel, queueName string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
}

// NewRMQQueue creates new copy of RMQQueue
func NewRMQQueue(addr string) *RMQQueue {
	conn, err := amqp.Dial(addr)
	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to rabbitmq: %s", err))
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Couldn't open rabbitmq channel: %s", err))
	}
	ciQueue, err := declareQueue(ch, CIQueue)
	if err != nil {
		panic(fmt.Sprintf("Couldn't declare CI jobs queue: %s", err))
	}
	Queue := &RMQQueue{
		rabbitConnection: conn,
		channel:          ch,
		CIJobsQueue:      ciQueue,
	}

	return Queue
}

// publish sends protobuf encoded message to provided queue
func (q *RMQQueue) publish(msg proto.Message, queue string) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return q.channel.Publish(
		"", queue, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}

// PublishCIJob publishes CIJob with provided data
func (q *RMQQueue) PublishCIJob(jobMsg *commonTypes.CIJob) error {
	return q.publish(jobMsg, q.CIJobsQueue.Name)
}

// Close gracefully breaks connection with rabbitMQ
func (q *RMQQueue) Close() {
	q.rabbitConnection.Close()
}
