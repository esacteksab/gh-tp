// SPDX-License-Identifier: MIT
package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

// It's possible that a person would run gh tp and no config file exists. We need to handle it.
func TestNoConfigFileFound(t *testing.T) {
	cmd := exec.Command("gh", "tp")
	msg := "ERRO Config file not found. Please run 'gh tp init' or run 'gh tp help' or refer to the documentation on how to create a config file. https://github.com/esacteksab/gh-tp"

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := string(stdout.String()), string(stderr.String())

	if assert.Error(t, err) {
		exitError, ok := err.(*exec.ExitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(t, msg, strings.TrimSuffix(errStr, "\n"))
	}
}

// It's possible that both tofu and terraform exists on a person's $PATH. We need to handle it.
func TestDuplicateBinaries(t *testing.T) {
	cmd := exec.Command("gh", "tp", "--config", "../testdata/duplicateBinaries/.tp.toml")
	msg := "ERRO Found both `tofu` and `terraform` in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your ../testdata/duplicateBinaries/.tp.toml config file to the binary you want to use."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := string(stdout.String()), string(stderr.String())

	if assert.Error(t, err) {
		exitError, ok := err.(*exec.ExitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(t, msg, strings.TrimSuffix(errStr, "\n"))
	}
}
