package cache

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"

	"github.com/oidc-mytoken/server/internal/config"
)

type redisCache struct {
	client *redis.Client
	ctx    context.Context
}

// Get implements the Cache interface
func (c redisCache) Get(key string) (any, bool) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.WithError(err).Error("error while obtaining from cache")
		}
		return nil, false
	}
	return val, true
}

// Set implements the Cache interface
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
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.WithError(err).Fatal("could not connect to redis cache")
	}
	SetCache(
		redisCache{
			client: rdb,
			ctx:    context.Background(),
		},
	)
}
