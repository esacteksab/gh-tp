// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/cli/safeexec"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/viper"
)

func createPlan() (planStr string, err error) {
	execPath, err := safeexec.LookPath(binary)
	if err != nil {
		Logger.Fatal(
			"Please ensure either `tofu` or `terraform` are installed and on your $PATH.",
		)
		// os.Exit(1)
	}

	workingDir = filepath.Base(".")
	// Initialize tf -- NOT terraform init
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		Logger.Fatalf("error calling binary: %s\n", err)
	}

	// Check for .terraform.lock.hcl -- do not need to do this every time
	// terraform init | installs providers, etc.
	// err = tf.Init(context.Background())
	// if err != nil {
	//	Logger.Fatalf("error running Init: %s", err)
	// }

	// the plan file
	planPath = viper.GetString("planFile")
	planOpts := []tfexec.PlanOption{
		// terraform plan --out planPath (plan.out)
		tfexec.Out(planPath),
	}

	mdParam = viper.GetString("mdFile")

	Logger.Debugf("Creating %s plan file %s...", binary, planPath)
	// terraform plan -out plan.out -no-color
	spinnerDuration = 100
	s := spinner.New(spinner.CharSets[14], spinnerDuration*time.Millisecond)
	s.Suffix = "  Creating the Plan...\n"
	s.Start()

	_, err = tf.Plan(context.Background(), planOpts...)
	if err != nil {
		// binary defined. .tf or .tofu files exist. Still errors. Show me the error
		Logger.With("err", err).Errorf("%s returned the follow error", binary)
		// There is a condition that exists where .tofu files exist, but terraform
		// is the binary, this error will occur. But we're not checking _explicitly_
		// for either .tf or .tofu in files above.
		// So .tf files _could_ exist, but tf.Plan could fail for
		// some reason not related to Terraform not finding any .tf files,
		// making this error inaccurate. Could be nice to identify and
		// handle this edge case, but Terraform/Tofu do it good enough for now.
		// if binary == "terraform" {
		// 	Logger.Infof("Detected `*.tofu` files, but you've defined %s
		// as the binary to use in your .tp.toml config file. Terraform does not support `.tofu` files.", binary)
		// }
		// We need to exit on this error. tf.Plan actually returns status 1
		// -- maybe some day we can intercept it or have awareness that it was returned.
		Logger.Infof(
			"Check the output of `%s plan` locally. If you believe this is a bug, please report the issue. TPE001.",
			binary,
		)
		os.Exit(1)
	}
	s.Stop()

	planStr, err = tf.ShowPlanFileRaw(context.Background(), planPath)
	if err != nil {
		Logger.Error(
			"error internally attempting to create the human-readable Plan: ",
			err,
		)
	}

	Logger.Debugf("Plan output is: \n%s\n", planStr)

	return planStr, err
}
