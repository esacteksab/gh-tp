// SPDX-License-Identifier: MIT

package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "backup-test")
	require.NoError(t, err)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil && !os.IsNotExist(err) { // Don't log error if dir already gone
			Logger.Errorf("Error removing temp dir %s: %v", path, err)
		}
	}(tempDir)

	// Test case 1: Successful backup (remains the same)
	t.Run("SuccessfulBackup", func(t *testing.T) {
		sourceContent := []byte("test file content")
		sourcePath := filepath.Join(tempDir, "source.txt")
		err := os.WriteFile(sourcePath, sourceContent, 0o600)
		require.NoError(t, err)
		destPath := filepath.Join(tempDir, "dest.txt")
		err = BackupFile(sourcePath, destPath)
		require.NoError(t, err) // Expect success
		destContent, err := os.ReadFile(destPath)
		require.NoError(t, err)
		assert.Equal(t, sourceContent, destContent)
	})

	// Test case 2: Source file does not exist
	t.Run("SourceFileNotFound", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.txt")
		destPath := filepath.Join(tempDir, "dest2.txt")

		err := BackupFile(nonExistentPath, destPath)

		// Assert error is returned
		require.Error(t, err)
		// Use errors.Is for robust check of wrapped errors
		assert.ErrorIs(t, err, os.ErrNotExist, "Expected os.ErrNotExist error")
	})

	// Test case 3: Permission denied (remains the same, may still be flaky)
	t.Run("PermissionDenied", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}
		sourcePath := filepath.Join(tempDir, "source4.txt")
		err := os.WriteFile(sourcePath, []byte("test"), 0o600)
		require.NoError(t, err)
		restrictedDir := filepath.Join(tempDir, "noperm")
		// Ensure parent dir exists before setting permissions
		err = os.MkdirAll(filepath.Dir(restrictedDir), 0o755)
		require.NoError(t, err)
		err = os.Mkdir(restrictedDir, 0o500) // read + execute only
		// Defer removing restricted dir first if needed, handle potential errors
		defer os.Remove(restrictedDir) // Simple remove, might fail if file created inside
		require.NoError(t, err)

		destPath := filepath.Join(restrictedDir, "dest4.txt")
		err = BackupFile(sourcePath, destPath)
		assert.Error(t, err) // Expect an error (likely permission denied)
	})
}

func TestCheckFilesByExtensionExist(t *testing.T) {
	fileExts := []string{".tofu", ".tf"}

	tf, err := os.CreateTemp("", "foo-*.tf")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tf.Name())

	tofu, err := os.CreateTemp("", "foo-*.tofu")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tofu.Name())

	files := checkFilesByExtension("/tmp", fileExts)

	require.FileExists(t, tf.Name())
	require.FileExists(t, tofu.Name())
	assert.True(t, files)
}

func TestCheckFilesByExtensionDoNotExist(t *testing.T) {
	fileExts := []string{".tofu", ".tf"}

	files := checkFilesByExtension("/tmp", fileExts)

	assert.False(t, files)
}

func TestExistsOrCreatedExists(t *testing.T) {
	createLogger(false)
	plan, err := os.CreateTemp("", "plan.out")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(plan.Name())

	md, err := os.CreateTemp("", "plan.md")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(md.Name())

	files := []tpFile{
		{Name: plan.Name(), Purpose: "Plan"},
		{Name: md.Name(), Purpose: "Markdown"},
	}

	r, w, _ := os.Pipe()
	color.Output = w

	exists := existsOrCreated(files)

	err = w.Close()
	if err != nil {
		log.Fatalf("Error closing pipe: %s", err)
	}

	// os.Stdout = oldStdout
	var buf bytes.Buffer

	_, err = io.Copy(&buf, r)
	if err != nil {
		log.Fatalf("Error copying from reader: %s", err)
	}

	// fmt.Println(buf.String())
	output := buf.String()
	expectedOutput := "✔  Plan Created...\n✔  Markdown Created..."
	assert.Contains(t, output, expectedOutput)
	assert.NoError(t, exists)
}

func TestExistsOrCreatedDoesNotExists(t *testing.T) {
	if Logger == nil {
		createLogger(false)
	}

	files := []tpFile{
		{Name: "plan.out", Purpose: "Plan"},
		{Name: "plan.md", Purpose: "Markdown"},
	}

	r, w, _ := os.Pipe()
	color.Output = w

	exists := existsOrCreated(files)

	err := w.Close()
	if err != nil {
		log.Fatalf("Error closing pipe: %s", err)
	}

	// os.Stdout = oldStdout
	var buf bytes.Buffer

	_, err = io.Copy(&buf, r)
	if err != nil {
		log.Fatalf("Error copying from reader: %s", err)
	}

	// fmt.Println(buf.String())
	output := buf.String()
	expectedOutput := "✕  Plan Failed to Create\n✕  Markdown Failed to Create\n"
	assert.Contains(t, output, expectedOutput)
	assert.NoError(t, exists)
}

func Test_ValidateFilePath(t *testing.T) {
	if Logger == nil { // Logger setup if needed
		Logger = log.NewWithOptions(os.Stderr, log.Options{Level: log.InfoLevel})
	}

	type args struct {
		path string
	}
	tests := []struct {
		name       string
		args       args
		wantPath   string // Expected returned path
		wantErr    bool
		wantErrMsg string // Optional: check error content
	}{
		{
			name:     "valid_filename",
			args:     args{path: "test.txt"},
			wantPath: "test.txt", // Simple filename
			wantErr:  false,
		},
		{
			name:     "already_has_current_directory_prefix",
			args:     args{path: "./test.txt"},
			wantPath: "test.txt", // Clean removes ./
			wantErr:  false,
		},
		{
			name:       "attempt_directory_traversal",
			args:       args{path: "../test.txt"},
			wantPath:   "../test.txt", // Return original invalid path
			wantErr:    true,
			wantErrMsg: "must be a filename only",
		},
		{
			name:       "attempt_absolute_path",
			args:       args{path: "/etc/passwd"},
			wantPath:   "/etc/passwd", // Return original invalid path
			wantErr:    true,
			wantErrMsg: "must be a filename only", // Fails this check first
		},
		{
			name:       "attempt_nested_directory",
			args:       args{path: "subdir/test.txt"},
			wantPath:   "subdir/test.txt", // Return original invalid path
			wantErr:    true,
			wantErrMsg: "must be a filename only",
		},
		{
			name:       "attempt_double_dot_hidden_directory",
			args:       args{path: "..hidden/test.txt"},
			wantPath:   "..hidden/test.txt", // Return original invalid path
			wantErr:    true,
			wantErrMsg: "must be a filename only",
		},
		{
			name:     "clean_path_with_dots",
			args:     args{path: "./././test.txt"},
			wantPath: "test.txt", // Clean removes extra ./
			wantErr:  false,
		},
		{
			name:     "filename_with_dots",
			args:     args{path: "test.file.with.dots.txt"},
			wantPath: "test.file.with.dots.txt", // Simple filename
			wantErr:  false,
		},
		{
			name:       "empty_path",
			args:       args{path: ""},
			wantPath:   "",   // Return original empty path
			wantErr:    true, // Is an error
			wantErrMsg: "filename cannot be empty",
		},
		{
			name:     "filename_with_special_characters",
			args:     args{path: "test-file_123.txt"},
			wantPath: "test-file_123.txt", // Simple filename
			wantErr:  false,
		},
		{
			name:       "command_injection_attempt",
			args:       args{path: "file.txt; rm -rf"},
			wantPath:   "file.txt; rm -rf",            // Return original invalid path
			wantErr:    true,                          // Is an error
			wantErrMsg: "contains invalid characters", // Fails regex check
		},
		{
			name:       "another_command_injection_attempt",
			args:       args{path: "$(cat /etc/passwd)"},
			wantPath:   "$(cat /etc/passwd)", // Return original invalid path
			wantErr:    true,                 // Is an error
			wantErrMsg: "must be a filename only",
		},
		{
			name:       "backtick_command_injection",
			args:       args{path: "`echo hello`"},
			wantPath:   "`echo hello`",                // Return original invalid path
			wantErr:    true,                          // Is an error
			wantErrMsg: "contains invalid characters", // Fails regex check
		},
		{
			name:       "pipe_command_injection",
			args:       args{path: "file.txt | cat /etc/passwd"},
			wantPath:   "file.txt | cat /etc/passwd", // Return original invalid path
			wantErr:    true,                         // Is an error
			wantErrMsg: "must be a filename only",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateFilePath(tt.args.path)

			// Check error status
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check error message content if expected
			if tt.wantErr && tt.wantErrMsg != "" {
				require.Error(t, err) // Ensure error is not nil before checking message
				assert.Contains(t, err.Error(), tt.wantErrMsg, "Error message mismatch")
			}

			// Check returned path (should match expectation regardless of error)
			if got != tt.wantPath {
				t.Errorf("validateFilePath() returned path = %q, want %q", got, tt.wantPath)
			}
		})
	}
}

func Test_createLogger(t *testing.T) {
	type args struct {
		verbose bool
	}
	tests := []struct {
		name           string
		args           args
		wantDebugLevel bool
	}{
		{
			name: "verbose true",
			args: args{
				verbose: true,
			},
			wantDebugLevel: true,
		},
		{
			name: "verbose false",
			args: args{
				verbose: false,
			},
			wantDebugLevel: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call function under test
			createLogger(tt.args.verbose)

			isDebugEnabled := Logger.GetLevel() == log.DebugLevel
			isInfoEnabled := Logger.GetLevel() == log.InfoLevel

			if isDebugEnabled != tt.wantDebugLevel {
				t.Errorf(
					"createLogger() debug level enabled = %v, want %v",
					isDebugEnabled,
					tt.wantDebugLevel,
				)
			}

			// If Logger.GetLevel() != log.DebugLevel
			// Logger.GetLevel() == log.InfoLevel
			if isInfoEnabled == tt.wantDebugLevel {
				t.Errorf(
					"createLogger() info level enabled = %v, want %v",
					isInfoEnabled,
					!tt.wantDebugLevel,
				)
			}
		})
	}
}

func Test_getDirectories(t *testing.T) {
	// Save original environment variables to restore later
	origHome := os.Getenv("HOME")
	origXdgConfig := os.Getenv("XDG_CONFIG_HOME")
	origPwd := os.Getenv("PWD")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("XDG_CONFIG_HOME", origXdgConfig)
		os.Setenv("PWD", origPwd)
	}()
	// Save original working directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	defer os.Chdir(originalWd)

	// Create a temporary directory for test files
	homeDir, err := os.MkdirTemp("", "foo")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(homeDir) // Clean up temp directory

	// Change to the temp directory
	if err := os.Chdir(homeDir); err != nil {
		t.Error(err)
	}

	// os.UserHomeDir looks for $HOME
	os.Setenv("HOME", homeDir)

	// os.UserConfigDir looks for `$XDG_CONFIG_HOME`
	os.Setenv("XDG_CONFIG_HOME", homeDir+"/.config")

	// os.Getwd() will use $PWD, so we set one
	os.Setenv("PWD", homeDir)

	tests := []struct {
		name          string
		setupEnv      func()
		wantHomeDir   string
		wantConfigDir string
		wantCwd       string
		wantErr       bool
	}{
		{
			name: "all directories exist",
			setupEnv: func() {
				os.Setenv("HOME", homeDir)
				os.Setenv("XDG_CONFIG_HOME", homeDir+"/.config")
				os.Setenv("PWD", homeDir)
			},
			wantHomeDir:   homeDir,
			wantConfigDir: homeDir + "/.config",
			wantCwd:       homeDir,
			wantErr:       false,
		},
		{
			name: "HOME not set",
			setupEnv: func() {
				os.Unsetenv("HOME")
				os.Setenv("XDG_CONFIG_HOME", homeDir+"/.config")
				os.Setenv("PWD", homeDir)
			},
			wantHomeDir:   "", // Will be empty as HOME is unset
			wantConfigDir: "",
			wantCwd:       "",
			wantErr:       true, // Expect an error
		},
		{
			name: "XDG_CONFIG_HOME not set",
			setupEnv: func() {
				os.Setenv("HOME", homeDir)
				os.Unsetenv("XDG_CONFIG_HOME") // Unset XDG_CONFIG_HOME
				os.Setenv("PWD", homeDir)
			},
			wantHomeDir:   homeDir,
			wantConfigDir: homeDir + "/.config", // Default is $HOME/.config on unix
			wantCwd:       homeDir,
			wantErr:       false, // No error, falls back to default
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup Env for test
			tt.setupEnv()
			gotHomeDir, gotConfigDir, gotCwd, err := getDirectories()
			if (err != nil) != tt.wantErr {
				t.Errorf("getDirectories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotHomeDir != tt.wantHomeDir {
				t.Errorf("getDirectories() gotHomeDir = %v, want %v", gotHomeDir, tt.wantHomeDir)
			}
			if gotConfigDir != tt.wantConfigDir {
				t.Errorf(
					"getDirectories() gotConfigDir = %v, want %v",
					gotConfigDir,
					tt.wantConfigDir,
				)
			}
			if gotCwd != tt.wantCwd {
				t.Errorf("getDirectories() gotCwd = %v, want %v", gotCwd, tt.wantCwd)
			}
		})
	}
}
