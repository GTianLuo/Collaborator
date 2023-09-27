package dao

import (
	"context"
	"github.com/go-redis/redis/v8"
	"test.com/project-user/config"
	"time"
)

var Rc *RedisCache

type RedisCache struct {
	Rdb *redis.Client
}

func init() {
	rdb := redis.NewClient(config.AppConf.InitRedisOptions())
	Rc = &RedisCache{
		Rdb: rdb,
	}
}
func (rc *RedisCache) Put(ctx context.Context, key, value string, expire time.Duration) error {
	err := rc.Rdb.Set(ctx, key, value, expire).Err()
	return err
}
func (rc *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := rc.Rdb.Get(ctx, key).Result()
	return result, err
}
