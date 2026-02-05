package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cwood/dotgraph/exec"
	"github.com/cwood/dotgraph/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test request
func newTestRequest[T any](t *testing.T, config T) (*Request[T], string) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
	require.NoError(t, err)

	mockExec := new(exec.MockExecutor)
	mockPkg := new(pkg.MockManager)

	req := &Request[T]{
		Env: Environment{
			OS:      "linux",
			Arch:    "amd64",
			WorkDir: tmpDir,
		},
		Services: Services{
			Executor:  mockExec,
			Installer: mockPkg,
		},
		Config: config,
	}

	return req, tmpDir
}

func TestFileExists_True(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, ".zpm")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	condition := FileExists[any]("$HOME/.zpm")
	result := condition(req)

	assert.True(t, result)
}

func TestFileExists_False(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	condition := FileExists[any]("$HOME/.nonexistent")
	result := condition(req)

	assert.False(t, result)
}

func TestFileExists_TildeExpansion(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	// Create a test file
	testFile := filepath.Join(tmpDir, ".ssh", "id_rsa")
	err := os.MkdirAll(filepath.Dir(testFile), 0700)
	require.NoError(t, err)
	err = os.WriteFile(testFile, []byte("key"), 0600)
	require.NoError(t, err)

	condition := FileExists[any]("~/.ssh/id_rsa")
	result := condition(req)

	assert.True(t, result)
}

func TestCommandExists_True(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	mockExec := req.Services.Executor.(*exec.MockExecutor)
	mockExec.ExpectCommandExists("git")

	condition := CommandExists[any]("git")
	result := condition(req)

	assert.True(t, result)
	mockExec.AssertExpectations(t)
}

func TestCommandExists_False(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	mockExec := req.Services.Executor.(*exec.MockExecutor)
	mockExec.ExpectCommandNotFound("yay")

	condition := CommandExists[any]("yay")
	result := condition(req)

	assert.False(t, result)
	mockExec.AssertExpectations(t)
}

func TestIsMac(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	// Default is linux
	condition := IsMac[any]()
	assert.False(t, condition(req))

	// Change to darwin
	req.Env.OS = "darwin"
	assert.True(t, condition(req))
}

func TestIsLinux(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	// Default is linux
	condition := IsLinux[any]()
	assert.True(t, condition(req))

	// Change to darwin
	req.Env.OS = "darwin"
	assert.False(t, condition(req))
}

func TestEnvSet_True(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	os.Setenv("TEST_VAR_12345", "value")
	defer os.Unsetenv("TEST_VAR_12345")

	condition := EnvSet[any]("TEST_VAR_12345")
	result := condition(req)

	assert.True(t, result)
}

func TestEnvSet_False(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	condition := EnvSet[any]("NONEXISTENT_VAR_12345")
	result := condition(req)

	assert.False(t, result)
}

func TestNot(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	// Create a condition that returns true
	trueCondition := func(req *Request[any]) bool { return true }

	notCondition := Not(trueCondition)
	result := notCondition(req)

	assert.False(t, result)
}

func TestAnd(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	trueCondition := func(req *Request[any]) bool { return true }
	falseCondition := func(req *Request[any]) bool { return false }

	// All true
	andAll := And(trueCondition, trueCondition)
	assert.True(t, andAll(req))

	// One false
	andOneFalse := And(trueCondition, falseCondition)
	assert.False(t, andOneFalse(req))

	// All false
	andAllFalse := And(falseCondition, falseCondition)
	assert.False(t, andAllFalse(req))
}

func TestOr(t *testing.T) {
	req, tmpDir := newTestRequest[any](t, nil)
	defer os.RemoveAll(tmpDir)

	trueCondition := func(req *Request[any]) bool { return true }
	falseCondition := func(req *Request[any]) bool { return false }

	// All true
	orAll := Or(trueCondition, trueCondition)
	assert.True(t, orAll(req))

	// One true
	orOneTrue := Or(falseCondition, trueCondition)
	assert.True(t, orOneTrue(req))

	// All false
	orAllFalse := Or(falseCondition, falseCondition)
	assert.False(t, orAllFalse(req))
}

func TestExpandPathWithWorkDir(t *testing.T) {
	tests := []struct {
		path     string
		workDir  string
		expected string
	}{
		{"$HOME/.zpm", "/home/user", "/home/user/.zpm"},
		{"~/.ssh/id_rsa", "/home/user", "/home/user/.ssh/id_rsa"},
		{"/etc/hosts", "/home/user", "/etc/hosts"},
		{"$HOME/dotfiles/../.zshrc", "/home/user", "/home/user/.zshrc"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := expandPathWithWorkDir(tt.path, tt.workDir)
			assert.Equal(t, tt.expected, result)
		})
	}
}
