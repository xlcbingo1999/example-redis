package rmq

import (
	"context"
	"log"
	"strconv"
	"sync"
	"testing"

	"github.com/xlcbingo1999/example-redis/utils"
)

func TestRmqList(t *testing.T) {
	client := utils.GetRedis()
	q := NewListMQ(client)

	topic := "listMQ"
	count := 10

	wg := sync.WaitGroup{}
	wg.Add(count)

	ctx, cancel := context.WithCancel(context.Background())
	go q.Consume(ctx, topic, func(msg *Msg) error {
		log.Printf("consume %s\n", msg)
		wg.Done()

		cancel()
		return nil
	})

	for i := 0; i < count; i++ {
		q.SendMsg(ctx, &Msg{
			Topic: topic,
			Body:  []byte(topic + strconv.Itoa(i)),
		})
	}

	cancel()
}
