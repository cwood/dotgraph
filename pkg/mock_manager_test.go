package pkg

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockManager_Install(t *testing.T) {
	mock := new(MockManager)

	mock.ExpectInstallSuccess("git", "vim")

	err := mock.Install("git", "vim")

	assert.NoError(t, err)
	mock.AssertExpectations(t)
}

func TestMockManager_Install_Error(t *testing.T) {
	mock := new(MockManager)

	expectedErr := errors.New("package not found")
	mock.ExpectInstall([]string{"nonexistent"}, expectedErr)

	err := mock.Install("nonexistent")

	assert.Equal(t, expectedErr, err)
	mock.AssertExpectations(t)
}

func TestMockManager_IsInstalled(t *testing.T) {
	mock := new(MockManager)

	mock.ExpectIsInstalled("git", true)
	mock.ExpectIsInstalled("vim", false)

	assert.True(t, mock.IsInstalled("git"))
	assert.False(t, mock.IsInstalled("vim"))
	mock.AssertExpectations(t)
}

func TestMockManager_Available(t *testing.T) {
	mock := new(MockManager)

	mock.ExpectAvailable(true)

	assert.True(t, mock.Available())
	mock.AssertExpectations(t)
}

func TestMockManager_Name(t *testing.T) {
	mock := new(MockManager)

	mock.ExpectName("mock")

	assert.Equal(t, "mock", mock.Name())
	mock.AssertExpectations(t)
}

func TestMockManager_ImplementsInterface(t *testing.T) {
	// Verify MockManager implements Manager interface
	var _ Manager = (*MockManager)(nil)
}
