package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "go-uptime-monitor", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr())
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	os.Setenv("APP_PORT", "9090")
	os.Setenv("DATABASE_NAME", "custom_monitor")
	defer os.Unsetenv("APP_PORT")
	defer os.Unsetenv("DATABASE_NAME")

	cfg, err := LoadConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, 9090, cfg.App.Port)
	assert.Equal(t, "custom_monitor", cfg.Database.Name)
}
