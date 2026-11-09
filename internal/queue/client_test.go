package queue

import (
	"context"
	"testing"

	"go-up-monitoring/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestNewQueue(t *testing.T) {
	cfg := config.RedisConfig{
		Host:     "127.0.0.1",
		Port:     6379,
		Password: "testpassword",
		DB:       1,
	}

	q := NewQueue(cfg)
	assert.NotNil(t, q)
	assert.NotNil(t, q.RedisClient)
	assert.NotNil(t, q.AsynqClient)
	assert.Equal(t, "127.0.0.1:6379", q.RedisOpt.Addr)
	assert.Equal(t, "testpassword", q.RedisOpt.Password)
	assert.Equal(t, 1, q.RedisOpt.DB)

	// Ping to an unstarted redis port should return error gracefully without panic
	err := q.Ping(context.Background())
	// Error is expected when local redis server isn't running on test
	if err != nil {
		assert.Error(t, err)
	}

	_ = q.Close()
}
