// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// ErrInterrupted indicates that the operation was cancelled by the user (e.g., Ctrl+C).
var ErrInterrupted = errors.New("operation interrupted by user")

// buildNoBinaryFoundError constructs the error message when no binary is found.
func buildNoBinaryFoundError() error {
	configPath := viper.ConfigFileUsed()
	errMsg := "could not find 'tofu' or 'terraform' in your PATH"
	if configPath != "" && doesExist(configPath) {
		errMsg += " and 'binary' not set in " + configPath
	} else if configPath == "" || !doesExist(configPath) {
		errMsg += ". Please install one, specify with -b, or set 'binary' in %s (if using config)" + ConfigName
	}
	return errors.New(errMsg)
}

// buildMultipleBinariesFoundError constructs the error message when multiple binaries are found.
func buildMultipleBinariesFoundError(foundBinaries []string) error {
	configPath := viper.ConfigFileUsed()
	errMsg := fmt.Sprintf("found both %s in your PATH", strings.Join(foundBinaries, " and "))
	if configPath != "" && doesExist(configPath) {
		errMsg += ". Specify the desired one using the -b flag or set the 'binary' parameter in " + configPath
	} else {
		errMsg += fmt.Sprintf(". Specify the desired one using the -b flag or create %s and set the 'binary' parameter", ConfigName)
	}
	return errors.New(errMsg)
}
