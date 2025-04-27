// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Verbose bool
	cfgFile string
)

// --- Environment variable for init-phase debugging ---
const ghTpInitDebugEnv = "GH_TP_INIT_DEBUG" // Or your preferred name

func Execute() {
	// Initial Logger -- InfoLevel
	// createLogger(false)
	// --- Check ENV VAR for Initial Verbosity ---
	debugEnvVal := os.Getenv(ghTpInitDebugEnv)
	// Parse bool allows "true", "TRUE", "True", "1"
	initialVerbose, _ := strconv.ParseBool(debugEnvVal)
	// If parsing fails (e.g., empty string), initialVerbose remains false

	// --- Create INITIAL logger based on ENV VAR ---
	createLogger(initialVerbose) // Initia/Configlize with level based on debug env var
	// This log will NOW appear if GH_TP_INIT_DEBUG=true
	Logger.Debugf(
		"Initial logger created in Execute(). Initial Verbose based on %s: %t",
		ghTpInitDebugEnv,
		initialVerbose,
	)

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.Flags().
		StringP("binary", "b", "", "expect either 'tofu' or 'terraform'. Must exist on your $PATH.")
	rootCmd.Flags().
		StringP("planFile", "o", "", "the name of the plan output file to be created by tp (e.g., plan.out).")
	rootCmd.Flags().
		StringP("mdFile", "m", "", "the name of the Markdown file to be created by tp (e.g., plan.md).")
	rootCmd.Flags().
		StringVarP(
			&cfgFile,
			"config",
			"c",
			"",
			`config file to use not in (default lookup:
			1. a .tp.toml file in your project's root
			2. $XDG_CONFIG_HOME/gh-tp/.tp.toml
			3. $HOME/.tp.toml)`,
		)

	// Local var for binding errors
	var bindErr error

	bindErr = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	if bindErr != nil {
		Logger.Fatalf("Internal error binding verbose flag: %v", bindErr)
	}
	bindErr = viper.BindPFlag("binary", rootCmd.Flags().Lookup("binary"))
	if bindErr != nil {
		Logger.Fatalf("Internal error binding binary flag: %v", bindErr)
	}
	bindErr = viper.BindPFlag("planFile", rootCmd.Flags().Lookup("planFile"))
	if bindErr != nil {
		Logger.Fatalf("Internal error binding planFile flag: %v", bindErr)
	}
	bindErr = viper.BindPFlag("mdFile", rootCmd.Flags().Lookup("mdFile"))
	if bindErr != nil {
		Logger.Fatalf("Internal error binding mdFile flag: %v", bindErr)
	}

	Logger.Debug("[EXECUTE_DEBUG] Calling rootCmd.Execute()...")
	executeErr := rootCmd.Execute()
	Logger.Debugf("[EXECUTE_DEBUG] rootCmd.Execute() returned. Error: %v", executeErr)

	// --- Defensive Check: Ensure Logger was created ---
	if Logger == nil {
		// This should ideally never happen if initConfig runs correctly
		fmt.Fprintln(os.Stderr, "[EXECUTE_DEBUG] FATAL: Logger is nil after Execute()!")
		// Create a fallback logger just to report the final state
		if executeErr != nil {
			Logger.Errorf("Command failed with error (logger was nil initially): %v", executeErr)
			os.Exit(1)
		} else {
			Logger.Debug("Command finished (logger was nil initially).")
		}
	}
	// --- End Defensive Check ---

	if executeErr != nil {
		Logger.Debugf(
			"[LOG 13] Exiting(1) because rootCmd.Execute() returned error: %v",
			executeErr,
		)
		os.Exit(1)
	}
	Logger.Debug("[LOG 14] rootCmd.Execute() completed without error.")
}

// init function defines flags and sets up version
func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.Version = buildVersion(Version, Commit, Date, BuiltBy)
	rootCmd.SetVersionTemplate(`{{printf "Version %s" .Version}}`)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	Logger.Debug("[INITCONFIG_DEBUG] Entering initConfig()...")

	configFile := ConfigFile{}

	// --- Viper config setup ---
	if cfgFile != "" {
		// Path 1: Config file specified via -c/--config flag
		viper.SetConfigFile(cfgFile)
		cfgFile = configFile.Path
		Logger.Debugf(
			"[INITCONFIG_DEBUG] Using explicit config file from flag: %s",
			cfgFile,
		)
		err := viper.ReadInConfig()
		Logger.Debugf(
			"[INITCONFIG_DEBUG] ReadInConfig (explicit file) returned error: %v",
			err,
		)
		if err != nil {
			if os.IsNotExist(err) {
				Logger.Error(
					"Config file specified via --config not found.")
				os.Exit(1)
			} else {
				Logger.Debugf("ERROR: Error reading specified config file %s: %v", cfgFile, err)
			}
		} else {
			Logger.Debugf("[INITCONFIG_DEBUG] Successfully read config file: %s", viper.ConfigFileUsed())
		}
	} else {
		// Path 2: No -c/--config flag, search default locations
		Logger.Debug("[INITCONFIG_DEBUG] Searching default locations for .tp.toml...")
		homeDir, configDir, _, dirErr := getDirectories()
		if dirErr != nil {
			Logger.Debugf("ERROR: Cannot determine home/config directories: %v. Relying on flags/env.", dirErr)
			// Is there a better way to handle this scenario? We would typically want to os.Exit(1) as these values are necessary
			// But this breaks `gh tp init`
		} else {
			// Search config in os.UserConfigDir/gh-tp with name ".tp.toml"
			// Search config in os.UserHomeDir with name ".tp.toml"
			// Current Working Directory '.' - Presumed project's root
			viper.SetConfigName(".tp.toml")
			viper.SetConfigType("toml")
			viper.AddConfigPath(".")
			viper.AddConfigPath(filepath.Join(configDir, TpDir))
			viper.AddConfigPath(homeDir)
			Logger.Debugf("[INITCONFIG_DEBUG] Viper search paths: ., %s, %s", filepath.Join(configDir, TpDir), homeDir)

			if err := viper.ReadInConfig(); err != nil {
				Logger.Debugf("[INITCONFIG_DEBUG] ReadInConfig (default search) returned error: %v", err)
				var unsupportedConfigError viper.UnsupportedConfigError
				if !errors.As(err, &unsupportedConfigError) {
					var configParseError viper.ConfigParseError
					if errors.As(err, &configParseError) {
						fmt.Fprintf(os.Stderr, "Error: %s\n", err)
						os.Exit(1) // There is something wrong with the config file, exit
					}
				} else if errors.As(err, &viper.ConfigFileNotFoundError{}) {
					// This is OK
					Logger.Debug("[INITCONFIG_DEBUG] No config file (.tp.toml) found in default locations.")
				} else {
					// Other error (permissions, parsing error in a found file)
					Logger.Debugf("ERROR: Error reading potential config file: %v", err)
					os.Exit(1)
				}
			} else {
				Logger.Debugf("[INITCONFIG_DEBUG] Successfully read config file: %s", viper.ConfigFileUsed())
			}
		}
	}
	// Set AutomaticEnv AFTER attempting to read config
	viper.AutomaticEnv()

	// --- Determine final verbosity from Viper ---
	v := viper.IsSet("verbose")
	if v {
		finalVerboseValue := viper.GetBool("verbose")
		createLogger(finalVerboseValue) // <<< Logger is CREATED HERE
		Verbose = finalVerboseValue
	}

	if Verbose {
		Logger.Debugf("Logger setup complete. Verbose: %t, Level: %s", Verbose, Logger.GetLevel())
		Logger.Debug("Exiting initConfig() function.")
	}
}
