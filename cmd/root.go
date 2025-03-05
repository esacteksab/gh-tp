// SPDX-License-Identifier: MIT

package cmd

import (
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
	planPath           string
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
	logger             = log.New(os.Stderr)
)

type tpFile struct {
	Name    string
	Purpose string
}

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
	Short:   "A GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output.",
	Long:    `tp is a GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output formatted in GitHub Flavored Markdown.`,
	Run: func(cmd *cobra.Command, args []string) {
		v := viper.IsSet("verbose")
		if v {
			logger.Debug("verbose is defined in .tp.toml")
			Verbose = viper.GetBool("verbose")
			logger.Debug("I'm inside runCmd 'if v' and verbose is: %t\n", Verbose)
		} else {
			logger.Debug("I'm inside runCmd and v is not defined in .tp.toml")
		}

		Verbose, err := cmd.Flags().GetBool("verbose")
		logger.Debug("I'm inside runCmd(), and Verbose is %t\n", Verbose)
		if err != nil {
			logger.Errorf("Unable to get verbose flag: %s", err)
		}
		if Verbose {
			logger.SetLevel(log.DebugLevel)
			logger.Debug("I'm inside runCmd Verbose and my value is %t\n", Verbose)
		}

		b := viper.IsSet("binary")
		if b {
			binary = viper.GetString("binary")
		} else {
			exists := []string{}
			binaries = []string{"tofu", "terraform"}
			for _, v := range binaries {
				bin, err := safeexec.LookPath(v)
				if err != nil {
					logger.Debugf("%s", err)
				}
				// It's possible for both `tofu` and `terraform` to exist on $PATH and we need to handle that.
				if len(bin) > 0 {
					exists = append(exists, bin)
				}
			}
			if len(exists) == len(binaries) {
				logger.Fatal("Seems both `tofu` and `terraform` exist in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your .tp.toml config file to whichever binary you want to use.")
			}
		}

		planPath = viper.GetString("planFile")
		exts = []string{".tf", ".tofu"}
		files := checkFilesByExtension(workingDir, exts)
		// we check to see if there are tf or tofu files in the current working directory. If not, we don't call tf.plan
		if files {
			if len(args) == 0 {
				planStr, err = createPlan()
				if err != nil {
					logger.Errorf("Unable to create plan: %s", err)
				}
				// Create the Markdown from the Plan.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					logger.Errorf("Something is not right, %s", err)
				}

			} else if args[0] == "-" {
				out = cmd.InOrStdin()
				content, err := io.ReadAll(out)
				if err != nil {
					logger.Errorf("unable to read stdIn: %s", err)
				}

				mdParam = viper.GetString("mdFile")

				planStr := string(content)

				logger.Debugf("Plan output is: %s\n", planStr)
				// Create the plan from Stdin.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					logger.Errorf("Something is not right, %s", err)
				}
				// the arg received looks like a file, we try to open it
			}
		} else {
			logger.Errorf("No %s files found. Please run this in a directory with %s files present.", cases.Title(language.English).String(binary), cases.Title(language.English).String(binary))
			os.Exit(1)
		}

		tpFiles := []tpFile{
			{planPath, "Plan"},
			{mdParam, "Markdown"},
		}

		tpFilesErr := existsOrCreated(tpFiles)
		if tpFilesErr != nil {
			logger.Error(err)
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
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "", false, "verbose output")

	err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if err != nil {
		logger.Debug("Unable to bind to verbose flag: ", err)
	}
	viper.RegisterAlias("debug", "verbose")
	Verbose, err := rootCmd.Flags().GetBool("verbose")
	logger.Debug("I'm inside init, Verbose is %t\n", Verbose)
	if err != nil {
		logger.Errorf("Unable to get verbose flag: %s", err)
	}
	if Verbose {
		logger.SetLevel(log.DebugLevel)
		logger.Debug("I'm inside !Verbose init() and my value is %t\n", Verbose)
	}

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
			logger.Error(
				"Missing Config File: Config file should be named .tp.toml and exist in your home directory or in your project's root.\n",
			)
			os.Exit(1)
		} else if _, ok := err.(viper.UnsupportedConfigError); ok {
			logger.Errorf("Unsupported Format. Config file should be named .tp %s.", err)
			os.Exit(1)
			// This handles the situation where a duplicate key exists.
		} else if _, ok := err.(viper.ConfigParseError); ok {
			logger.Errorf("There is an issue %s.", err)
			os.Exit(1)
		}
	}
	// Verbose = viper.GetBool("verbose")
	logger.Debug("I'm inside initConfig() and Verbose is %t:\n", Verbose)
	v := viper.IsSet("verbose")
	if v {
		logger.Debugf("Verbose is %t:\n", v)
		Verbose = viper.GetBool("verbose")
	}
	if err != nil {
		logger.Fatal("Unable to enable verbose output:", err)
	}
	if Verbose {
		logger.SetLevel(log.DebugLevel)
		logger.Debug("I'm a Debug statement in initConfig().")
	}
	keys := viper.AllKeys()
	logger.Debugf("Defined keys in .tp.toml: %s", keys)

	// // Check to see if required 'planFile' parameter is set
	o := viper.IsSet("planFile")
	if !o {
		logger.Error(
			"Missing Parameter: 'planFile' (type: string) is not defined in the config file. This is the name of the plan's output file that will be created by `gh tp`.\n",
		)
		os.Exit(1)
	}

	// // Check to see if required 'mdFile' parameter is set
	m := viper.IsSet("mdFile")
	if !m {
		logger.Error(
			"Missing Parameter: 'mdFile' (type: string) is not defined in the config file. This is the name of the Markdown file that will be created by `gh tp`.\n",
		)
		os.Exit(1)
	}
	logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
}
