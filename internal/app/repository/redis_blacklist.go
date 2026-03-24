package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func redisAddr() string {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	return host + ":" + port
}

func connectRedis() (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr(),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	return rdb, nil
}

func blacklistKeyForToken(tokenString string) string {
	h := sha256.Sum256([]byte(tokenString))
	return "blacklist:" + hex.EncodeToString(h[:])
}

// AddTokenToBlacklist помещает JWT в Redis до истечения срока (отзыв сессии).
func (r *Repository) AddTokenToBlacklist(ctx context.Context, tokenString string, ttl time.Duration) error {
	if r.rd == nil || ttl <= 0 {
		return nil
	}
	key := blacklistKeyForToken(tokenString)
	return r.rd.Set(ctx, key, "1", ttl).Err()
}

// IsTokenBlacklisted проверяет, отозван ли токен.
func (r *Repository) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	if r.rd == nil {
		return false, nil
	}
	key := blacklistKeyForToken(tokenString)
	n, err := r.rd.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
