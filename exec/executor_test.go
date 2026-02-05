package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealExecutor_Run_Success(t *testing.T) {
	executor := NewRealExecutor()

	result := executor.Run("echo", "hello")

	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.Empty(t, result.LogFile)
}

func TestRealExecutor_Run_Failure(t *testing.T) {
	executor := NewRealExecutor()

	// Run a command that will fail
	result := executor.Run("false")

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
	// Log file should be created on failure
	assert.NotEmpty(t, result.LogFile)
}

func TestRealExecutor_Run_CommandNotFound(t *testing.T) {
	executor := NewRealExecutor()

	result := executor.Run("nonexistent-command-12345")

	assert.False(t, result.Success)
	assert.Error(t, result.Error)
}

func TestRealExecutor_LookPath_Exists(t *testing.T) {
	executor := NewRealExecutor()

	path, err := executor.LookPath("echo")

	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestRealExecutor_LookPath_NotFound(t *testing.T) {
	executor := NewRealExecutor()

	_, err := executor.LookPath("nonexistent-command-12345")

	assert.Error(t, err)
}

func TestMockExecutor_Run(t *testing.T) {
	mock := new(MockExecutor)

	// Set up expectation
	mock.ExpectRunSuccess("git", []string{"status"})

	// Call the mock
	result := mock.Run("git", "status")

	assert.True(t, result.Success)
	mock.AssertExpectations(t)
}

func TestMockExecutor_Run_Failure(t *testing.T) {
	mock := new(MockExecutor)

	expectedErr := &CommandNotFoundError{Cmd: "missing"}
	mock.ExpectRunFailure("missing", nil, expectedErr)

	result := mock.Run("missing")

	assert.False(t, result.Success)
	assert.Equal(t, expectedErr, result.Error)
	mock.AssertExpectations(t)
}

func TestMockExecutor_LookPath_Exists(t *testing.T) {
	mock := new(MockExecutor)

	mock.ExpectCommandExists("brew")

	path, err := mock.LookPath("brew")

	assert.NoError(t, err)
	assert.Equal(t, "/usr/bin/brew", path)
	mock.AssertExpectations(t)
}

func TestMockExecutor_LookPath_NotFound(t *testing.T) {
	mock := new(MockExecutor)

	mock.ExpectCommandNotFound("yay")

	_, err := mock.LookPath("yay")

	assert.Error(t, err)
	assert.IsType(t, &CommandNotFoundError{}, err)
	mock.AssertExpectations(t)
}

func TestMockExecutor_MultipleExpectations(t *testing.T) {
	mock := new(MockExecutor)

	mock.ExpectRunSuccess("git", []string{"clone", "repo"})
	mock.ExpectRunSuccess("git", []string{"checkout", "main"})
	mock.ExpectCommandExists("git")

	result1 := mock.Run("git", "clone", "repo")
	result2 := mock.Run("git", "checkout", "main")
	_, err := mock.LookPath("git")

	assert.True(t, result1.Success)
	assert.True(t, result2.Success)
	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestCommandNotFoundError(t *testing.T) {
	err := &CommandNotFoundError{Cmd: "brew"}

	assert.Equal(t, "command not found: brew", err.Error())
}
