package hlutil

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx                = context.Background()
	DEFAULT_CACHE_TIME = 5 * time.Minute
)

func ResetCache() {
	err := redisClient.FlushAll(ctx).Err()
	if err != nil {
		panic(err)
	}
}

type RContent struct {
	url    string
	method string
}

// Redis key for content on specified method
func (c RContent) Key() string {
	return c.url + ":" + c.method
}

// Redis HTTP headers for content on specific method
func (c RContent) KeyHeaders() string {
	return c.Key() + ":headers"
}

// Redis HTTP status code for content on specific method
func (c RContent) KeyStatus() string {
	return c.Key() + ":status"
}

func CacheContent(url string, method string, content []byte, responseHeaders map[string]string, statusCode int) {
	rc := RContent{url: url, method: method}

	err := redisClient.Set(ctx, rc.Key(), content, DEFAULT_CACHE_TIME).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.HSet(ctx, rc.KeyHeaders(), responseHeaders).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.Expire(ctx, rc.KeyHeaders(), DEFAULT_CACHE_TIME).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.Set(ctx, rc.KeyStatus(), statusCode, DEFAULT_CACHE_TIME).Err()
	if err != nil {
		panic(err)
	}
}

func GetCachedContent(url string, method string) ([]byte, map[string]string, int, bool) {
	rc := RContent{url: url, method: method}

	statusCodeStr, err := redisClient.Get(ctx, rc.KeyStatus()).Result()
	if err == redis.Nil {
		return nil, nil, 0, false
	} else if err != nil {
		panic(err)
	}

	statusCode, err := strconv.ParseInt(statusCodeStr, 10, 32)
	if err != nil {
		panic(err)
	}

	val, err := redisClient.Get(ctx, rc.Key()).Bytes()
	if err == redis.Nil {
		return nil, nil, 0, false
	} else if err != nil {
		panic(err)
	}

	headers, err := redisClient.HGetAll(ctx, rc.KeyHeaders()).Result()
	if err != nil {
		panic(err)
	}

	return val, headers, int(statusCode), true
}

func CloseRedisClient() {
	if redisClient != nil {
		_ = redisClient.Close()
	}
}
