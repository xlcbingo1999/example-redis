package pipeline

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/xlcbingo1999/example-redis/utils"
)

func RunPipeline() {
	rdb := utils.GetRedis()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	cmds, err := rdb.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i := 0; i < 100; i++ {
			// pipe.Set(ctx, fmt.Sprintf("key%d", i), strconv.Itoa(i), -1).Result()
			str, err := pipe.Get(ctx, fmt.Sprintf("key%d", i)).Result()
			if err == nil {
				log.Println(str)
			} else if err == redis.Nil {
				log.Fatalln("not in redis")
				break
			} else {
				panic(err.Error())
			}
		}
		return nil
	})
	if err != nil {
		cancel()
		panic(err.Error())
	}

	for _, cmd := range cmds {
		fmt.Println(cmd.(*redis.StringCmd).Val())
	}
	cancel()
}
