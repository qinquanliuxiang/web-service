package data

import (
	"context"
	"fmt"
	"time"
	"web-service/base/conf"

	"github.com/redis/go-redis/v9"
)

var NeverExpires time.Duration = 0

// Redis redis 客户端
type Redis struct {
	client     *redis.Client
	expireTime time.Duration
	keyPrefix  string
}

func NewRedis(client *redis.Client) (*Redis, func()) {
	expireTime, err := conf.GetRedisExpireTime()
	if err != nil {
		panic(fmt.Sprintf("get redis expire time faild, err: %s", err.Error()))
	}

	claeup := func() {
		client.Close()
	}
	return &Redis{
		client:     client,
		expireTime: expireTime,
		keyPrefix:  conf.GetRedisKeyPrefix(),
	}, claeup
}

func CreateRDB(ctx context.Context) *redis.Client {
	switch conf.GetRdisMode() {
	case "sentinel":
		return initSentinelRedis(ctx)
	case "single":
		return initSingleRedis(ctx)
	default:
		panic("unsupported redis mode, please check the configuration redis.mode")
	}
}

func initSingleRedis(ctx context.Context) *redis.Client {
	host := conf.GetRdisHost()
	prot := conf.GetRdisPort()
	address := fmt.Sprintf("%s:%s", host, prot)
	if address == "" {
		panic("redis address is empty")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: conf.GetRdisPassword(),
		DB:       conf.GetRdisDB(),
	})
	s := rdb.Ping(ctx).Err()
	if s != nil {
		panic(s)
	}
	return rdb
}

func initSentinelRedis(ctx context.Context) *redis.Client {
	sentinelHosts := conf.GetRdisSentinelHosts()

	if len(sentinelHosts) == 0 {
		panic("redis sentinel is empty")
	}
	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       conf.GetRdisMasterName(),
		SentinelAddrs:    sentinelHosts,
		Password:         conf.GetRdisPassword(),
		SentinelPassword: conf.GetRdisSentinelPassword(),
		RouteByLatency:   true,
		DB:               conf.GetRdisDB(),
	})
	s := rdb.Ping(ctx).Err()
	if s != nil {
		panic(s)
	}
	return rdb
}

func (c *Redis) GetString(ctx context.Context, key string) (string, error) {
	saveKey := fmt.Sprintf("%s_%s", c.keyPrefix, key)
	return c.client.Get(ctx, saveKey).Result()
}

func (c *Redis) SetString(ctx context.Context, key string, value string, expireTime *time.Duration) error {
	saveKey := fmt.Sprintf("%s_%s", c.keyPrefix, key)

	if expireTime == nil {
		return c.client.Set(ctx, saveKey, value, c.expireTime).Err()
	}
	if expireTime == &NeverExpires {
		return c.client.Set(ctx, saveKey, value, 0).Err()
	}
	return c.client.Set(ctx, saveKey, value, *expireTime).Err()
}

func (c *Redis) GetInt64(ctx context.Context, key string) (int64, error) {
	saveKey := fmt.Sprintf("%s_%s", c.keyPrefix, key)
	return c.client.Get(ctx, saveKey).Int64()
}

func (c *Redis) SetInt64(ctx context.Context, key string, value int64, expireTime *time.Duration) error {
	saveKey := fmt.Sprintf("%s_%s", c.keyPrefix, key)

	if expireTime == nil {
		return c.client.Set(ctx, saveKey, value, c.expireTime).Err()
	}
	if expireTime == &NeverExpires {
		return c.client.Set(ctx, saveKey, value, 0).Err()
	}
	return c.client.Set(ctx, saveKey, value, *expireTime).Err()
}

func (c *Redis) Del(ctx context.Context, key string) error {
	saveKey := fmt.Sprintf("%s_%s", c.keyPrefix, key)
	return c.client.Del(ctx, saveKey).Err()
}

func (c *Redis) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}
