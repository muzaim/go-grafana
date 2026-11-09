package database

import (
	"testing"

	"go-up-monitoring/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestConnectPostgres_InvalidConfig(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     54329, // Invalid port
		User:     "invalid",
		Password: "invalid",
		Name:     "invalid_db",
		SSLMode:  "disable",
	}

	db, err := ConnectPostgres(cfg)
	// It should fail to connect when pinged or during open
	if err == nil {
		err = db.Ping()
		assert.Error(t, err)
	} else {
		assert.Error(t, err)
	}
}
