package score

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateUserID(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateUserID(""))
	require.Error(t, ValidateUserID("   "))
	require.Error(t, ValidateUserID(strings.Repeat("u", MaxUserIDLength+1)))
	require.NoError(t, ValidateUserID("user_123"))
	require.NoError(t, ValidateUserID("a"))
	require.NoError(t, ValidateUserID(strings.Repeat("u", MaxUserIDLength)))
}

func TestValidateUserIDCountsRunes(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateUserID(strings.Repeat("世", MaxUserIDLength)))
	require.Error(t, ValidateUserID(strings.Repeat("世", MaxUserIDLength+1)))
}

func TestValidateValue(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateValue(-1))
	require.Error(t, ValidateValue(-100))
	require.NoError(t, ValidateValue(0))
	require.NoError(t, ValidateValue(1500))
	require.NoError(t, ValidateValue(9223372036854775807))
}

func TestValidateLimit(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateLimit(0))
	require.Error(t, ValidateLimit(-1))
	require.Error(t, ValidateLimit(MaxQueryLimit+1))
	require.NoError(t, ValidateLimit(MinQueryLimit))
	require.NoError(t, ValidateLimit(MaxQueryLimit))
	require.NoError(t, ValidateLimit(10))
}

func TestValidateWriteCombinesErrors(t *testing.T) {
	t.Parallel()

	err := ValidateWrite("", -1)

	require.Error(t, err)
	require.Contains(t, err.Error(), "user ID must not be empty")
	require.Contains(t, err.Error(), "score must be greater than or equal to")
}

func TestValidateWriteReturnsNilWhenValid(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateWrite("user_1", 100))
}
