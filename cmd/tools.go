// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
)

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

// existsOrCreate checks to see if a plan file and markdown exist or were
// created Prints to terminal stating so
func existsOrCreated(files []tpFile) error {
	for _, v := range files {
		exists := doesExist(v.Name)
		if !exists {
			Logger.Debugf(
				"%s file %s was not created.", v.Purpose, v.Name,
			)
			fmt.Fprintf(
				color.Output, "%s  %s%s\n",
				bold(red("✕")), v.Purpose, " Failed to Create",
			)
		} else {
			Logger.Debugf("%s file %s was created.", v.Purpose, v.Name)
			fmt.Fprintf(
				color.Output, "%s  %s%s\n",
				bold(green("✔")), v.Purpose, " Created...",
			)
		}
	}
	return nil
}

// doesExist takes a path string and returns a bool on whether the path exists
func doesExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
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

// backupFile copies the file at source to dest we use this when creating a
// config file. if a existing config file is present, we back it up prior
// to overwriting it.
func backupFile(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			Logger.Error(err)
		}
	}(srcFile)

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		err := destFile.Close()
		if err != nil {
			Logger.Error(err)
		}
	}(destFile)

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	err = destFile.Sync()
	return err
}

// initLogger is how we initially create Logger. The values passed are based on
// 'Verbose' being true Colors are defined here
// https://github.com/charmbracelet/x/blob/aedd0cd23ed703ff7cbccc5c4f9ab51a4768a9e6/ansi/color.go#L15-L32
// 14 is Bright Cyan, 9 is Red -- no more purple
func initLogger(
	ReportCaller, ReportTimestamp bool, TimeFormat string,
) (Logger *log.Logger) {
	Logger = log.NewWithOptions(
		os.Stderr, log.Options{
			ReportCaller:    ReportCaller,
			ReportTimestamp: ReportTimestamp,
			TimeFormat:      TimeFormat,
		},
	)
	MaxWidth := 4
	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper(log.DebugLevel.String())).
		Bold(true).MaxWidth(MaxWidth).Foreground(lipgloss.Color("14"))

	styles.Levels[log.FatalLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper(log.FatalLevel.String())).
		Bold(true).MaxWidth(MaxWidth).Foreground(lipgloss.Color("9"))
	Logger.SetStyles(styles)
	Logger.SetLevel(log.DebugLevel)
	log.SetDefault(Logger)
	return Logger
}

func createLogger(verbose bool) {
	Verbose = verbose
	if Verbose {
		Logger = initLogger(true, true, "2006/01/02 15:04:05")
		log.SetLevel(log.DebugLevel)
		log.SetDefault(Logger)
	} else {
		Logger = initLogger(false, false, time.Kitchen)
		log.SetLevel(log.InfoLevel)
		log.SetDefault(Logger)
	}
}
