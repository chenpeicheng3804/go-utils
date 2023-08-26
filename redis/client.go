package redisson

import (
	"github.com/go-redis/redis"
)

type Rdb struct {
	Client *redis.Client
	RLock  *RLock
}

func NewClient(Addr, Password string, DB int) *Rdb {
	return &Rdb{
		Client: redis.NewClient(&redis.Options{
			Addr:     Addr,
			Password: Password,
			DB:       DB,
		}),
	}
}
