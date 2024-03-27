package delayqueue

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"gotest.tools/assert"
)

// AddToQueue
// 1. add the timePiece(sample: "1645614542") to sorted set
// 2. rpush the real data to timePiece
//
// @delaySecond, the expected delay seconds, 600 means delay 600 second
// @maxTTL, the max time data will live if there is no consumer
func AddToQueue(ctx context.Context, rdb *redis.Client, key string, value string, delaySecond, maxTTL int64) error {
	expireSecond := time.Now().Unix() + delaySecond
	// generate time piece to store v
	timePiece := fmt.Sprintf("dq:%s:%d", key, expireSecond)
	z := redis.Z{Score: float64(expireSecond), Member: timePiece} // 创建了一个zset
	v, err := rdb.ZAddNX(ctx, key, &z).Result()                   // 添加zset的成员, 返回数量
	if err != nil {
		return err
	}
	_, err = rdb.RPush(ctx, timePiece, value).Result()
	if err != nil {
		return err
	}

	// new timePiece will set expire time
	if v > 0 {
		// consumer will also deleted the item
		rdb.Expire(ctx, timePiece, time.Second*time.Duration(maxTTL+delaySecond))
		// sorted set max live time
		rdb.Expire(ctx, key, time.Hour*24*3)
	}
	return err
}

// GetFromQueue
// 1. get a timePiece from sorted set which is before time.Now()
// 2. lpop the real data from timePiece
//
// Usage: Use it in a script or goroutine
func GetFromQueue(ctx context.Context, rdb *redis.Client, key string) (chan string, chan error) {
	resCh := make(chan string)
	errCh := make(chan error, 1)
	go func() {
		defer close(resCh)
		defer close(errCh)
		for {
			now := time.Now().Unix()
			opt := redis.ZRangeBy{Min: "0", Max: strconv.FormatInt(now, 10), Count: 1}
			val, err := rdb.ZRangeByScore(ctx, key, &opt).Result()
			if err != nil {
				errCh <- err
				return
			}
			// sleep 1s if the queue is empty
			if len(val) == 0 {
				select {
				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				case <-time.After(time.Second):
					continue
				}
			}
			for _, listK := range val {
				for {
					// read from the timePiece
					s, err := rdb.LPop(ctx, listK).Result()
					if err == nil {
						select {
						case resCh <- s:
						case <-ctx.Done():
							errCh <- ctx.Err()
							return
						}
					} else if err == redis.Nil {
						rdb.ZRem(ctx, key, listK)
						rdb.Del(ctx, listK)
						break
					} else {
						errCh <- err
						return
					}
				}
			}
		}
	}()
	return resCh, errCh
}

func getRdb() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "172.18.162.3:30379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

// add 4 items to queue and read from queue
func TestBase(t *testing.T) {
	rdb := getRdb()
	ctx, cancel := context.WithCancel(context.Background())

	delayQueueName := "delay_queue_name"
	sleepSecond := int64(1)

	// add 4 items
	values := []string{"a", "b", "c", "d"}
	for _, v := range values {
		err := AddToQueue(ctx, rdb, delayQueueName, v, sleepSecond, 100)
		if err != nil {
			t.Fatal(err)
		}
	}

	// read from delay queue
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		resCh, errCh := GetFromQueue(ctx, rdb, delayQueueName)
		for res := range resCh { // 阻塞读, 不断从管道里面去读取新的值, 这就是延迟队列的基本实现方式
			t.Log("get from queue, v:", res)
			h := values[0]
			values = values[1:]
			assert.Equal(t, h, res)
			if len(values) == 0 {
				cancel()
			}
		}
		// check error
		for err := range errCh { // 阻塞读新的错误
			assert.Error(t, err, context.Canceled.Error())
		}
		wg.Done()
	}()

	// add timeout check
	select {
	case <-time.After(time.Second * time.Duration(sleepSecond+1)):
		t.Fatal("error timeout")
	case <-ctx.Done():
		wg.Wait()
	}
}

func TestAutoExpire(t *testing.T) {
	rdb := getRdb()
	ctx, cancel := context.WithCancel(context.Background())

	delayQueueName := "delay_queue"
	sleepSecond := int64(1)
	maxTTL := int64(2)

	// add 2 items
	values := []string{"a"}
	for _, v := range values {
		err := AddToQueue(ctx, rdb, delayQueueName, v, sleepSecond, maxTTL)
		if err != nil {
			t.Fatal(err)
		}
	}

	// wait till expire
	select {
	case <-time.After(time.Second * time.Duration(sleepSecond+maxTTL+1)):
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	// try consume
	resCh, errCh := GetFromQueue(ctx, rdb, delayQueueName)
	select {
	case <-resCh:
		t.Fatal(fmt.Errorf("data should expired"))
	case <-time.After(time.Second):
		t.Log("check success")
	}

	// cancel ctx
	cancel()

	// check error
	for err := range errCh {
		assert.Error(t, err, context.Canceled.Error())
	}
}

func clean(ctx context.Context, rdb *redis.Client, key string) {
	res, err := rdb.ZRange(ctx, key, 0, 1000).Result()
	if err != nil {
		panic(err)
	}
	for _, v := range res {
		err = rdb.Del(ctx, v).Err()
		if err != nil {
			panic(err)
		}
	}

}

// Use: go test -bench=. -run=none
func BenchmarkAddToQueue(b *testing.B) {
	rdb := getRdb()
	ctx := context.Background()
	delayQueueName := "delay_queue"
	clean(ctx, rdb, delayQueueName)
	for i := 0; i < b.N; i++ {
		err := AddToQueue(ctx, rdb, delayQueueName, strconv.Itoa(i+1), -1, 100)
		if err != nil {
			b.FailNow()
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	b.ResetTimer()
	var res int64
	var count int32
	// equals to runtime.GOMAXPROCS(0)
	b.RunParallel(func(pb *testing.PB) {
		resCh, errCh := GetFromQueue(ctx, rdb, delayQueueName)
		for {
			select {
			case x := <-resCh:
				c, _ := strconv.Atoi(x)
				atomic.AddInt64(&res, int64(c))
				atomic.AddInt32(&count, 1)
			case <-ctx.Done():
				break
			}
			if atomic.LoadInt32(&count) >= int32(b.N) {
				cancel()
				break
			}
		}
		// can't use pb.next
		for pb.Next() {
		}

		for err := range errCh {
			if err != context.Canceled && err != nil {
				b.FailNow()
			}
		}
	})

	assert.Equal(b, int64(1+b.N)*int64(b.N)/2, atomic.LoadInt64(&res))
}
