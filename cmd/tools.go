// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

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
			logger.Debugf("%s file %s was created.", v.Purpose, v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n", bold(green("✔")), v.Purpose, " Created...")
		} else if !exists {
			//
			logger.Errorf("Markdown file %s was not created.", v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n", bold(red("✕")), v.Purpose, " Failed to Create ...")
		} else {
			// I'm only human. NFC how you got here. I hope to never have to find out.
			logger.Errorf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
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

func getDirectories() (homeDir, configDir, cwd string, err error) {
	homeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	configDir, err = os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	cwd, cwderr := os.Getwd()
	if cwderr != nil {
		logger.Errorf("Error: %s", err)
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
