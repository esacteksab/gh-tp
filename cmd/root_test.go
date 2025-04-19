// SPDX-License-Identifier: MIT
package cmd

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

// It's possible that a person would run gh tp and no config file exists. We need to handle it.
func TestNoConfigFileFound(t *testing.T) {
	cmd := exec.Command("gh-tp", "--config", "")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
	}
}

// It's possible that both tofu and terraform exists on a person's $PATH. We need to handle it.
func TestDuplicateBinaries(t *testing.T) {
	cmd := exec.Command("gh-tp", "--config", "../testdata/duplicateBinaries/.tp.toml")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
	}
}

// It's possible a user doesn't define planFile in config file. We need to handle it.
func TestAbsentPlanFile(t *testing.T) {
	cmd := exec.Command("gh-tp", "--config", "../testdata/missingParameters/noPlanFile/.tp.toml")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
	}
}

// It's possible a user doesn't define mdFile in config file. We need to handle it.
func TestAbsentMdFile(t *testing.T) {
	cmd := exec.Command("gh-tp", "--config", "../testdata/missingParameters/noMdFile/.tp.toml")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
	}
}
