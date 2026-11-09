package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_BeforeCreate(t *testing.T) {
	u := &User{
		Email:    "test@example.com",
		FullName: "Test User",
	}

	assert.Equal(t, uuid.Nil, u.ID)
	err := u.BeforeCreate(nil)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, u.ID)
}
