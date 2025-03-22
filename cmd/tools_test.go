// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "backup-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			Logger.Error(err)
		}
	}(tempDir)

	// Test case 1: Successful backup
	t.Run("SuccessfulBackup", func(t *testing.T) {
		// Create source file with test content
		sourceContent := []byte("test file content")
		sourcePath := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourcePath, sourceContent, 0o600)
		require.NoError(t, err)

		// Set destination path
		destPath := filepath.Join(tempDir, "dest.txt")

		// Execute backupFile function
		err = backupFile(sourcePath, destPath)

		// Assert no error occurred
		require.NoError(t, err)

		// Verify destination file exists and has correct content
		destContent, err := os.ReadFile(destPath)
		require.NoError(t, err)
		assert.Equal(t, sourceContent, destContent)
	})

	// Test case 2: Source file does not exist
	t.Run("SourceFileNotFound", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")
		destPath := filepath.Join(tempDir, "dest2.txt")

		err := backupFile(nonExistentPath, destPath)

		// Assert error is returned
		require.Error(t, err)
		// Verify the error is a file not found error
		assert.True(t, os.IsNotExist(err))
	})

	// Test case 3: Cannot create destination (invalid path)
	t.Run("CannotCreateDestination", func(t *testing.T) {
		// Create valid source file
		sourcePath := filepath.Join(tempDir, "source3.txt")
		err := os.WriteFile(sourcePath, []byte("test"), 0o600)
		require.NoError(t, err)

		// Use invalid destination path (directory that doesn't exist)
		invalidDestPath := filepath.Join(tempDir, "nonexistent-dir", "dest3.txt")

		err = backupFile(sourcePath, invalidDestPath)

		// Assert error is returned
		assert.Error(t, err)
	})

	// Test case 4: Permission denied (if possible to test)
	// Note: This test might not work on all systems or with all permissions
	t.Run("PermissionDenied", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		// Create source file
		sourcePath := filepath.Join(tempDir, "source4.txt")
		err := os.WriteFile(sourcePath, []byte("test"), 0o600)
		require.NoError(t, err)

		// Create a directory with no write permission
		restrictedDir := filepath.Join(tempDir, "noperm")
		err = os.Mkdir(restrictedDir, 0o500) // read + execute only
		require.NoError(t, err)

		destPath := filepath.Join(restrictedDir, "dest4.txt")

		err = backupFile(sourcePath, destPath)

		// Assert error occurred (may be permission denied)
		assert.Error(t, err)
	})
}
