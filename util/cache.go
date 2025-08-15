package hlbaby

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
	ctx = context.Background()
)

func ResetCache() {
	err := redisClient.FlushAll(ctx).Err()
	if err != nil {
		panic(err)
	}
}

func CacheContent(url string, method string, content []byte, responseHeaders map[string]string, statusCode int) {
	key := url + ":" + method
	keyHeaders := key + ":headers"
	err := redisClient.Set(ctx, key, content, 5*time.Minute).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.HSet(ctx, keyHeaders, responseHeaders).Err()
	if err != nil {
		panic(err)
	}

	err = redisClient.Expire(ctx, keyHeaders, 5*time.Minute).Err()
	if err != nil {
		panic(err)
	}

	keyStatusCode := key + ":statusCode"
	err = redisClient.Set(ctx, keyStatusCode, statusCode, 5*time.Minute).Err()
	if err != nil {
		panic(err)
	}
}

func GetCachedContent(url string, method string) ([]byte, map[string]string, int, bool) {
	key := url + ":" + method
	keyHeaders := key + ":headers"

	keyStatusCode := key + ":statusCode"
	statusCodeStr, err := redisClient.Get(ctx, keyStatusCode).Result()
	if err == redis.Nil {
		return nil, nil, 0, false
	} else if err != nil {
		panic(err)
	}

	statusCode, err := strconv.ParseInt(statusCodeStr, 10, 32)
	if err != nil {
		panic(err)
	}

	val, err := redisClient.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil, 0, false
	} else if err != nil {
		panic(err)
	}

	headers, err := redisClient.HGetAll(ctx, keyHeaders).Result()
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
