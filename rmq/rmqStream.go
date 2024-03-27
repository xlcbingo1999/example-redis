package rmq

import (
	"context"
	"log"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/xlcbingo1999/example-redis/utils"
)

var errBusyGroup = "BUSYGROUP Consumer Group name already exists"

type StreamMQ struct {
	client *redis.Client
	maxLen int  // 如果消息的数量大于这个值就删除
	approx bool // 精确地删除消息
}

type StreamMsg struct {
	ID        string
	Topic     string
	Body      []byte
	Partition int    // 分区
	Group     string // 消费者组
	Consumer  string // 消费者组里面的消费者
}

type StreamHandler func(msg *StreamMsg) error

func NewStreamMQ(client *redis.Client, maxLen int, approx bool) *StreamMQ {
	return &StreamMQ{
		client: client,
		maxLen: maxLen,
		approx: approx,
	}
}

func (q *StreamMQ) SendMsg(ctx context.Context, msg *StreamMsg) error {
	return q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: msg.Topic,
		MaxLen: int64(q.maxLen),
		Approx: q.approx,
		ID:     "*", // 这种方式就是让redis自己去生成时间戳和序列号
		Values: []interface{}{"body", msg.Body},
	}).Err()
}

func (q *StreamMQ) Consume(ctx context.Context, topic, group, consumer, start string, batchSize int, h StreamHandler) error {
	// 创建消费组
	err := q.client.XGroupCreateMkStream(ctx, topic, group, start).Err()
	if err != nil && err.Error() != errBusyGroup {
		return err
	}
	for {
		// 消费新的消息
		if err := q.consume(ctx, topic, group, consumer, ">", batchSize, h); err != nil {
			return err
		}
		// 消费没有ack的消息 确保消息都可以被消费一次
		if err := q.consume(ctx, topic, group, consumer, "0", batchSize, h); err != nil {
			return err
		}
	}
}

func (q *StreamMQ) consume(ctx context.Context, topic, group, consumer, id string, batchSize int, h StreamHandler) error {
	result, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  []string{topic, id},
		Count:    int64(batchSize),
	}).Result()
	if err != nil {
		return err
	}

	for _, msg := range result[0].Messages {
		err := h(&StreamMsg{
			ID:       msg.ID,
			Topic:    topic,
			Body:     []byte(msg.Values["body"].(string)),
			Group:    group,
			Consumer: consumer,
		})

		if err == nil {
			// 原生支持的Stream ACK
			err := q.client.XAck(ctx, topic, group, msg.ID).Err()
			if err != nil {
				return nil
			}
		}
	}
	return nil
}

func RunRmqStream() {
	client := utils.GetRedis()
	q := NewStreamMQ(client, 100, true)
	topic := "rmqstream"
	count := 10

	wg := sync.WaitGroup{}
	wg.Add(count * 4)

	ctx, cancel := context.WithCancel(context.Background())

	go q.Consume(ctx, topic, "group1", "consumer1", "$", 5, func(msg *StreamMsg) error {
		log.Printf("consume group1 consumer1: %s\n", msg)
		wg.Done()
		return nil
	})

	go q.Consume(ctx, topic, "group1", "consumer2", "$", 5, func(msg *StreamMsg) error {
		log.Printf("consume group1 consumer2: %s\n", msg)
		wg.Done()
		return nil
	})

	go q.Consume(ctx, topic, "group2", "consumer1", "$", 5, func(msg *StreamMsg) error {
		log.Printf("consume group2 consumer1: %s\n", msg)
		wg.Done()
		return nil
	})

	go q.Consume(ctx, topic, "group2", "consumer2", "$", 5, func(msg *StreamMsg) error {
		log.Printf("consume group2 consumer2: %s\n", msg)
		wg.Done()
		return nil
	})

	for i := 0; i < count; i++ {
		q.SendMsg(context.Background(), &StreamMsg{
			Topic: topic,
			Body:  []byte(topic + strconv.Itoa(i)),
		})
	}
	wg.Wait()

	cancel()
}
