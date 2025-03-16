// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	binaries           []string
	out                io.Reader
	logger             *log.Logger
	MaxWidth           int
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
)

type tpFile struct {
	Name    string
	Purpose string
}

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
	Long: heredoc.Doc(`
		'tp' is a GitHub CLI extension to create GitHub pull requests with GitHub Flavored Markdown
		containing the output from an OpenTofu or Terraform plan output, wrapped around
		'<details></details>' element so the plan output is collapsed for easier reading on longer outputs.
		The body of your pull request will look like this https://github.com/esacteksab/gh-tp/example/EXAMPLE-PR.md

		View the README at https://github.com/esacteksab/gh-tp or run 'gh tp init'
		to use a prompt-based form with suggested values to create your .tp.toml config file now.`),

	Run: func(cmd *cobra.Command, args []string) {
		v := viper.IsSet("verbose")
		if v {
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
		logger.Debugf("Defined keys: %s in %s", keys, viper.ConfigFileUsed())

		if doesNotExist(viper.ConfigFileUsed()) {
			logger.Debug(viper.ConfigFileUsed())
			logger.Error("Config file not found. Please run 'gh tp init' or run 'gh tp help' or refer to the documentation on how to create a config file. https://github.com/esacteksab/gh-tp")
			os.Exit(1)
			// May want to put cmd.Help() or something about expectations with config parameters.
		} else {
			// Check to see if required 'planFile' parameter is set
			o := viper.IsSet("planFile")
			if !o {
				logger.Errorf(
					"Missing Parameter: 'planFile' (type: string) is not defined in %s. This is the name of the plan's output file that will be created by `gh tp`.\n", viper.ConfigFileUsed())
				os.Exit(1)
			}

			// Check to see if required 'mdFile' parameter is set
			m := viper.IsSet("mdFile")
			if !m {
				logger.Errorf(
					"Missing Parameter: 'mdFile' (type: string) is not defined in %s. This is the name of the Markdown file that will be created by `gh tp`.\n", viper.ConfigFileUsed())
				os.Exit(1)
			}
			logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
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
				logger.Errorf("Found both `tofu` and `terraform` in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your %s config file to the binary you want to use.", viper.ConfigFileUsed())
				os.Exit(1)
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
			logger.Errorf("No %s files found. Please run this in a directory with %s files present.",
				cases.Title(language.English).String(binary), cases.Title(language.English).String(binary))
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
	logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		TimeFormat:      time.Kitchen,
	})
	MaxWidth = 4
	styles := log.DefaultStyles()
	styles.Levels[log.DebugLevel] = lipgloss.NewStyle().
		SetString(strings.ToUpper(log.DebugLevel.String())).
		Bold(true).MaxWidth(MaxWidth).Foreground(lipgloss.Color("12"))
	logger.SetStyles(styles)
	logger.Debug("Testing new Debug")
	logger.Debugf("I'm inside initConfig() and Verbose is %t:\n", Verbose)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.Flags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().StringP("binary", "b", "", "The name of the binary to use. Expect either `tofu` or `terraform`. Must exist on your $PATH.")
	rootCmd.Flags().StringP("outFile", "o", "", "The name of the plan output file created by tp.")
	rootCmd.Flags().StringP("mdFile", "m", "", "The name of the Markdown file created by tp.")

	err := viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	if err != nil {
		logger.Error("Unable to bind to verbose flag: ", err)
	}
	err = viper.BindPFlag("binary", rootCmd.Flags().Lookup("binary"))
	if err != nil {
		logger.Error("Unable to bind to binary flag: ", err)
	}
	err = viper.BindPFlag("planFile", rootCmd.Flags().Lookup("outFile"))
	if err != nil {
		logger.Error("Unable to bind to planFile flag: ", err)
	}
	err = viper.BindPFlag("mdFile", rootCmd.Flags().Lookup("mdFile"))
	if err != nil {
		logger.Error("Unable to bind to mdFile flag: ", err)
	}

	Verbose, err := rootCmd.Flags().GetBool("verbose")
	logger.Debug("I'm inside init, Verbose is %t\n", Verbose)
	if err != nil {
		logger.Errorf("Unable to get verbose flag: %s", err)
	}

	if Verbose {
		logger.SetLevel(log.DebugLevel)
		logger.Errorf("I'm inside !Verbose init() and my value is %t\n", Verbose)
	}

	rootCmd.Flags().
		StringVarP(&cfgFile,
			"config",
			"c",
			"",
			`use this configuration file (default lookup:
			1. a .tp.toml file in your project's root
			2. $XDG_CONFIG_HOME/.tp.toml
			3. $HOME/.tp.toml)`)

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
		// Find home directory and home config directory.
		homeDir, configDir, _, _ = getDirectories()

		// Search config in home directory with name ".tp.toml"
		viper.SetConfigName(".tp.toml")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(homeDir)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.UnsupportedConfigError); ok {
			logger.Errorf("Unsupported Format. Config file should be named .tp.toml %s.", err)
			os.Exit(1)
			// This handles the situation where a duplicate key exists.
		} else if _, ok := err.(viper.ConfigParseError); ok {
			logger.Errorf("There is an issue %s.", err)
			os.Exit(1)
		}
	}
}
