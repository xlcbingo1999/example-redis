package bloom

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/xlcbingo1999/example-redis/utils"
)

type Bloom struct {
	Client    *redis.Client
	Key       string
	HashFuncs []F
}

func NewBloom(client *redis.Client) *Bloom {
	return &Bloom{
		Client:    client,
		Key:       "bloom",
		HashFuncs: NewFuncs(),
	}
}

func (b *Bloom) CloseBloom() {
	b.Client.Close()
}

func (b *Bloom) Add(str string) error {
	// 使用多个hash函数, 每次hash后得到的位置就设置为1即可
	ctx, cancel := context.WithCancel(context.Background())
	for _, f := range b.HashFuncs {
		offset := f(str)
		_, err := b.Client.SetBit(ctx, str, offset, 1).Result()
		if err != nil {
			cancel()
			panic(err.Error())
		}
	}
	cancel()
	return nil
}

func (b *Bloom) Exist(str string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	for _, f := range b.HashFuncs {
		offset := f(str)
		bitValue, err := b.Client.GetBit(ctx, str, offset).Result()
		if err != nil {
			cancel()
			panic(err.Error())
		}
		if bitValue != 1 {
			cancel()
			return false
		}
	}
	cancel()
	return true
}

func RunBloom() {
	bloom := NewBloom(utils.GetRedis())
	defer bloom.CloseBloom()

	bloom.Add("beijing")
	exist := bloom.Exist("beijing1")
	log.Println(exist)
}
