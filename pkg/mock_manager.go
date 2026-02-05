package pkg

import (
	"github.com/stretchr/testify/mock"
)

// MockManager is a mock implementation of Manager for testing
type MockManager struct {
	mock.Mock
}

// Install mocks package installation
func (m *MockManager) Install(packages ...string) error {
	args := m.Called(packages)
	return args.Error(0)
}

// IsInstalled mocks checking if a package is installed
func (m *MockManager) IsInstalled(pkg string) bool {
	args := m.Called(pkg)
	return args.Bool(0)
}

// Available mocks checking if the package manager is available
func (m *MockManager) Available() bool {
	args := m.Called()
	return args.Bool(0)
}

// Name mocks returning the package manager name
func (m *MockManager) Name() string {
	args := m.Called()
	return args.String(0)
}

// ExpectInstall sets up an expectation for Install with specific packages
func (m *MockManager) ExpectInstall(packages []string, err error) *mock.Call {
	return m.On("Install", packages).Return(err)
}

// ExpectInstallSuccess sets up an expectation for a successful Install
func (m *MockManager) ExpectInstallSuccess(packages ...string) *mock.Call {
	return m.ExpectInstall(packages, nil)
}

// ExpectIsInstalled sets up an expectation for IsInstalled
func (m *MockManager) ExpectIsInstalled(pkg string, installed bool) *mock.Call {
	return m.On("IsInstalled", pkg).Return(installed)
}

// ExpectAvailable sets up an expectation for Available
func (m *MockManager) ExpectAvailable(available bool) *mock.Call {
	return m.On("Available").Return(available)
}

// ExpectName sets up an expectation for Name
func (m *MockManager) ExpectName(name string) *mock.Call {
	return m.On("Name").Return(name)
}
