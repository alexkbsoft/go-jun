package rmqout

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/domain"
)

type Publisher struct {
	// ch       *amqp.Channel
	exchange string
}

// func NewPublisher(ch *amqp.Channel, exchange string) (*Publisher, error) {
func NewPublisher(exchange string) (*Publisher, error) {
	// if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
	// 	return nil, err
	// }
	// return &Publisher{ch: ch, exchange: exchange}, nil
	return &Publisher{exchange: exchange}, nil
}

// Формат сообщения в проводе: свой DTO, а не domain.Event
type eventMessage struct {
	Type       string    `json:"type"`
	TaskID     string    `json:"taskId"`
	ListID     string    `json:"listId"`
	OccurredAt time.Time `json:"occurredAt"`
}

// func (p *Publisher) Publish(ctx context.Context, e domain.Event) error {
func (p *Publisher) Publish(ctx context.Context, e domain.Event) error {
	body, err := json.Marshal(eventMessage{string(e.Type), e.TaskID, e.ListID, e.OccurredAt})
	if err != nil {
		return err
	}
	fmt.Println("EVENT published: ", string(body))
	// return p.ch.PublishWithContext(ctx, p.exchange, string(e.Type), false, false,
	// 	amqp.Publishing{
	// 		ContentType:  "application/json",
	// 		DeliveryMode: amqp.Persistent,
	// 		Body:         body,
	// 	})
	return nil
}
