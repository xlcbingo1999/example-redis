package utils

import "github.com/go-redis/redis/v8"

func GetRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "172.18.162.3:30379",
		Password: "",
		DB:       1,
	})
}
