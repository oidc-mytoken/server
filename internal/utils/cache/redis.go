package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
)

type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

func (c redisCache) Get(key string) (any, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			log.WithError(err).Error("error while obtaining from cache")
		}
		return nil, false
	}
	return val, true
}

func (c redisCache) Set(key string, value any, expiration time.Duration) {
	c.client.Set(c.ctx, key, value, expiration)
}

func initRedisCache() {
	rc := config.Get().Caching.External.Redis
	rdb := redis.NewClient(
		&redis.Options{
			Addr:     rc.Addr,
			Username: rc.Username,
			Password: rc.Password,
			DB:       rc.DB,
		},
	)
	setCache(
		redisCache{
			client: rdb,
			ctx:    context.Background(),
		},
	)
}
