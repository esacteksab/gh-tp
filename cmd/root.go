// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/cli/safeexec"
	"github.com/hashicorp/terraform-exec/tfexec"
	md "github.com/nao1215/markdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	binary     string
	binaries   []string
	cfgFile    string
	out        io.Reader
	mdParam    string
	planBody   *md.Markdown
	planMd     *os.File
	planPath   string
	planStr    string
	sb         strings.Builder
	sbPlan     string
	Verbose    bool
	Version    string
	Date       string
	Commit     string
	BuiltBy    string
	exts       []string
	workingDir string
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
		result = fmt.Sprintf(
			"%s\nmodule version: %s, checksum: %s",
			result,
			info.Main.Version,
			info.Main.Sum,
		)
	}
	return result
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Version: buildVersion(Version, Commit, Date, BuiltBy),
	Use:     "tp",
	Short:   "A GitHub CLI extension to submit a pull request with Terraform or Tofu plan output.",
	Long:    `tp is a GitHub CLI extension to submit a pull request with Terraform or Tofu plan output formatted in GitHub Flavored Markdown.`,
	Run: func(cmd *cobra.Command, args []string) {
		b := viper.IsSet("binary")
		if b {
			binary = viper.GetString("binary")
		} else {
			exists := []string{}
			binaries = []string{"tofu", "terraform"}
			for _, v := range binaries {
				bin, err := safeexec.LookPath(v)
				if err != nil {
					log.Debugf("%s", err)
				}
				// It's possible for both `tofu` and `terraform` to exist on $PATH and we need to handle that.
				if len(bin) > 0 {
					exists = append(exists, bin)
				}
			}
			if len(exists) == len(binaries) {
				log.Fatal("Seems both `tofu` and `terraform` exist in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your .tp.toml config file to whichever binary you want to use.")
			}
		}

		// the arg received looks like a file, we try to open it
		if len(args) == 0 {
			execPath, err := safeexec.LookPath(binary)
			if err != nil {
				log.Fatal(
					"Please ensure either `tofu` or `terraform` are installed and on your $PATH.",
				)
				// os.Exit(1)
			}

			workingDir = filepath.Base(".")
			// Initialize tf -- NOT terraform init
			tf, err := tfexec.NewTerraform(workingDir, execPath)
			if err != nil {
				log.Fatalf("error calling binary: %s\n", err)
			}

			//Check for .terraform.lock.hcl -- do not need to do this every time
			//terraform init | installs providers, etc.
			//err = tf.Init(context.Background())
			//if err != nil {
			//	log.Fatalf("error running Init: %s", err)
			//}

			// the plan file
			planPath = viper.GetString("planFile")
			planOpts := []tfexec.PlanOption{
				// terraform plan --out planPath (plan.out)
				tfexec.Out(planPath),
			}

			exts = []string{".tf", ".tofu"}
			files := checkFilesByExtension(workingDir, exts)
			// we check to see if there are tf or tofu files in the current working directory. If not, we don't call tf.plan
			if files {
				log.Infof("Creating %s plan file %s...", binary, planPath)
				// terraform plan -out plan.out -no-color
				_, err := tf.Plan(context.Background(), planOpts...)
				if err != nil {
					// binary defined. .tf or .tofu files exist. Still errors. Show me the error
					log.With("err", err).Errorf("%s returned the follow error", binary)
					// Edge case exists where we detect .tofu file but terraform was called,
					// which doesn't support .tofu files. tf.Plan returns error.
					// There is a condition that exists where .tofu files exist, but terraform
					// is the binary, this error will occur. But we're not checking _explicitly_
					// for either .tf or .tofu in files above.
					// So .tf files _could_ exist, but tf.Plan could fail for
					// some reason not related to Terraform not finding any .tf files,
					// making this error inaccurate. Could be nice to identify and
					// handle this edge case, but Terraform/Tofu do it good enough for now.
					// if binary == "terraform" {
					// 	log.Infof("Detected `*.tofu` files, but you've defined %s
					// as the binary to use in your .tp.toml config file. Terraform does not support `.tofu` files.", binary)
					// }
					// We need to exit on this error. tf.Plan actually returns status 1
					// -- maybe some day we can intercept it or have awareness that it was returned.
					log.Infof("Check the output of `%s plan` locally. If you believe this is a bug, please report the issue. TPE001.", binary)
					os.Exit(1)
				}

				planStr, err = tf.ShowPlanFileRaw(context.Background(), planPath)
				if err != nil {
					log.Error(
						"error internally attempting to create the human-readable Plan: ",
						err,
					)
				}

				log.Debug((planStr))

				// fmt.Printf("plan output: %s", planStr)
				mdParam = viper.GetString("mdFile")

				planMd, err = os.Create(mdParam)
				if err != nil {
					log.Errorf("failed to create Markdown: %s", err)
				}
				// Close the file when we're done with it
				defer planMd.Close()

				// This has the plan wrapped in a code block in Markdown

				planBody = md.NewMarkdown(os.Stdout).
					CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
				if err != nil {
					log.Errorf("error generating plan Markdown: %s", err)
				}

				// NewMarkdown returns io.Writer
				fmt.Fprintf(&sb, "\n%s\n", planBody)

				// This turns NewMarkdown io.Writer into a String, which .Details expects
				sbPlan = sb.String()

				// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
				mderr := md.NewMarkdown(planMd).Details("Terraform Plan", sbPlan).Build()
				if mderr != nil {
					log.Errorf("error generating %s markdown file, error: %s", mdParam, err)
				}

				// Checking to see if plan file was created.
				if _, err := os.Stat(planPath); err == nil {
					log.Infof("Plan file %s was created.", planPath)
				} else if errors.Is(err, os.ErrNotExist) {
					// Apparently the binary exists, tf.Plan shit the bed and didn't tell us.
					log.Errorf("Plan file %s was not created.", planPath)
				} else {
					// I'm only human. NFC how you got here. I hope to never have to find out.
					log.Infof("If you see this error message, please open a bug. Error Code: TPE002. Error: %s", err)
				}

				// Checking to see if Markdown file was created.
				if _, err := os.Stat(mdParam); err == nil {
					log.Infof("Markdown file %s was created.", mdParam)
				} else if errors.Is(err, os.ErrNotExist) {
					//
					log.Errorf("Markdown file %s was not created.", mdParam)
				} else {
					// I'm only human. NFC how you got here. I hope to never have to find out.
					log.Infof("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
				}
			} else {
				log.Errorf("No %s files found. Please run this in a directory with %s files present.", cases.Title(language.English).String(binary), cases.Title(language.English).String(binary))
			}

		} else if args[0] == "-" {
			out = cmd.InOrStdin()
			content, err := io.ReadAll(out)
			if err != nil {
				log.Fatalf("unable to read stdIn: %s", err)
			}

			planStr := string(content)
			log.Debugf("plan output: %s", planStr)

			mdParam = viper.GetString("mdFile")

			planMd, err := os.Create(mdParam)
			if err != nil {
				log.Fatalf("failed to create Markdown: %s", err)
			}
			// Close the file when we're done with it
			defer planMd.Close()

			// This has the plan wrapped in a code block in Markdown
			planBody := md.NewMarkdown(os.Stdout).CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
			if err != nil {
				log.Fatalf("error generating plan Markdown: %s", err)
			}

			// NewMarkdown returns io.Writer
			fmt.Fprintf(&sb, "\n%s\n", planBody)

			// This turns NewMarkdown io.Writer into a String, which .Details expects
			sbPlan := sb.String()

			// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
			mderr := md.NewMarkdown(planMd).Details("Terraform Plan", sbPlan).Build()
			if mderr != nil {
				log.Errorf("error generating %s markdown file, error: %s", mdParam, err)
			}
			// Checking to see if Markdown file was created.
			if _, err := os.Stat(mdParam); err == nil {
				log.Infof("Markdown file %s was created.", mdParam)
			} else if errors.Is(err, os.ErrNotExist) {
				//
				log.Fatalf("Markdown file %s was not created.", mdParam)
			} else {
				// I'm only human. NFC how you got here. I hope to never have to find out.
				log.Fatalf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
			}
		}
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

	rootCmd.Flags().
		StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tp.toml, can also exist in your project's root directory.)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name ".tp.toml"
		viper.SetConfigName(".tp.toml")
		viper.SetConfigType("toml")
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Error(
				"Missing Config File: Config file should be named .tp.toml and exist in your home directory or in your project's root.\n",
			)
			os.Exit(1)
		} else if _, ok := err.(viper.UnsupportedConfigError); ok {
			log.Errorf("Unsupported Format. Config file should be named .tp %s.", err)
			os.Exit(1)
			// This handles the situation where a duplicate key exists.
		} else if _, ok := err.(viper.ConfigParseError); ok {
			log.Errorf("There is an issue %s.", err)
			os.Exit(1)
		}
	}

	keys := viper.AllKeys()
	log.Debug(keys)

	// // Check to see if required 'planFile' parameter is set
	o := viper.IsSet("planFile")
	if !o {
		log.Error(
			"Missing Parameter: 'planFile' (type: string) is not defined in the config file. This is the name of the plan's output file that will be created by `gh tp`.\n",
		)
		os.Exit(1)
	}

	// // Check to see if required 'mdFile' parameter is set
	m := viper.IsSet("mdFile")
	if !m {
		log.Error(
			"Missing Parameter: 'mdFile' (type: string) is not defined in the config file. This is the name of the Markdown file that will be created by `gh tp`.\n",
		)
		os.Exit(1)
	}
	log.Infof("Using config file: %s", viper.ConfigFileUsed())
}
