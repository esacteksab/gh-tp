// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/viper"
)

func createPlan() (planStr string, err error) {
	// --- Parameter Validation & Setup ---
	workingDir := "."
	tfBinaryPath := viper.GetString("binary")
	if tfBinaryPath == "" {
		tfBinaryPath = binary
		if tfBinaryPath == "" {
			return "", errors.New("binary not configured")
		}
	}
	pf := viper.GetString("planFile")
	planPath, err := validateFilePath(pf)
	if err != nil {
		return "", fmt.Errorf("invalid 'planFile' (%q): %w", pf, err)
	}

	tf, err := tfexec.NewTerraform(workingDir, tfBinaryPath)
	if err != nil {
		return "", fmt.Errorf("tfexec init failed: %w", err)
	}
	// _ = tf.SetWaitDelay(60 * time.Second)
	planOpts := []tfexec.PlanOption{tfexec.Out(planPath)}

	// --- Signal Handling & Atomic Flag ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var interrupted atomic.Bool

	cleanupSignalResources := func() {
		Logger.Debug("Attempting signal resource cleanup...")
		signal.Stop(sigChan)
		select {
		case <-sigChan:
			Logger.Debug("Drained signal during cleanup.")
		default:
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					Logger.Debugf("Recovered from closing already closed sigChan: %v", r)
				}
			}()
			close(sigChan)
			Logger.Debug("Signal channel closed.")
		}()
		Logger.Debug("Signal handler resources cleanup finished.")
	}

	go func() {
		defer Logger.Debug("Signal listener goroutine finished.")
		sig, ok := <-sigChan
		if ok {
			Logger.Warnf("Signal %v received by Go process. Setting interruption flag.", sig)
			interrupted.Store(true)
		} else {
			Logger.Debug("Signal channel closed while listener goroutine was active.")
		}
	}()

	// --- Execute Terraform Plan ---
	Logger.Debugf(
		"Running %s plan (outputting to %s)...",
		tfBinaryPath,
		planPath,
	)
	s := spinner.New(spinner.CharSets[14], spinnerDuration)
	s.Suffix = " Creating Plan..."
	s.Start()

	planCtx := context.Background()
	_, err = tf.Plan(planCtx, planOpts...)

	// --- Handle Plan Result ---
	if interrupted.Load() {
		s.Stop()
		Logger.Warnf("Interruption flag set. Terraform process likely interrupted.")

		cleanupSignalResources()
		Logger.Debugf("[DIAG] Skipping signal cleanup call for test.")
		Logger.Debugf("[DIAG] About to return ErrInterrupted from createPlan.")

		return "", ErrInterrupted // Return the specific error
	}

	// Handle other errors
	if err != nil {
		s.Stop()
		Logger.Errorf("tf.Plan finished with non-interruption error. Type: %T, Value: %v", err, err)
		cleanupSignalResources()
		// Presumably an unusable plan, so let's clean things up -- we may not want this long-term or maybe make this a parameter
		_ = os.Remove(planPath) // Attempt cleanup for other errors
		return "", fmt.Errorf("terraform plan failed: %w", err)
	}

	// --- Plan Successful ---
	s.Stop()
	cleanupSignalResources()
	Logger.Debug("Terraform plan completed successfully.")

	// --- Show Plan Output ---
	Logger.Debug("Generating plan output...")
	showCtx, showCancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:mnd
	defer showCancel()
	planStr, err = tf.ShowPlanFileRaw(showCtx, planPath)
	if err != nil {
		Logger.Errorf("Plan created, but failed to read/show plan file %q: %v", planPath, err)
		return "", fmt.Errorf("failed to show plan file %q: %w", planPath, err)
	}

	Logger.Debug("Plan output generated successfully.")
	return planStr, nil
}
