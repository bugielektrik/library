package store

import (
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	url    string
	Client *redis.Client
}

func NewRedis(url string) (redis *Redis, err error) {
	redis = &Redis{
		url: url,
	}
	redis.Client, err = redis.connection()

	return
}

func (s Redis) connection() (client *redis.Client, err error) {
	opt, err := redis.ParseURL(s.url)
	if err != nil {
		return
	}
	client = redis.NewClient(opt)

	return
}
