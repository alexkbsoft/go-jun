package rmqin

import (
	"context"

	"github.com/alexkbsoft/go-jun/cmd/api-server/internal/application"
)

type Consumer struct {
	// ch    *amqp.Channel
	queue string
	uc    application.TaskUseCases // тот же порт, что у HTTP
}

func NewConsumer(queue string, uc application.TaskUseCases) *Consumer {
	// return &Consumer{ch: ch, queue: queue, uc: uc}
	return &Consumer{queue: queue, uc: uc}
}

type createTaskMsg struct {
	ListID string `json:"listId"`
	Title  string `json:"title"`
}

func (c *Consumer) Run(ctx context.Context) error {
	// msgs, err := c.ch.Consume(c.queue, "", false /* autoAck */, false, false, false, nil)
	// if err != nil {
	// 	return err
	// }
	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return nil
	// 		case m, ok := <-msgs:
	// 			if !ok {
	// 				return errors.New("rabbitmq: delivery channel closed")
	// 			}
	// 			c.handle(ctx, m)
	// 	}
	// }
	<-ctx.Done()
	return nil
}

// func (c *Consumer) handle(ctx context.Context, m amqp.Delivery) {
// 	var msg createTaskMsg
// 	if err := json.Unmarshal(m.Body, &msg); err != nil {
// 		_ = m.Nack(false, false) // битое сообщение: не возвращать в очередь
// 		return
// 	}
// 	_, err := c.uc.CreateTask(ctx, msg.ListID, msg.Title, false)
// 	switch {
// 	case err == nil:
// 		_ = m.Ack(false)
// 	case errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrNotFound):
// 		_ = m.Nack(false, false) // повтор не поможет
// 	default:
// 		_ = m.Nack(false, true) // временная ошибка: вернуть в очередь
// 	}
// }
