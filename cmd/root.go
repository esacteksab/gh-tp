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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	bin         string
	binary      string
	binaries    []string
	cfgFile     string
	out         io.Reader
	flagNoColor bool
	mdParam     string
	planBody    *md.Markdown
	planMd      *os.File
	planPath    string
	planStr     string
	sb          strings.Builder
	sbPlan      string
	Verbose     bool
	Version     string
	Date        string
	Commit      string
	BuiltBy     string
	bold        = color.New(color.Bold).SprintFunc()
	hiBlack     = color.New(color.FgHiBlack).SprintFunc()
	green       = color.New(color.FgHiGreen).SprintFunc()
	yellow      = color.New(color.FgHiYellow).SprintFunc()
	red         = color.New(color.FgHiRed).SprintFunc()
	exts        []string
	workingDir  string
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
					if Verbose {
						log.Printf("%s", err)
					}
				}
				// It's possible for both `tofu` and `terraform` to exist on $PATH and we need to handle that.
				if len(bin) > 0 {
					exists = append(exists, bin)
				}
			}
			if len(exists) == len(binaries) {
				log.Fatal(bold(red("Ooops! ")), "Seems both `tofu` and `terraform` exist in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your .tp.toml config file to whichever binary you want to use.")
			}
			//cmd.Help()
			//fmt.Println("foo")
		}

		// the arg received looks like a file, we try to open it
		if len(args) == 0 {
			execPath, err := safeexec.LookPath(binary)
			if err != nil {
				log.Fatal(bold(red("Attention! ")), "Please ensure either `tofu` or `terraform` are installed and on your $PATH.")
				//os.Exit(1)
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

			//the plan file
			planPath = viper.GetString("planFile")
			planOpts := []tfexec.PlanOption{
				// terraform plan --out planPath (plan.out)
				tfexec.Out(planPath),
			}

			exts = []string{".tf", ".tofu"}
			files := checkFilesByExtension(workingDir, exts)
			// we check to see if there are tf or tofu files in the current working directory. If not, we don't call tf.plan
			if files {
				// terraform plan -out plan.out -noColor
				_, err := tf.Plan(context.Background(), planOpts...)
				if err != nil {
					// binary defined. .tf or .tofu files exist. Still errors. Show me the error
					log.Println(bold(red("Terraform returned the following error: ")), err)
					// Edge case exists where we detect .tofu file but terraform was called, which doesn't support .tofu files. tf.Plan returns error.
					// BUG: There is a condition that exists where .tofu files exist, but terraform is the binary, this error will occur. But we're not checking _explicitly_ for either .tf or .tofu in files above.
					// So .tf files _could_ exist, but tf.Plan could fail for some reason not related to Terraform not finding any .tf files, making this error inaccurate. Could be nice to identify and handle this edge case, but Terraform/Tofu do it good enough for now.
					// if binary == "terraform" {
					// 	log.Printf("Detected `*.tofu` files, but you've defined %s as the binary to use in your .tp.toml config file. Terraform does not support `.tofu` files.", binary)
					// }
					// We need to exit on this error. tf.Plan actually returns status 1 -- maybe some day we can intercept it or have awareness that it was returned.
					os.Exit(1)
				}

				planStr, err = tf.ShowPlanFileRaw(context.Background(), planPath)
				if err != nil {
					log.Fatal("error internally attempting to create the human-readable Plan: ", err)
				}

				if Verbose {
					log.Println((planStr))
				}

				//fmt.Printf("plan output: %s", planStr)
				mdParam = viper.GetString("mdFile")

				planMd, err = os.Create(mdParam)
				if err != nil {
					log.Fatalf("failed to create Markdown: %s", err)
				}
				// Close the file when we're done with it
				defer planMd.Close()

				// This has the plan wrapped in a code block in Markdown
				planBody = md.NewMarkdown(os.Stdout).CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
				if err != nil {
					log.Fatalf("error generating plan Markdown: %s", err)
				}

				// NewMarkdown returns io.Writer
				fmt.Fprintf(&sb, "\n%s\n", planBody)

				// This turns NewMarkdown io.Writer into a String, which .Details expects
				sbPlan = sb.String()

				// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
				md.NewMarkdown(planMd).Details("Terraform Plan", sbPlan).Build()

				// Checking to see if plan file was created.
				if _, err := os.Stat(planPath); err == nil {
					log.Printf("Plan file %s was created.", planPath)

				} else if errors.Is(err, os.ErrNotExist) {

					// Apparently the binary exists, tf.Plan shit the bed and didn't tell us.
					log.Fatalf("Plan file %s was not created.", planPath)

				} else {

					// I'm only human. NFC how you got here. I hope to never have to find out.
					log.Printf("If you see this error message, please open a bug. Error Code: TPE002. Error: %s", err)
				}

				// Checking to see if Markdown file was created.
				if _, err := os.Stat(mdParam); err == nil {
					log.Printf("Markdown file %s was created.", mdParam)

				} else if errors.Is(err, os.ErrNotExist) {

					//
					log.Fatalf("Markdown file %s was not created.", mdParam)

				} else {

					// I'm only human. NFC how you got here. I hope to never have to find out.
					log.Printf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
				}
			} else {
				log.Fatalf("No %s files found. Please run this in a directory with %s files present.", cases.Title(language.English).String(binary), cases.Title(language.English).String(binary))
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
			fmt.Printf("Plan output: %s", planStr)

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
			md.NewMarkdown(planMd).Details("Terraform Plan", sbPlan).Build()

			// Checking to see if Markdown file was created.
			if _, err := os.Stat(mdParam); err == nil {
				log.Printf("Markdown file %s was created.", mdParam)

			} else if errors.Is(err, os.ErrNotExist) {

				//
				log.Fatalf("Markdown file %s was not created.", mdParam)

			} else {

				// I'm only human. NFC how you got here. I hope to never have to find out.
				log.Printf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
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

	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tp.toml, can also exist in your project's root directory.)")
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
			log.Fatal(bold(red("Attention! Missing Config File: "), "Config file should be named .tp.toml and exist in your home directory or in your project's root.\n"))
			os.Exit(1)
		} else if _, ok := err.(viper.UnsupportedConfigError); ok {
			log.Fatalf("Unsupported Format. Config file should be named .tp %s", err)
			// This handles the situation where a duplicate key exists.
		} else if _, ok := err.(viper.ConfigParseError); ok {
			log.Fatalf("There is an issue with parsing your config file, the error is error: %s", err)
		}
		//if Verbose {
		//fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		//}
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	flagNoColor = viper.GetBool("noColor")
	if flagNoColor {
		color.NoColor = true // disables colorized output
	}
	if Verbose {
		keys := viper.AllKeys()
		log.Println(keys)
	}
	// Validate that required 'binary' parameter is set
	b := viper.IsSet("binary")
	if !b {
		log.Print(bold(red("Attention! Missing Parameter: "), bold("'binary':"), " (type: string) is not defined in the config file. The value of the binary parameter should be either 'tofu' or 'terraform'. This binary is expected to exist on your $PATH.\n"))
	}

	// // Check to see if required 'planFile' parameter is set
	o := viper.IsSet("planFile")
	if !o {
		log.Fatal(bold(red("Attention! Missing Parameter: "), bold("'planFile':"), " (type: string) is not defined in the config file. This is the name of the plan's output file that will be created by `gh tp`.\n"))
	}

	// // Check to see if required 'mdFile' parameter is set
	m := viper.IsSet("mdFile")
	if !m {
		log.Fatal(bold(red("Attention! Missing Parameter: "), bold("'mdFile':"), " (type: string) is not defined in the config file. This is the name of the Markdown file that will be created by `gh tp`.\n"))
	}
}
