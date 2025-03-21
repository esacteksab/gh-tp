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

// existsOrCreate checks to see if a plan file and markdown exist or were created
// Prints to terminal stating so
func existsOrCreated(files []tpFile) error {
	for _, v := range files {
		exists := doesExist(v.Name)
		if exists {
			Logger.Debugf("%s file %s was created.", v.Purpose, v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n",
				bold(green("✔")), v.Purpose, " Created...")
		} else if !exists {
			//
			Logger.Debugf("%s file %s was not created.", v.Purpose, v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n",
				bold(red("✕")), v.Purpose, " Failed to Create")
		} else {
			// I'm only human. NFC how you got here. I hope to never have to find out.
			Logger.Errorf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
		}
	}
	return err
}

// doesExist takes a path string and returns a bool
// on whether or not the path exists
func doesExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// getDirectories returns User's home directory, $XDG_CONFIG_HOME
// and Current Working Directory
func getDirectories() (homeDir, configDir, cwd string, err error) {
	homeDir, err = os.UserHomeDir()
	if err != nil {
		Logger.Fatal(err)
	}

	configDir, err = os.UserConfigDir()
	if err != nil {
		Logger.Fatal(err)
	}

	cwd, cwderr := os.Getwd()
	if cwderr != nil {
		Logger.Errorf("Error: %s", err)
	}
	return homeDir, configDir, cwd, err
}

// backupFile copies the file at source to dest
// we use this when creating a config file.
// if a existing config file is present, we back it up
// prior to overwriting it.
func backupFile(source, dest string) error {
	srcFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	err = destFile.Sync()
	return err
}

// initLogger is how we initially create Logger. The values passed are based on 'Verbose' being true
// Colors are defined here https://github.com/charmbracelet/x/blob/aedd0cd23ed703ff7cbccc5c4f9ab51a4768a9e6/ansi/color.go#L15-L32
// 14 is Bright Cyan, 9 is Red -- no more purple
func initLogger(ReportCaller, ReportTimestamp bool, TimeFormat string) (Logger *log.Logger) {
	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    ReportCaller,
		ReportTimestamp: ReportTimestamp,
		TimeFormat:      TimeFormat,
	})
	MaxWidth = 4
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
