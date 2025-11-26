package store

import (
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Connection *redis.Client
}

func NewRedis(url string) (store Redis, err error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return
	}
	store.Connection = redis.NewClient(opt)

	return
}
