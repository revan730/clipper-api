package queue

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	commonTypes "github.com/revan730/clipper-common/types"
	"github.com/streadway/amqp"
)

// CIJobsQueue represents rabbitMQ queue with CI jobs
type CIJobsQueue struct {
	rabbitConnection *amqp.Connection
	channel          *amqp.Channel
	jobsQueue        amqp.Queue
}

// NewQueue creates new copy of CIJobsQueue
func NewQueue(addr, queue string) *CIJobsQueue {
	conn, err := amqp.Dial(addr)
	if err != nil {
		panic(fmt.Sprintf("Couldn't connect to rabbitmq: %s", err))
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Couldn't open rabbitmq channel: %s", err))
	}
	q, err := ch.QueueDeclare(
		queue, // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	CIJobsQueue := &CIJobsQueue{
		rabbitConnection: conn,
		channel:          ch,
		jobsQueue:        q,
	}

	return CIJobsQueue
}

// PublishJob publishes CIJob with provided data
func (jq *CIJobsQueue) PublishJob(jobMsg *commonTypes.CIJob) error {
	body, err := proto.Marshal(jobMsg)
	if err != nil {
		return err
	}
	return jq.channel.Publish(
		"", jq.jobsQueue.Name, false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
}
