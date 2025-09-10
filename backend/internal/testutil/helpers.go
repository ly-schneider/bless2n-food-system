package testutil

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// AssertNoError is a helper function that fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

// AssertError is a helper function that fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
}

// AssertErrorContains is a helper function that fails the test if err is nil or doesn't contain the expected string
func AssertErrorContains(t *testing.T, err error, expectedString string) {
	t.Helper()
	require.Error(t, err)
	assert.Contains(t, err.Error(), expectedString)
}

// AssertValidObjectID checks if a string is a valid MongoDB ObjectID
func AssertValidObjectID(t *testing.T, id string) {
	t.Helper()
	_, err := primitive.ObjectIDFromHex(id)
	assert.NoError(t, err, "Expected valid ObjectID")
}

// Intentionally left out service-specific assertions here to avoid import cycles.

// TestContext returns a context for testing
func TestContext() context.Context {
	return context.Background()
}

// StringPtr returns a pointer to the given string
func StringPtr(s string) *string {
	return &s
}

// Float64Ptr returns a pointer to the given float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// IntPtr returns a pointer to the given int
func IntPtr(i int) *int {
	return &i
}

// BoolPtr returns a pointer to the given bool
func BoolPtr(b bool) *bool {
	return &b
}
