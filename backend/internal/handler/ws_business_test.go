package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// ============================================
// Health Handler Tests (补充)
// ============================================

func TestHealthHandlerNew_Creation(t *testing.T) {
	startTime := time.Now()
	handler := NewHealthHandlerNew(startTime)

	require.NotNil(t, handler)
}
