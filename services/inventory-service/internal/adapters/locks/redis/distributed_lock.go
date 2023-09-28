package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"bangungrid/services/inventory-service/internal/core/ports"
)

type RedisDistributedLock struct {
	client *redis.Client
}

func NewRedisDistributedLock(client *redis.Client) ports.DistributedLockPort {
	return &RedisDistributedLock{client: client}
}

func (l *RedisDistributedLock) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, func(), error) {
	if l.client == nil {
		// Fallback jika Redis tidak terkonfigurasi (misal di mock / memory)
		return true, func() {}, nil
	}

	token := uuid.New().String()
	success, err := l.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return false, nil, err
	}
	if !success {
		return false, nil, nil
	}

	unlock := func() {
		// Safe unlock script agar hanya token pemegang kunci yang bisa menghapus
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`
		_ = l.client.Eval(context.Background(), script, []string{lockKey}, token).Err()
	}

	return true, unlock, nil
}
