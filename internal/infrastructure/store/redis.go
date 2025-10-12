package store

import (
	"github.com/redis/go-redis/v9"
)

// redis://username:password@localhost:6789/3?dial_timeout=3&db=1&read_timeout=6s&max_retries=2

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
