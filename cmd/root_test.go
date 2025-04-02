// SPDX-License-Identifier: MIT
package cmd

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
)

// It's possible that a person would run gh tp and no config file exists. We need to handle it.
func TestNoConfigFileFound(t *testing.T) { //nolint:dupl
	cmd := exec.Command("gh-tp", "--config", "")
	msg := "ERRO Config file not found. Please run 'gh tp init' or run 'gh tp help' or refer to the documentation on how to create a config file. https://github.com/esacteksab/gh-tp"

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := stdout.String(), stderr.String()

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(
			t,
			msg,
			strings.TrimSuffix(errStr, "\n"),
		)
	}
}

// It's possible that both tofu and terraform exists on a person's $PATH. We need to handle it.
func TestDuplicateBinaries(t *testing.T) { //nolint:dupl
	cmd := exec.Command("gh-tp", "--config", "../testdata/duplicateBinaries/.tp.toml")
	msg := "ERRO Found both tofu and terraform in your $PATH. We're not sure which one to use. Please set the binary parameter in ../testdata/duplicateBinaries/.tp.toml to the binary you want to use."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := stdout.String(), stderr.String()

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(t, msg, strings.TrimSuffix(errStr, "\n"))
	}
}

// It's possible a user doesn't define planFile in config file. We need to handle it.
func TestAbsentPlanFile(t *testing.T) { //nolint:dupl
	cmd := exec.Command("gh-tp", "--config", "../testdata/missingParameters/noPlanFile/.tp.toml")
	msg := "ERRO Missing Parameter: 'planFile' (type: string) is not defined in ../testdata/missingParameters/noPlanFile/.tp.toml. This is the name of the plan's output file that will be created by `gh tp`."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := stdout.String(), stderr.String()

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(t, msg, strings.TrimSuffix(errStr, "\n"))
	}
}

// It's possible a user doesn't define mdFile in config file. We need to handle it.
func TestAbsentMdFile(t *testing.T) { //nolint:dupl
	cmd := exec.Command("gh-tp", "--config", "../testdata/missingParameters/noMdFile/.tp.toml")
	msg := "ERRO Missing Parameter: 'mdFile' (type: string) is not defined in ../testdata/missingParameters/noMdFile/.tp.toml. This is the name of the Markdown file that will be created by `gh tp`."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Errorf("cmd.Run() failed with %s\n", err)
	}
	_, errStr := stdout.String(), stderr.String()

	if assert.Error(t, err) {
		var exitError *exec.ExitError
		ok := errors.As(err, &exitError)
		assert.Error(t, err, "Expected an error.")
		assert.True(t, ok, "Expected *exec.ExitError, got: %T", err)
		assert.Equal(t, 1, exitError.ExitCode(), "Expected exit code 1")
		assert.Equal(t, msg, strings.TrimSuffix(errStr, "\n"))
	}
}
