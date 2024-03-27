package delock

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/xlcbingo1999/example-redis/utils"
)

type RedisLock struct {
	lockChan  chan struct{}
	rwLock    sync.RWMutex
	lockKey   string
	lockValue string
	client    *redis.Client
}

type InvokeMethod func() error

func GetRedisLock(client *redis.Client, lockKey string) *RedisLock {
	return &RedisLock{
		lockChan: make(chan struct{}, 1),
		lockKey:  lockKey,
		client:   client,
	}
}

func (lock *RedisLock) TryLock(ctx context.Context, interval time.Duration, timeout time.Duration, method InvokeMethod) error {
	// 1. 主线程
	if interval == 0 {
		interval = 100 * time.Millisecond
	}
	lock.rwLock.Lock()
	defer lock.rwLock.Unlock()

	var err error
	// 2. 协程启动业务逻辑
	go func() {
		for {
			if lock.lockValue == "" {
				lock.lockValue = uuid.New().String()
				hasSet, setErr := lock.client.SetNX(ctx, lock.lockKey, lock.lockValue, timeout).Result()
				if setErr != nil {
					err = setErr
					lock.lockChan <- struct{}{}
					return
				}

				if hasSet {
					lock.lockChan <- struct{}{}
					return
				}
			}

			time.Sleep(interval)
		}
	}()

	select {
	case <-lock.lockChan:
		if err != nil {
			return err
		}
		return method()
	case <-time.After(timeout):
		return errors.New("lock timeout")
	}
}

func (lock *RedisLock) UnLock(ctx context.Context) (bool, error) {
	if lock.lockValue == "" {
		return false, errors.New("锁释放")
	}

	script := "if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end"
	result, err := lock.client.Do(ctx, "EVAL", script, 1, lock.lockKey, lock.lockValue).Bool()
	if err != nil {
		return false, err
	}

	if !result {
		return false, errors.New("分布式锁释放错误")
	}

	lock.lockValue = ""
	return true, nil
}

func Yewu() error {
	time.Sleep(1 * time.Second)
	log.Println("DO yewu")
	return nil
}

func RunDelock() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	client := utils.GetRedis()

	lock := GetRedisLock(client, "delock")
	internal := 100 * time.Millisecond
	timeout := 2 * time.Second
	err := lock.TryLock(ctx, internal, timeout, Yewu)
	if err != nil {
		cancel()
		panic(err.Error())
	}

	result, err := lock.UnLock(ctx)
	if err != nil {
		cancel()
		panic(err.Error())
	}
	log.Println("result: ", result)
	cancel()

}
