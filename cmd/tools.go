// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/spf13/viper"
)

const (
	// Max filename length (common limit)
	maxFilenameLength = 255
)

// determineBinary finds the IaC binary to use based on flags, config, or PATH discovery.
func determineBinary() (string, error) {
	// 1. Check Viper (which checks flags first, then config)
	binaryFromConfig, err := getBinaryFromConfig()
	if err != nil {
		return "", err
	}
	if binaryFromConfig != "" {
		return binaryFromConfig, nil
	}

	// 2. Auto-detect if not specified
	detectedBinary, err := autoDetectBinary()
	if err != nil {
		return "", err
	}
	if detectedBinary != "" {
		return detectedBinary, nil
	}

	// 3. Handle the case where no binary is found
	return "", buildNoBinaryFoundError()
}

// getBinaryFromConfig checks for a binary specified via flag or config.
func getBinaryFromConfig() (string, error) {
	v := viper.IsSet("binary")
	Logger.Debugf("Binary is set: %v", v)
	viperBinary := viper.GetString("binary")
	if viperBinary == "" {
		return "", nil // Not set
	}

	// Validate if specified
	if viperBinary != "terraform" && viperBinary != "tofu" {
		return "", fmt.Errorf(
			"invalid binary specified ('%s'): must be 'terraform' or 'tofu'",
			viperBinary,
		)
	}

	// Ensure it's actually findable
	_, err := safeexec.LookPath(viperBinary)
	if err != nil {
		return "", fmt.Errorf(
			"binary '%s' specified but not found in PATH: %w",
			viperBinary,
			err,
		)
	}

	Logger.Debugf("Using binary specified via flag or config: %s", viperBinary)
	return viperBinary, nil
}

// autoDetectBinary attempts to find 'tofu' or 'terraform' in the PATH.
func autoDetectBinary() (string, error) {
	Logger.Debug("Binary not specified, attempting auto-detection...")
	binariesToFind := []string{"tofu", "terraform"}
	var foundBinaries []string
	for _, binName := range binariesToFind {
		binPath, lookupErr := safeexec.LookPath(binName)
		if lookupErr == nil && len(binPath) > 0 {
			foundBinaries = append(foundBinaries, binName)
			Logger.Debugf("Found '%s' in PATH at '%s'", binName, binPath)
		} else {
			Logger.Debugf("Did not find '%s' in PATH: %v", binName, lookupErr)
		}
	}

	// Evaluate auto-detection results
	if len(foundBinaries) == 0 {
		return "", nil // No binaries found, handle in the main function
	}

	if len(foundBinaries) > 1 {
		return "", buildMultipleBinariesFoundError(foundBinaries)
	}

	// Exactly one binary found
	detectedBinary := foundBinaries[0]
	Logger.Debugf("Auto-detected binary: %s", detectedBinary)
	return detectedBinary, nil
}

// Regex for allowed filename characters
var validFilenameChars = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

// checkFilesByExtension checks if files with any of the specified extensions exist in a directory
//
// This function iterates through a list of file extensions and uses filepath.Glob to find
// any files matching those extensions in the given directory. It returns true as soon as it
// finds at least one file with any of the specified extensions, and false if no matching
// files are found or if an error occurs.
//
// Parameters:
//
//	dir - The directory path to search for files
//	exts - A slice of file extensions to check for (should include the dot, e.g., ".tf", ".tofu")
//
// Returns:
//
//	bool - true if at least one file with any of the specified extensions exists,
//	       false if no matching files are found or if an error occurs
func checkFilesByExtension(dir string, exts []string) bool {
	var exists bool
	for _, v := range exts {
		files, err := filepath.Glob(filepath.Join(dir, "*"+v))
		if err != nil {
			exists = false
			return exists
		}
		if len(files) > 0 {
			exists = true
			return exists
		}
	}
	return exists
}

// existsOrCreated checks if specified files exist or were created and reports their status.
// It logs the status of each file and displays colored indicators to the user.
//
// Parameters:
//   - files: A slice of tpFile structures containing file information
//
// Returns:
//   - error: Returns nil if status reporting completes, or an error if writing to output fails
func existsOrCreated(files []tpFile) error {
	for _, v := range files {
		// First check if the file exists
		exists := doesExist(v.Name)
		var err error

		if !exists {
			// File doesn't exist - log debug info and display failure status
			Logger.Debugf("%s file %s was not created", v.Purpose, v.Name)
			_, err = fmt.Fprintf(color.Output, "%s  %s%s",
				bold(red("✕")), v.Purpose, " Failed to Create\n")
		} else {
			// File exists - log debug info and display success status
			Logger.Debugf("%s file %s was created", v.Purpose, v.Name)
			_, err = fmt.Fprintf(color.Output, "%s  %s%s",
				bold(green("✔")), v.Purpose, " Created...\n")
		}
		if err != nil {
			return fmt.Errorf("failed to display status: %w", err)
		}
	}
	return nil
}

// doesExist checks if a file or directory exists at the specified path.
//
// This function uses os.Stat to determine if the path exists in the filesystem.
//
// Parameters:
//
//	path - The file system path to check for existence
//
// Returns:
//
//	bool - true if the path exists, false otherwise
func doesExist(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false
	}
	return true
}

// getDirectories returns the user's home directory, config directory, and current working directory.
// It handles platform-specific differences for config directories.
func getDirectories() (homeDir, configDir, cwd string, err error) {
	// Get home directory
	homeDir, err = os.UserHomeDir()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Get config directory
	configDir, err = os.UserConfigDir()
	if err != nil {
		return homeDir, "", "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Get current working directory
	cwd, err = os.Getwd()
	if err != nil {
		return homeDir, configDir, "", fmt.Errorf(
			"failed to get current working directory: %w",
			err,
		)
	}

	return homeDir, configDir, cwd, nil
}

// BackupFile copies a file from source to destination.
// It relies on os package functions for path handling and permissions.
//
// Parameters:
//   - source: Path to the source file.
//   - dest: Path to the destination file.
//
// Returns:
//   - error: nil on success, or an error describing what went wrong (file ops).
func BackupFile(source, dest string) error {
	// Check if source exists using os.Stat
	sourceInfo, statErr := os.Stat(source)
	if statErr != nil {
		if errors.Is(statErr, fs.ErrNotExist) {
			// If the source doesn't exist for a backup, this should be an error
			return fmt.Errorf("backup source file %q does not exist: %w", source, os.ErrNotExist)
		}
		// Other error stating the file
		return fmt.Errorf("cannot access source file %q: %w", source, statErr)
	}
	// Check if source is a directory
	if sourceInfo.IsDir() {
		return fmt.Errorf("backup source %q is a directory, expected a file", source)
	}

	// Open source file
	srcFile, err := os.Open( //nolint:gosec // source path provided by trusted caller context (e.g., config backup)
		source,
	)
	if err != nil {
		// I wouldn't expect this given the above checking
		return fmt.Errorf("failed to open source file %q: %w", source, err)
	}
	defer func() {
		if err = srcFile.Close(); err != nil {
			Logger.Errorf("Error closing source file %q: %v", source, err)
		}
	}()

	// Create destination file
	destFile, err := os.Create( //nolint:gosec // dest path provided by trusted caller context (e.g., config backup)
		dest,
	)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", dest, err)
	}
	defer func() {
		if err := destFile.Close(); err != nil {
			Logger.Errorf("Error closing destination file %q: %v", dest, err)
		}
	}()

	// Copy file contents
	bytesCopied, err := io.Copy(destFile, srcFile)
	if err != nil {
		_ = os.Remove(dest) // Attempt cleanup
		return fmt.Errorf("failed to copy content from %q to %q: %w", source, dest, err)
	}
	Logger.Debugf("Copied %d bytes from %s to %s", bytesCopied, source, dest)

	// Sync destination file
	if err = destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file %q: %w", dest, err)
	}

	Logger.Debugf("Successfully backed up %s to %s", source, dest)
	return nil // Success
}

// createLogger creates and configures the package-level Logger instance
// based on the desired verbosity.
func createLogger(verbose bool) {
	var level log.Level
	var reportCaller, reportTimestamp bool
	var timeFormat string

	// Define options based on verbose
	if verbose {
		reportCaller = true
		reportTimestamp = true
		timeFormat = "2006/01/02 15:04:05"
		level = log.DebugLevel
	} else {
		reportCaller = false
		reportTimestamp = false
		timeFormat = time.Kitchen
		level = log.InfoLevel
	}

	var instanceToUse *log.Logger // Use a local variable first

	if Logger == nil {
		instanceToUse = log.NewWithOptions(os.Stderr, log.Options{
			ReportCaller:    reportCaller,
			ReportTimestamp: reportTimestamp,
			TimeFormat:      timeFormat,
			Level:           level, // Set level on creation
		})
		if instanceToUse == nil {
			os.Exit(1)
		}
	} else {
		instanceToUse = Logger // Reconfigure the existing package Logger
		instanceToUse.SetLevel(level)
		instanceToUse.SetReportTimestamp(reportTimestamp)
		instanceToUse.SetTimeFormat(timeFormat)
		instanceToUse.SetReportCaller(reportCaller)
	}

	maxWidth := 4 // Use lowercase for local var
	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper(log.DebugLevel.String())).
		Bold(true).MaxWidth(maxWidth).Foreground(lipgloss.Color("14"))
	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper(log.FatalLevel.String())).
		Bold(true).MaxWidth(maxWidth).Foreground(lipgloss.Color("9"))
	instanceToUse.SetStyles(styles)

	Logger = instanceToUse // Assign the created/reconfigured instance

	log.SetDefault(Logger)

	// Check Logger again just to be paranoid before logging
	if Logger != nil {
		// Use the package Logger variable for the final confirmation log
		Logger.Debugf(
			"Logger configured. Verbose: %t, Level set to: %s",
			verbose,
			Logger.GetLevel(),
		)
	}
}

// validateFilePath checks if a given path string represents a simple, safe filename
// intended for use within the current directory.
// It performs checks for:
// - Emptiness
// - Directory traversal components (e.g., "..", "/") after cleaning
// - Allowed characters (alphanumeric, underscore, hyphen, period)
// - Maximum length
// - Null bytes
//
// Parameters:
//
//	path - The input path string to validate.
//
// Returns:
//
//	string - The validated simple filename (without "./") if validation succeeds.
//	error - An error detailing the validation failure if any check fails. On failure,
//	        the returned string is the original input path.
func validateFilePath(path string) (string, error) {
	// --- Validate the filename parameter ---
	if path == "" {
		err := errors.New("invalid file path: filename cannot be empty")
		// Return original path (empty) and error
		return path, err
	}

	// 1. Basic cleaning (removes ., .., extra slashes)
	validatedFilename := filepath.Clean(path)

	// 2. Enforce filename only (check for separators *after* cleaning)
	//    Also reject "." and ".." explicitly as filenames.
	if filepath.Base(validatedFilename) != validatedFilename || validatedFilename == "." ||
		validatedFilename == ".." {
		err := fmt.Errorf(
			"invalid file path: %q must be a filename only (no directory separators)",
			path, // Use original path in error message for clarity
		)
		// Return original path and error
		return path, err
	}

	// 3. Check for allowed characters using regex
	if !validFilenameChars.MatchString(validatedFilename) {
		err := fmt.Errorf(
			"invalid file path: filename %q contains invalid characters (allowed: a-z, A-Z, 0-9, _, -, .)",
			validatedFilename, // Use validated filename here as it's the one checked
		)
		// Return original path and error
		return path, err
	}

	// 4. Check filename length
	if len(validatedFilename) > maxFilenameLength {
		err := fmt.Errorf(
			"invalid file path: filename %q exceeds maximum length of %d",
			validatedFilename,
			maxFilenameLength,
		)
		// Return original path and error
		return path, err
	}

	// 5. Check for null bytes
	if strings.ContainsRune(validatedFilename, '\x00') {
		err := fmt.Errorf("invalid file path: filename %q contains null byte", validatedFilename)
		// Return original path and error
		return path, err
	}

	// If all checks pass, return the validated filename (which is just the base name) and nil error
	return validatedFilename, nil
}
