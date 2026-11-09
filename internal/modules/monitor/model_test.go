package monitor

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMonitor_BeforeCreate(t *testing.T) {
	m := &Monitor{
		Name: "Google Homepage",
		URL:  "https://google.com",
	}

	assert.Equal(t, uuid.Nil, m.ID)
	assert.Empty(t, m.Status)

	err := m.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, m.ID)
	assert.Equal(t, StatusPending, m.Status)
}
