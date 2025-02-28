/*
Copyright © 2025 Barry Morrison

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/hashicorp/terraform-exec/tfexec"
	md "github.com/nao1215/markdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	out         io.Reader
	flagNoColor bool
	planStr     string
	sb          strings.Builder
	Verbose     bool
	Version     string
	Date        string
	Commit      string
	BuiltBy     string
	bold        = color.New(color.Bold).SprintFunc()
	hiBlack     = color.New(color.FgHiBlack).SprintFunc()
	green       = color.New(color.FgGreen).SprintFunc()
	red         = color.New(color.FgRed).SprintFunc()
)

type SyntaxHighlight string

const (
	SyntaxHighlightTerraform SyntaxHighlight = "terraform"
)

func buildVersion(version, commit, date, builtBy string) string {
	result := version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}
	return result
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: buildVersion(Version, Commit, Date, BuiltBy),
	Use:     "tp [file]",
	Short:   "do stuff with a [file]",
	Long:    `With no FILE, or when FILE is '-', read stdIn.`,
	Run: func(cmd *cobra.Command, args []string) {
		flagNoColor = viper.GetBool("no-color")
		if flagNoColor {
			color.NoColor = true // disables colorized output
		}
		binary := viper.GetString("binary")

		// the arg received looks like a file, we try to open it
		if len(args) == 0 {
			execPath, err := safeexec.LookPath(binary)
			if err != nil {
				log.Fatal("Please ensure terraform or tofu are installed and on your $PATH")
				os.Exit(1)
			}

			workingDir := filepath.Base(".")
			// Initialize tf -- NOT terraform init
			tf, err := tfexec.NewTerraform(workingDir, execPath)
			if err != nil {
				log.Fatalf("error running NewTerraform: %s\n", err)
			}

			//Check for .terraform.lock.hcl -- do not need to do this every time
			//terraform init | installs providers, etc.
			err = tf.Init(context.Background())
			if err != nil {
				log.Fatalf("error running Init: %s", err)
			}

			//the plan file
			planPath := viper.GetString("planfile")
			planOpts := []tfexec.PlanOption{
				// terraform plan --out planPath (plan.out)
				tfexec.Out(planPath),
			}

			// terraform plan -out plan.out -no-color
			hasChanges, err := tf.Plan(context.Background(), planOpts...)
			if err != nil {
				log.Fatalf("error running Plan: %s", err)
			}
			// if tf.Plan has no changes, hasChanges is empty or false
			if hasChanges {
				// This is the actual plan output from terraform plan -out plan.out -no-color
				planStr, err = tf.ShowPlanFileRaw(context.Background(), planPath)
				if err != nil {
					log.Fatalf("error creating Plan: %s", err)
				}
				if Verbose {
					log.Println((planStr))
				}
				//fmt.Printf("plan output: %s", planStr)
				planmd := viper.GetString("mdfile")
				plmd, err := os.Create(planmd)
				if err != nil {
					log.Fatalf("failed to create Markdown: %s", err)
				}
				// Close the file when we're done with it
				defer plmd.Close()
				// This has the plan wrapped in a code block in Markdown
				plbody := md.NewMarkdown(os.Stdout).CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
				if err != nil {
					log.Fatalf("error generating plan Markdown: %s", err)
				}
				// NewMarkdown returns io.Writer
				fmt.Fprintf(&sb, "\n%s\n", plbody)
				// This turns NewMarkdown io.Writer into a String, which .Details expects
				sbplan := sb.String()
				// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
				md.NewMarkdown(plmd).Details("Terraform Plan", sbplan).Build()
			} else {
				log.Println(bold(green("No changes."), "Your infrastructure matches the configuration."))
			}
			if _, err := os.Stat(planPath); err == nil {
				// Only tell me about it if -v is passed
				if Verbose {
					log.Printf("%s was created", planPath)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				// Apparently the binary exists, tf.Plan shit the bed and didn't tell us
				log.Fatalf("%s was not created", planPath)
			} else {
				// I'm only human. NFC how you got here. I hope to never have to find out
				log.Println("No F'n Clue How you got here.")
			}
		} else if args[0] == "-" {
			out = cmd.InOrStdin()
			content, err := io.ReadAll(out)
			if err != nil {
				log.Fatalf("unable to read stdIn: %s", err)
			}

			planStr := string(content)
			if Verbose {
				fmt.Printf("plan output: %s", planStr)
			}
			planmd := viper.GetString("mdfile")
			plmd, err := os.Create(planmd)
			if err != nil {
				log.Fatalf("failed to create Markdown: %s", err)
			}
			// Close the file when we're done with it
			defer plmd.Close()
			// This has the plan wrapped in a code block in Markdown
			plbody := md.NewMarkdown(os.Stdout).CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
			if err != nil {
				log.Fatalf("error generating plan Markdown: %s", err)
			}
			// NewMarkdown returns io.Writer
			fmt.Fprintf(&sb, "\n%s\n", plbody)
			// This turns NewMarkdown io.Writer into a String, which .Details expects
			sbplan := sb.String()
			// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
			md.NewMarkdown(plmd).Details("Terraform Plan", sbplan).Build()
		}
		//fmt.Println("If you got here, W.T.F")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tp)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gh-tp" (without extension).
		viper.SetConfigName(".tp")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
