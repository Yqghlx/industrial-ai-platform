package service

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed version of TestHybridTokenBlacklist_Stop_Coverage
func TestHybridTokenBlacklist_Stop_Coverage_Fixed(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	blacklist := NewHybridTokenBlacklist(client)
	assert.NotNil(t, blacklist)

	// Stop closes shutdown channel
	blacklist.Stop()
}
