// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/cli/safeexec"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	configDir          string
	cfgFile            string
	out                *bufio.Reader
	mdParam            string
	spinnerDuration    time.Duration
	titleCaseConverter cases.Caser
	planPath           string
	Verbose            bool
	Version            string
	Date               string
	Commit             string
	BuiltBy            string
	workingDir         string
	Logger             *log.Logger
	bold               = color.New(color.Bold).SprintFunc()
	green              = color.New(color.FgGreen).SprintFunc()
	red                = color.New(color.FgRed).SprintFunc()
)

const TpDir = "gh-tp"

const ConfigName = ".tp.toml"

// A struct representing the files created by tp
// An Plan (Purpose) file named (Name)
// A  Markdown (Purpose) file named (Name)
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
	result = fmt.Sprintf(
		"%s\nGOOS: %s\nGOARCH: %s", result, runtime.GOOS, runtime.GOARCH,
	)
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

var rootCmd = &cobra.Command{
	Use:   "tp",
	Short: "A GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output.",
	Long: heredoc.Doc(
		`
		'tp' is a GitHub CLI extension to create GitHub pull requests with
		GitHub Flavored Markdown containing the output from an OpenTofu or
		Terraform plan output, wrapped around '<details></details>' element so
		the plan output is collapsed for easier reading on longer outputs. The
		body of your pull request will look like this
		https://github.com/esacteksab/gh-tp/example/EXAMPLE-PR.md

		View the README at https://github.com/esacteksab/gh-tp or run
		'gh tp init' to create your .tp.toml config file now.`,
	),

	Run: func(cmd *cobra.Command, args []string) {
		v := viper.IsSet("verbose")
		if v {
			Verbose = viper.GetBool("verbose")
			createLogger(Verbose)
		}

		keys := viper.AllKeys()
		Logger.Debugf(
			"Defined keys: %s in %s", keys, viper.ConfigFileUsed(),
		)

		configExists := doesExist(viper.ConfigFileUsed())
		if !configExists {
			Logger.Debug(viper.ConfigFileUsed())
			Logger.Error(
				"Config file not found. Please run 'gh tp init' or run 'gh tp help' or refer to the documentation on how to create a config file. https://github.com/esacteksab/gh-tp")
			os.Exit(1)
		} else {
			// Check to see if required 'planFile' parameter is set
			o := viper.IsSet("planFile")
			if !o {
				Logger.Errorf(
					"Missing Parameter: 'planFile' (type: string) is not defined in %s. This is the name of the plan's output file that will be created by `gh tp`.",
					viper.ConfigFileUsed(),
				)
				os.Exit(1)
			}

			// Check to see if required 'mdFile' parameter is set
			m := viper.IsSet("mdFile")
			if !m {
				Logger.Errorf(
					"Missing Parameter: 'mdFile' (type: string) is not defined in %s. This is the name of the Markdown file that will be created by `gh tp`.",
					viper.ConfigFileUsed(),
				)
				os.Exit(1)
			}
			Logger.Debugf("Using config file: %s", viper.ConfigFileUsed())
		}

		b := viper.IsSet("binary")
		if b {
			binary = viper.GetString("binary")
		} else {
			var binaries []string
			var exists []string

			exists = []string{}
			binaries = []string{"tofu", "terraform"}

			for _, v := range binaries {
				bin, err := safeexec.LookPath(v)
				if err != nil {
					Logger.Debugf("%s", err)
				}
				// It's possible for both `tofu` and `terraform` to exist on $PATH, and we need to handle that.
				if len(bin) > 0 {
					exists = append(exists, bin)
				}
			}
			if len(exists) == len(binaries) {
				Logger.Errorf(
					"Found both `tofu` and `terraform` in your $PATH. We're not sure which one to use. Please set the 'binary' parameter in your %s config file to the binary you want to use.",
					viper.ConfigFileUsed(),
				)
				os.Exit(1)
			}
		}

		planPath = viper.GetString("planFile")

		fileExts := []string{".tf", ".tofu"}
		files := checkFilesByExtension(workingDir, fileExts)
		// we check to see if there are tf or tofu files in the current working
		// directory. If not, there is no sense in tf.plan
		if files {
			if len(args) == 0 {
				Logger.Debugf("args: %s", args)
				planStr, err = createPlan()
				if err != nil {
					Logger.Errorf("Unable to create plan: %s", err)
				}
				// Create the Markdown from the Plan.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					Logger.Errorf("Something is not right, %s", err)
				}

			} else if args[0] == "-" {
				spinnerDuration = 100
				s := spinner.New(
					spinner.CharSets[14], spinnerDuration*time.Millisecond,
				)
				s.Suffix = "  Creating the Plan...\n"
				s.Start()

				Logger.Debugf("args: %s", args)

				out = bufio.NewReader(cmd.InOrStdin())
				// os.Stdin is *os.File, checking the size to see if it holds any data
				fi, err := os.Stdin.Stat()
				if err != nil {
					Logger.Error(err)
				}

				// stdin is a file size of 0, so we check if the os.ModeNamedPipe
				// is set in *os.File's Mode()
				// https://cs.opensource.google/go/go/+/refs/tags/go1.24.1:src/os/types.go;l=46
				if fi.Size() == 0 && fi.Mode()&os.ModeNamedPipe == 0 {
					Logger.Error("No input provided via stdin")
					os.Exit(1)
				}

				content, err := io.ReadAll(out)
				if err != nil {
					Logger.Errorf("unable to read stdin: %s", err)
				}

				mdParam = viper.GetString("mdFile")

				planStr := string(content)

				Logger.Debugf("Plan output is: %s\n", planStr)
				// Create the plan from Stdin.
				planMd, mdParam, err = createMarkdown(mdParam, planStr)
				if err != nil {
					Logger.Errorf("Something is not right, %s", err)
				}
				s.Stop()
			}
		} else {
			Logger.Errorf(
				"No %s files found. Please run this in a directory with %s files present.",
				cases.Title(language.English).String(binary),
				cases.Title(language.English).String(binary),
			)
			os.Exit(1)
		}

		tpFiles := []tpFile{
			{planPath, "Plan"},
			{mdParam, "Markdown"},
		}

		tpFilesErr := existsOrCreated(tpFiles)
		if tpFilesErr != nil {
			Logger.Error(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	createLogger(Verbose)
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
	rootCmd.PersistentFlags().BoolVarP(
		&Verbose, "verbose", "v", false, "Verbose output",
	)
	rootCmd.Flags().StringP(
		"binary", "b", "",
		"Expect either 'tofu' or 'terraform'. Must exist on your $PATH.",
	)
	rootCmd.Flags().StringP(
		"outFile", "o", "",
		"The name of the plan output file to be created by tp.",
	)
	rootCmd.Flags().StringP(
		"mdFile", "m", "", "The name of the Markdown file to be created by tp.",
	)

	err := viper.BindPFlag(
		"verbose", rootCmd.PersistentFlags().Lookup("verbose"),
	)
	if err != nil {
		Logger.Error("Unable to bind to verbose flag: ", err)
	}
	err = viper.BindPFlag("binary", rootCmd.Flags().Lookup("binary"))
	if err != nil {
		Logger.Error("Unable to bind to binary flag: ", err)
	}
	err = viper.BindPFlag("planFile", rootCmd.Flags().Lookup("outFile"))
	if err != nil {
		Logger.Error("Unable to bind to planFile flag: ", err)
	}
	err = viper.BindPFlag("mdFile", rootCmd.Flags().Lookup("mdFile"))
	if err != nil {
		Logger.Error("Unable to bind to mdFile flag: ", err)
	}

	rootCmd.Flags().
		StringVarP(
			&cfgFile,
			"config",
			"c",
			"",
			`Config file to use (default lookup:
		1. a .tp.toml file in your project's root
		2. $XDG_CONFIG_HOME/gh-tp/.tp.toml
		3. $HOME/.tp.toml)`,
		)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Version = buildVersion(Version, Commit, Date, BuiltBy)
	rootCmd.SetVersionTemplate(`{{printf "Version %s\n" .Version}}`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := ConfigFile{}
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		// Set the Path in ConfigFile struct
		cfgFile = configFile.Path
	} else {
		// Find home directory and home config directory.
		homeDir, configDir, _, _ := getDirectories()

		// Search config in os.UserConfigDir/gh-tp with name ".tp.toml"
		// Search config in os.UserHomeDir with name ".tp.toml"
		// Current Working Directory '.' - Presumed project's root
		viper.SetConfigName(".tp.toml")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
		// $XDG_CONFIG_HOME/gh-tp
		viper.AddConfigPath(configDir + "/" + TpDir)
		// os.UserHomeDir
		viper.AddConfigPath(homeDir)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		var unsupportedConfigError viper.UnsupportedConfigError
		if !errors.As(err, &unsupportedConfigError) {
			var configParseError viper.ConfigParseError
			if errors.As(err, &configParseError) {
				Logger.Error(err)
				os.Exit(1)
			}
		}
	}
}
