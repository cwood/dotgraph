package exec

import (
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of CommandExecutor for testing
type MockExecutor struct {
	mock.Mock
}

// Run mocks command execution
func (m *MockExecutor) Run(name string, args ...string) RunResult {
	callArgs := m.Called(name, args)
	return callArgs.Get(0).(RunResult)
}

// LookPath mocks PATH lookup
func (m *MockExecutor) LookPath(cmd string) (string, error) {
	args := m.Called(cmd)
	return args.String(0), args.Error(1)
}

// ExpectRun sets up an expectation for Run with the given command and args
func (m *MockExecutor) ExpectRun(name string, args []string, result RunResult) *mock.Call {
	return m.On("Run", name, args).Return(result)
}

// ExpectRunSuccess sets up an expectation for a successful Run
func (m *MockExecutor) ExpectRunSuccess(name string, args []string) *mock.Call {
	return m.ExpectRun(name, args, RunResult{Success: true})
}

// ExpectRunFailure sets up an expectation for a failed Run
func (m *MockExecutor) ExpectRunFailure(name string, args []string, err error) *mock.Call {
	return m.ExpectRun(name, args, RunResult{Success: false, Error: err})
}

// ExpectLookPath sets up an expectation for LookPath
func (m *MockExecutor) ExpectLookPath(cmd string, path string, err error) *mock.Call {
	return m.On("LookPath", cmd).Return(path, err)
}

// ExpectCommandExists sets up an expectation that a command exists in PATH
func (m *MockExecutor) ExpectCommandExists(cmd string) *mock.Call {
	return m.ExpectLookPath(cmd, "/usr/bin/"+cmd, nil)
}

// ExpectCommandNotFound sets up an expectation that a command is not in PATH
func (m *MockExecutor) ExpectCommandNotFound(cmd string) *mock.Call {
	return m.ExpectLookPath(cmd, "", &CommandNotFoundError{Cmd: cmd})
}

// CommandNotFoundError is returned when a command is not found in PATH
type CommandNotFoundError struct {
	Cmd string
}

func (e *CommandNotFoundError) Error() string {
	return "command not found: " + e.Cmd
}
