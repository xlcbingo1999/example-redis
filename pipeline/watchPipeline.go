package pipeline

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xlcbingo1999/example-redis/utils"
)

// watchDemo 在key值不变的情况下将其值+1
func watchDemo(ctx context.Context, key string) error {
	rdb := utils.GetRedis()
	// 因为redis是单线程的, 所以使用watch就相当于加上了互斥锁
	return rdb.Watch(ctx, func(tx *redis.Tx) error {
		n, err := tx.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			return err
		}
		// 假设操作耗时5秒
		// 5秒内我们通过其他的客户端修改key，当前事务就会失败
		time.Sleep(5 * time.Second)
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, key, n+1, time.Hour)
			return nil
		})
		return err
	}, key)
}
