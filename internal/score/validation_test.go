package score

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUserID(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateUserID(""))
	require.Error(t, ValidateUserID(strings.Repeat("u", MaxUserIDLength+1)))
	require.NoError(t, ValidateUserID("user_123"))
}

func TestValidateValue(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateValue(-1))
	require.NoError(t, ValidateValue(0))
	require.NoError(t, ValidateValue(1500))
}

func TestValidateLimit(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateLimit(0))
	require.Error(t, ValidateLimit(MaxQueryLimit+1))
	require.NoError(t, ValidateLimit(10))
}
