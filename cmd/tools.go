// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/charmbracelet/huh"
	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"
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

func existsOrCreated(files []tpFile) error {
	for _, v := range files {
		if _, err := os.Stat(v.Name); err == nil {
			logger.Debugf("%s file %s was created.", v.Purpose, v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n", bold(green("✔")), v.Purpose, " Created...")
		} else if errors.Is(err, os.ErrNotExist) {
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

// Feels like a bit of a duplicate to the above function, this takes a path string an returns a bool
// on whether or not the path exists -- TODO #66 probably worth deduping this eventually
func doesNotExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

func getDirectories() (homeDir, configDir, cwd string, err error) {
	homeDir = xdg.Home

	configDir = xdg.ConfigHome

	cwd, cwderr := os.Getwd()
	if cwderr != nil {
		logger.Errorf("Error: %s", err)
	}
	return homeDir, configDir, cwd, err
}

func genConfig(conf ConfigParams) (data []byte, err error) {
	data, err = toml.Marshal(conf)
	if err != nil {
		logger.Fatalf("Failed marshalling TOML: %s", err)
	}
	return data, err
}

// takes cfgFile, appends ".tp.toml" to it
// checks to see if file exists
// based on existence, asks to create (doesn't exist)
// or overwrite (exists)
func mkFile(cfgFile string) (exists, createFile bool) {
	configName = ".tp.toml"
	noConfig := doesNotExist(cfgFile + "/" + configName)
	logger.Debug(cfgFile + configName)
	if noConfig {
		logger.Debugf("%s/%s doesn't exist\n", cfgFile, configName)
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Create new file?").
					Affirmative("Yes").
					Negative("No").
					Value(&createFile),
			)).WithTheme(huh.ThemeBase16()).Run()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Debugf("Inside mkFile() and config is %s/%s\n", cfgFile, configName)
		return noConfig, createFile
	} else if !noConfig {
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Overwrite existing config file?").
					Affirmative("Yes").
					Negative("No").
					Value(&createFile),
			)).WithTheme(huh.ThemeBase16()).Run()
		if err != nil {
			logger.Fatal(err)
		}
		logger.Debugf("Inside mkFile if exists, config is %s/%s\n", cfgFile, configName)
		return noConfig, createFile
	}

	logger.Debugf("inside mkFile() noConfig is %t\n", noConfig)
	logger.Debugf("Inside mkFile() config is %s/%s\n", cfgFile, configName)
	return noConfig, createFile
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
