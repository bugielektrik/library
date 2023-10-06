package currency

import (
	"context"
	"time"

	"github.com/patrickmn/go-cache"
)

func (c *Client) initCacheRefresher() {
	c.GetRateFromCacheByID("USD")

	timer := time.NewTicker(4 * time.Minute)
	go func() {
		for {
			<-timer.C
			c.GetRateFromCacheByID("USD")
		}
	}()
}

func (c *Client) GetRateFromCacheByID(id string) (dest Rate, err error) {
	if data, found := c.caches.Get(id); found {
		return data.(Rate), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dest, err = c.GetRateByID(ctx, id, time.Now())
	if err != nil {
		return
	}
	c.caches.Set(id, dest, cache.DefaultExpiration)

	return
}
