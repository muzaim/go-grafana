package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestInitLogger_JSONFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	cfg := Config{
		Level:      "info",
		Filename:   logPath,
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
		Compress:   false,
		Stdout:     false,
	}

	log, err := InitLogger(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, log)

	log.Info("test log entry", zap.String("module", "testing"), zap.Int("status", 200))
	_ = log.Sync()

	data, err := os.ReadFile(logPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	var logMap map[string]interface{}
	err = json.Unmarshal(data, &logMap)
	assert.NoError(t, err, "Log output should be valid JSON")
	assert.Equal(t, "info", logMap["level"])
	assert.Equal(t, "test log entry", logMap["message"])
	assert.Equal(t, "testing", logMap["module"])
	assert.Equal(t, float64(200), logMap["status"])
}
