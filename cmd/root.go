// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/log"
	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	binary             string
	binaries           []string
	cfgFile            string
	out                io.Reader
	mdParam            string
	spinnerDuration    time.Duration
	titleCaseConverter cases.Caser
	Verbose            bool
	Version            string
	Date               string
	Commit             string
	BuiltBy            string
	exts               []string
	workingDir         string
	bold               = color.New(color.Bold).SprintFunc()
	green              = color.New(color.FgGreen).SprintFunc()
	red                = color.New(color.FgRed).SprintFunc()
)

func buildVersion(Version, Commit, Date, BuiltBy string) string {
	result := Version
	if Commit != "" {
		result = fmt.Sprintf("%s\nCommit: %s", result, Commit)
	}
	if Date != "" {
		result = fmt.Sprintf("%s\nBuilt at: %s", result, Date)
	}
	if BuiltBy != "" {
		result = fmt.Sprintf("%s\nBuilt by: %s", result, BuiltBy)
	}
	result = fmt.Sprintf("%s\nGOOS: %s\nGOARCH: %s", result, runtime.GOOS, runtime.GOARCH)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf(
			"%s\nmodule Version: %s, checksum: %s",
			result,
			info.Main.Version,
			info.Main.Sum,
		)
	}
	return result
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tp",
	Short: "A GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output.",
	Long:  `tp is a GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output formatted in GitHub Flavored Markdown.`,
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
		exts = []string{".tf", ".tofu"}
		files := checkFilesByExtension(workingDir, exts)
		// we check to see if there are tf or tofu files in the current working directory. If not, we don't call tf.plan
		if files {
			if len(args) == 0 {
				planStr, err = createPlan()
				if err != nil {
					log.Errorf("Unable to create plan: %s", err)
				}
				// Create the Markdown from the Plan.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					log.Errorf("Something is not right, %s", err)
				}

				// Checking to see if plan file was created.
				if _, err := os.Stat(planPath); err == nil {
					fmt.Fprintf(color.Output, "%s%s\n", bold(green("✔")), "  Plan Created...")
					log.Debugf("Plan file %s was created.", planPath)
				} else if errors.Is(err, os.ErrNotExist) {
					// Apparently the binary exists, tf.Plan shit the bed and didn't tell us.
					fmt.Fprintf(color.Output, "%s%s\n", bold(red("✕")), "  Failed to Create Plan...")
					log.Errorf("Plan file %s was not created.", planPath)
				} else {
					// I'm only human. NFC how you got here. I hope to never have to find out.
					log.Errorf("If you see this error message, please open a bug. Error Code: TPE002. Error: %s", err)
				}
			} else if args[0] == "-" {
				out = cmd.InOrStdin()
				content, err := io.ReadAll(out)
				if err != nil {
					log.Errorf("unable to read stdIn: %s", err)
				}

				log.Debugf("plan output: %s", planStr)
				mdParam = viper.GetString("mdFile")

				planStr := string(content)

				log.Debug(planStr)
				// Create the plan from Stdin.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					log.Errorf("Something is not right, %s", err)
				}
				// the arg received looks like a file, we try to open it
			}
		} else {
			log.Errorf("No %s files found. Please run this in a directory with %s files present.",
				cases.Title(language.English).String(binary), cases.Title(language.English).String(binary))
			os.Exit(1)
		}

		// Checking to see if Markdown file was created.
		if _, err := os.Stat(mdParam); err == nil {
			log.Debugf("Markdown file %s was created.", mdParam)
			fmt.Fprintf(color.Output, "%s%s\n", bold(green("✔")), "  Markdown Created...")
		} else if errors.Is(err, os.ErrNotExist) {
			//
			log.Errorf("Markdown file %s was not created.", mdParam)
			fmt.Fprintf(color.Output, "%s%s\n", bold(red("✕")), "  Failed to Create Markdown...")
		} else {
			// I'm only human. NFC how you got here. I hope to never have to find out.
			log.Errorf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
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
		StringVar(&cfgFile,
			"config",
			"",
			"config file (default is $HOME/.tp.toml, can also exist in your project's root directory.)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Version = buildVersion(Version, Commit, Date, BuiltBy)
	rootCmd.SetVersionTemplate(`{{printf "Version %s\n" .Version}}`)
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
	log.Debugf("Using config file: %s", viper.ConfigFileUsed())
}
