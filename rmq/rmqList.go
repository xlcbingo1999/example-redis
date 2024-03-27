package rmq

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Msg struct {
	Topic string
	Body  []byte
}

type ListMQ struct {
	client *redis.Client
}

type Handler func(msg *Msg) error

func NewListMQ(client *redis.Client) *ListMQ {
	return &ListMQ{client: client}
}

func (q *ListMQ) Consume(ctx context.Context, topic string, h Handler) error {
	for {
		// 会从list中移除和获取的最后一个元素, 如果没有阻塞等待
		result, err := q.client.BRPop(ctx, 0, topic).Result()
		if err != nil {
			return err
		}
		h(&Msg{
			Topic: result[0],
			Body:  []byte(result[1]),
		})
	}
}

func (q *ListMQ) SendMsg(ctx context.Context, msg *Msg) error {
	return q.client.LPush(ctx, msg.Topic, msg.Body).Err()
}
