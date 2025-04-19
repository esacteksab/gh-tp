// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
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
	configDir       string
	cfgFile         string
	out             *bufio.Reader
	mdParam         string // Keep if needed globally, otherwise make local
	spinnerDuration = 100 * time.Millisecond
	Verbose         bool
	Version         string
	Date            string
	Commit          string
	BuiltBy         string
	Logger          *log.Logger
	bold            = color.New(color.Bold).SprintFunc()
	green           = color.New(color.FgGreen).SprintFunc()
	red             = color.New(color.FgRed).SprintFunc()
	binary          string // Keep if set before RunE, otherwise determine locally
	planStr         string // Keep if needed globally, otherwise make local
)

const TpDir = "gh-tp"

const ConfigName = ".tp.toml"

// --- Environment variable for init-phase debugging ---
const ghTpInitDebugEnv = "GH_TP_INIT_DEBUG" // Or your preferred name

// A struct representing the files created by tp
type tpFile struct {
	Name    string
	Purpose string
}

// buildVersion function (no changes)
func buildVersion(Version, Commit, Date, BuiltBy string) string {
	result := Version
	if Commit != "" {
		result = fmt.Sprintf("%s\nCommit: %s\n", result, Commit)
	}
	if Date != "" {
		result = fmt.Sprintf("%sBuilt at: %s\n", result, Date)
	}
	if BuiltBy != "" {
		result = fmt.Sprintf("%sBuilt by: %s\n", result, BuiltBy)
	}
	result = fmt.Sprintf(
		"%sGOOS: %s\nGOARCH: %s\n", result, runtime.GOOS, runtime.GOARCH,
	)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf(
			"%smodule Version: %s, checksum: %s",
			result,
			info.Main.Version,
			info.Main.Sum,
		)
	}
	return result
}

var rootCmd = &cobra.Command{
	Use:          "tp",
	SilenceUsage: true,
	Short:        "A GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output.",
	Long: heredoc.Doc(`
	'tp' is a GitHub CLI extension to create GitHub pull requests with
	GitHub Flavored Markdown containing the output from an OpenTofu or
	Terraform plan output, wrapped around '<details></details>' element so
	the plan output is collapsed for easier reading on longer outputs. The
	body of your pull request will look like this
	https://github.com/esacteksab/gh-tp/example/EXAMPLE-PR.md

	View the README at https://github.com/esacteksab/gh-tp or run
	'gh tp init' to create your .tp.toml config file now.
	`),

	RunE: func(cmd *cobra.Command, args []string) error { // Returns local 'error' type
		Logger.Debug("===> Entered RunE")
		// We will still use 'err' for intermediate results, but the final return
		// value determination will be explicit.
		// --- Declare local err variable FOR THIS FUNCTION'S RETURN ---
		var err error
		var planFileRaw string
		var mdFileRaw string
		var planFileValidated string
		var mdFileValidated string

		// v := viper.IsSet("verbose")
		// if v {
		// Verbose = viper.GetBool("verbose")
		// Consider if logger needs update here if verbosity changes post-init
		// }

		keys := viper.AllKeys()
		Logger.Debugf(
			"Defined keys: %s in %s", keys, viper.ConfigFileUsed(),
		)

		// Check config existence
		fmt.Printf("viper.ConfigFileUsed(): %s", viper.ConfigFileUsed())
		configExists := doesExist(viper.ConfigFileUsed())
		fmt.Printf("configExists: %v\n", configExists)
		if !configExists {
			Logger.Debug(viper.ConfigFileUsed())
			return errors.New(
				"config file not found. Please run 'gh tp init' or refer to the documentation")
		}

		// Check required parameters
		o := viper.IsSet("planFile")
		if !o {
			return fmt.Errorf(
				"missing Parameter: 'planFile' is not defined in %s",
				viper.ConfigFileUsed(),
			)
		}
		m := viper.IsSet("mdFile")
		if !m {
			return fmt.Errorf(
				"missing Parameter: 'mdFile' is not defined in %s",
				viper.ConfigFileUsed(),
			)
		}
		Logger.Debugf("Using config file: %s", viper.ConfigFileUsed())

		// Determine binary (local scope for `exists`)
		b := viper.IsSet("binary")
		if b {
			binary = viper.GetString("binary")
		} else {
			binaries := []string{"tofu", "terraform"}
			var exists []string // Local scope
			for _, binName := range binaries {
				binPath, lookupErr := safeexec.LookPath(binName) // Use different var name
				if lookupErr == nil && len(binPath) > 0 {
					exists = append(exists, binName)
					binary = binName
				} else {
					Logger.Debugf("Did not find '%s' in PATH: %v", binName, lookupErr)
				}
			}
			if len(exists) == 0 {
				return errors.New("could not find 'tofu' or 'terraform' in your PATH")
			}
			if len(exists) > 1 {
				return fmt.Errorf(
					"found both %s in your PATH. Set the 'binary' parameter in %s",
					strings.Join(exists, " and "), viper.ConfigFileUsed(),
				)
			}
		}
		Logger.Debugf("Using binary: %s", binary) // Log determined binary

		// Validate Files (use local err)
		planFileRaw = viper.GetString("planFile")
		planFileValidated, err = validateFilePath(planFileRaw) // Assign to local err
		if err != nil {
			Logger.Debugf("planFileRaw validation failed: %s", planFileRaw)
			return fmt.Errorf("invalid 'planFile' configuration (%q): %w", planFileRaw, err)
		}

		mdFileRaw = viper.GetString("mdFile")
		mdFileValidated, err = validateFilePath(mdFileRaw) // Assign to local err
		if err != nil {
			Logger.Debugf("mdFileRaw validation failed: %s", mdFileRaw)
			// --- FIX: Correct variable name in error message ---
			return fmt.Errorf("invalid 'mdFile' configuration (%q): %w", mdFileRaw, err)
		}

		// Check for existence of .tf or .tofu files
		fileExts := []string{".tf", ".tofu"}
		files := checkFilesByExtension(".", fileExts)
		if !files {
			titleCaser := cases.Title(language.English)
			return fmt.Errorf(
				"no %s files found in current directory. Please run this in a directory with %s files",
				titleCaser.String(binary),
				titleCaser.String(binary),
			)
		}

		// --- Execution Logic ---
		Logger.Debug("[LOG 1] Starting RunE execution...")

		if len(args) == 0 {
			// --- Assign to LOCAL err ---
			planStr, err = createPlan() // Uses local err

			Logger.Debugf("[LOG 2] createPlan returned. err: %v (type: %T)", err, err)

			// --- Check LOCAL err ---
			if err != nil {
				Logger.Debug("[LOG 3] Entered RunE error handling block.")
				if errors.Is(err, ErrInterrupted) {
					Logger.Debug("[LOG 4] Detected ErrInterrupted.")
					Logger.Debug("Operation cancelled.")

					planPathForCleanup := planFileValidated // Use local validated name
					Logger.Debugf("[LOG 5b] Attempting final cleanup of %q...", planPathForCleanup)
					removeErr := os.Remove(planPathForCleanup)
					if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
						Logger.Warnf(
							"[LOG 5c] Cleanup failed for %q: %v",
							planPathForCleanup,
							removeErr,
						)
					} else if removeErr == nil {
						Logger.Debugf("[LOG 5d] Cleanup success for %q.", planPathForCleanup)
					}

					Logger.Debug("[LOG 6] Flushing stderr...")
					_ = os.Stderr.Sync()
					Logger.Debug("[LOG 7] Calling os.Exit(1) due to interruption.")
					os.Exit(1) // Exit directly
				} else { // Other errors
					Logger.Debugf("[LOG 8] Error was not ErrInterrupted: %v.", err)
					Logger.Debugf("Error: Plan creation failed: %s", err)
					return err // Return non-interrupt error
				}
			}

			Logger.Debug("[LOG 9] createPlan returned nil error. Proceeding.")
			// --- Success Path ---
			Logger.Debug("Plan created successfully. Generating Markdown...")
			// Use local mdErr, assign to local mdParam
			var mdErr error
			mdParam, mdErr = createMarkdown(mdFileValidated, planStr, binary)
			if mdErr != nil {
				Logger.Debugf("Error: Markdown creation failed: %s", mdErr)
				return fmt.Errorf("markdown creation failed for '%s': %w", mdFileValidated, mdErr)
			}
			Logger.Debugf("Markdown file '%s' created successfully.", mdParam)

			// --- Handle stdin ---
		} else if args[0] == "-" {
			s := spinner.New(spinner.CharSets[14], spinnerDuration)
			s.Suffix = " Reading plan from stdin and creating Markdown..."
			s.Start()
			defer s.Stop()

			Logger.Debugf("Reading plan from stdin...")
			out = bufio.NewReader(cmd.InOrStdin())
			fi, statErr := os.Stdin.Stat() // Use local statErr
			if statErr != nil {
				err = fmt.Errorf("failed to stat stdin: %w", statErr) // Assign to local err
				Logger.Debugf("Error: %s", err)
				return err
			}
			if fi.Size() == 0 && fi.Mode()&os.ModeNamedPipe == 0 {
				err = errors.New("no input provided via stdin pipe") // Assign to local err
				Logger.Debugf("Error: %s", err)
				return err
			}
			content, readErr := io.ReadAll(out) // Use local readErr
			if readErr != nil {
				err = fmt.Errorf("failed to read from stdin: %w", readErr) // Assign to local err
				Logger.Debugf("Error: %s", err)
				return err
			}
			planStr = string(content) // Assign to local planStr for this scope
			if planStr == "" {
				err = errors.New("received empty plan from stdin") // Assign to local err
				Logger.Debugf("Error: %s", err)
				return err
			}

			// Use mdFileValidated name determined earlier
			currentMdParam := mdFileValidated

			Logger.Debugf("Read %d bytes from stdin. Creating Markdown file '%s'...", len(planStr), currentMdParam)
			// Use local mdErr
			var mdErr error
			mdParam, mdErr = createMarkdown(currentMdParam, planStr, binary)
			if mdErr != nil {
				err = fmt.Errorf("markdown creation failed for '%s': %w", currentMdParam, mdErr) // Assign to local err
				Logger.Debugf("Error: %s", err)
				return err
			}
			Logger.Infof("Markdown file '%s' created successfully from stdin.", mdParam) // Use Info for user success

		} else { // Handle unexpected arguments
			err = fmt.Errorf("unexpected argument: %s. Use '-' to read from stdin or no arguments to run plan", args[0])
			Logger.Debugf("Error: %s", err)
			return err
		}

		// --- Final Check ---
		Logger.Debug("[LOG 10] Reached final check.")
		tpFiles := []tpFile{{planFileValidated, "Plan"}, {mdParam, "Markdown"}}
		if len(args) > 0 && args[0] == "-" {
			tpFiles = []tpFile{{mdParam, "Markdown"}}
		}
		err = existsOrCreated(tpFiles) // Assign to local err
		if err != nil {
			Logger.Debugf("Error: File verification failed: %s", err)
			return fmt.Errorf("output file verification failed: %w", err)
		}

		Logger.Debug("✔ Processing complete.")
		Logger.Debug("[LOG 11] RunE finished successfully.")
		return nil // Success!
	},
}

func Execute() {
	// --- Check ENV VAR for Initial Verbosity ---
	debugEnvVal := os.Getenv(ghTpInitDebugEnv)
	// Parse bool allows "true", "TRUE", "True", "1"
	initialVerbose, _ := strconv.ParseBool(debugEnvVal)
	// If parsing fails (e.g., empty string), initialVerbose remains false

	// --- Create INITIAL logger based on ENV VAR ---
	createLogger(initialVerbose) // Initialize with level based on debug env var
	// This log will NOW appear if GH_TP_INIT_DEBUG=true
	Logger.Debugf(
		"Initial logger created in Execute(). Initial Verbose based on %s: %t",
		ghTpInitDebugEnv,
		initialVerbose,
	)

	// Set Silence flags
	// rootCmd.SilenceUsage = true
	// rootCmd.SilenceErrors = true

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
		// os.Exit(1)
	}
	Logger.Debug("[LOG 14] rootCmd.Execute() completed without error.")
}

// init function defines flags and sets up version
func init() {
	cobra.OnInitialize(initConfig)

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
				// Use fmt because Logger might not exist if exit happens
				Logger.Debugf(
					"ERROR: Config file specified via --config (%s) not found.",
					cfgFile,
				)
				os.Exit(1)
			} else {
				Logger.Debugf("ERROR: Error reading specified config file %s: %v", cfgFile, err)
				os.Exit(1)
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

			err := viper.ReadInConfig()
			Logger.Debugf("[INITCONFIG_DEBUG] ReadInConfig (default search) returned error: %v", err)

			if err != nil {
				if errors.As(err, &viper.ConfigFileNotFoundError{}) {
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
