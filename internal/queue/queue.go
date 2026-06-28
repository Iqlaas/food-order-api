package queue

import (
	"context"
	"encoding/json"
	"food-order-api/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

const QueueName = "order_notifications"

type QueueManager struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewQueueManager(url string) (*QueueManager, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(QueueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &QueueManager{conn: conn, ch: ch}, nil
}

func (q *QueueManager) PublishOrderEvent(ctx context.Context, event models.OrderPlacedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return q.ch.PublishWithContext(ctx, "", QueueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}

func (q *QueueManager) ConsumeEvents() (<-chan amqp.Delivery, error) {
	return q.ch.Consume(QueueName, "", true, false, false, false, nil)
}

func (q *QueueManager) Close() {
	if q.ch != nil {
		q.ch.Close()
	}
	if q.conn != nil {
		q.conn.Close()
	}
}