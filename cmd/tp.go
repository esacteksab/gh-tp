// SPDX-License-Identifier: MIT

package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/briandowns/spinner"
	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	configDir       string
	out             *bufio.Reader
	mdParam         string // Keep if needed globally, otherwise make local
	spinnerDuration = 100 * time.Millisecond
	Version         string
	Date            string
	Commit          string
	BuiltBy         string
	Logger          *log.Logger
	bold            = color.New(color.Bold).SprintFunc()
	green           = color.New(color.FgGreen).SprintFunc()
	red             = color.New(color.FgRed).SprintFunc()
	binary          string // Deterined binary (terraform or tofu)
	planStr         string // Contents of the plan output
)

// A struct representing the files created by tp
type tpFile struct {
	Name    string
	Purpose string
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:          "tp [-o <planfile>] [-m <mdfile>] [-b <binary>] [-t <templateFile>] | tp -",
	SilenceUsage: true,
	Short:        "A GitHub CLI extension to submit a pull request with Terraform or OpenTofu plan output.",
	Long: heredoc.Doc(`
	'tp' is a GitHub CLI extension to create GitHub pull requests with
	GitHub Flavored Markdown containing the output from an OpenTofu or
	Terraform plan output, wrapped around '<details></details>' element so
	the plan output is collapsed for easier reading on longer outputs. The
	body of your pull request will look like this
	https://github.com/esacteksab/gh-tp/example/EXAMPLE-PR.md

	Flags (-o, -m, -b, -t) can be used instead of a config file if a unique
	binary (terraform or tofu) is found in your PATH. If flags are provided,
	they override any config file settings.

	Use 'tp -' to read plan output directly from stdin.

	View the README at https://github.com/esacteksab/gh-tp or run
	'gh tp init' to create your .tp.toml config file now.
	`),

	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var planFileRaw string
		var mdFileRaw string
		var planFileValidated string
		var mdFileValidated string

		// --- Determine Binary ---
		binary, err = determineBinary()
		if err != nil {
			return err
		}
		Logger.Debugf("Using binary: %s", binary)

		// --- Get Config File Path (if loaded) ---
		loadedConfigFile := viper.ConfigFileUsed() // Get path Viper actually used, if any
		Logger.Debugf("loadedConfigFile in RunE is: %s", loadedConfigFile)

		// --- Determine Plan File Path ---
		if !viper.IsSet("planFile") {
			if loadedConfigFile == "" {
				return fmt.Errorf(
					"required parameter 'planFile' not defined via flag (-o/--planFile) and no loadable config file was found (checked standard locations for '%s', or specified via --config). Use the flag or run 'gh tp init'",
					ConfigName,
				)
			} else {
				return fmt.Errorf(
					"required parameter 'planFile' is not defined via flag (-o/--planFile) or in the loaded config file: %s",
					loadedConfigFile,
				)
			}
		}
		planFileRaw = viper.GetString("planFile")
		planFileValidated, err = validateFilePath(planFileRaw)
		if err != nil {
			Logger.Debugf("planFile validation failed: %s", planFileRaw)
			return fmt.Errorf("invalid 'planFile' configuration/flag (%q): %w", planFileRaw, err)
		}
		Logger.Debugf("Using plan file: %s", planFileValidated)

		// --- Determine Markdown File Path ---
		if !viper.IsSet("mdFile") {
			if loadedConfigFile == "" {
				return fmt.Errorf(
					"required parameter 'mdFile' not defined via flag (-m/--mdFile) and no loadable config file was found (checked standard locations for '%s', or specified via --config). Use the flag or run 'gh tp init'",
					ConfigName,
				)
			} else {
				return fmt.Errorf(
					"required parameter 'mdFile' is not defined via flag (-m/--mdFile) or in the loaded config file: %s",
					loadedConfigFile,
				)
			}
		}
		mdFileRaw = viper.GetString("mdFile")
		mdFileValidated, err = validateFilePath(mdFileRaw)
		if err != nil {
			Logger.Debugf("mdFile validation failed: %s", mdFileRaw)
			return fmt.Errorf("invalid 'mdFile' configuration/flag (%q): %w", mdFileRaw, err)
		}
		Logger.Debugf("Using markdown file: %s", mdFileValidated)

		// --- Logging & File Checks ---
		if loadedConfigFile != "" {
			Logger.Debugf("Effective config file used: %s", loadedConfigFile)
			keys := viper.AllKeys()
			Logger.Debugf("Effective Viper keys (flags > config): %s", keys)
		} else {
			Logger.Debug("No config file loaded; using flags and/or auto-detection for parameters.")
		}

		// Check for existence of .tf or .tofu files (only if not reading from stdin)
		if len(args) == 0 {
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
		}

		// --- Execution Logic ---
		Logger.Debug("[LOG 1] Starting RunE execution...")

		if len(args) == 0 { // Run plan mode
			planStr, err = createPlan()
			Logger.Debugf("[LOG 2] createPlan returned. err: %v (type: %T)", err, err)

			if err != nil {
				Logger.Debug("[LOG 3] Entered RunE error handling block.")
				if errors.Is(err, ErrInterrupted) {
					Logger.Debug("[LOG 4] Detected ErrInterrupted.")
					Logger.Info("Operation cancelled by user.") // Use Info for user feedback

					planPathForCleanup := planFileValidated
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
					// The GitHub CLI often exits with 0 on SIGINT, let's try that first.
					// If issues persist, revert to os.Exit(1) but standard gh extensions often return 0 here.
					Logger.Debug("[LOG 6] Returning nil error after user interrupt cleanup.")
					return nil // Exit gracefully after cancellation
					// os.Exit(1) // Alternative if returning nil doesn't work as expected upstream
				} else { // Other errors from createPlan
					Logger.Debugf("[LOG 8] Error was not ErrInterrupted: %v.", err)
					// Error already logged within createPlan, just return it
					return err
				}
			}

			Logger.Debug("[LOG 9] createPlan returned nil error. Proceeding.")
			// Logger.Info(green("✔ ") + " Plan Created...") // User feedback

			// --- Generate Markdown ---
			Logger.Debugf("Generating Markdown file '%s'...", mdFileValidated)
			var mdErr error
			// Use mdFileValidated for the target path
			mdParam, mdErr = createMarkdown(mdFileValidated, planStr, binary)
			if mdErr != nil {
				Logger.Debugf("Error: Markdown creation failed: %s", mdErr)
				return fmt.Errorf("markdown creation failed for '%s': %w", mdFileValidated, mdErr)
			}
			Logger.Debugf("Markdown file '%s' created successfully.", mdParam)
			// Logger.Info(green("✔ ") + " Markdown Created...") // User feedback

		} else if args[0] == "-" { // Stdin mode
			s := spinner.New(spinner.CharSets[14], spinnerDuration)
			s.Suffix = " Reading plan from stdin and creating Markdown..."
			s.Start()

			Logger.Debugf("Reading plan from stdin...")
			out = bufio.NewReader(cmd.InOrStdin())
			fi, statErr := os.Stdin.Stat()
			if statErr != nil {
				s.Stop() // Stop spinner before returning error
				err = fmt.Errorf("failed to stat stdin: %w", statErr)
				Logger.Debugf("Error: %s", err)
				return err
			}
			// Check if stdin is empty or not a pipe/redirect
			if fi.Size() == 0 && fi.Mode()&os.ModeCharDevice != 0 {
				s.Stop() // Stop spinner before returning error
				err = errors.New("no input provided via stdin pipe or redirect")
				Logger.Debugf("Error: %s", err)
				return err
			}
			content, readErr := io.ReadAll(out)
			if readErr != nil {
				s.Stop() // Stop spinner before returning error
				err = fmt.Errorf("failed to read from stdin: %w", readErr)
				Logger.Debugf("Error: %s", err)
				return err
			}
			s.Stop() // Stop spinner after reading

			planStr = string(content)
			if planStr == "" {
				err = errors.New("received empty plan from stdin")
				Logger.Debugf("Error: %s", err)
				return err
			}

			// Use mdFileValidated determined earlier
			currentMdParam := mdFileValidated
			Logger.Debugf("Read %d bytes from stdin. Creating Markdown file '%s'...", len(planStr), currentMdParam)

			// --- Generate Markdown ---
			var mdErr error
			mdParam, mdErr = createMarkdown(currentMdParam, planStr, binary)
			if mdErr != nil {
				err = fmt.Errorf("markdown creation failed for '%s': %w", currentMdParam, mdErr)
				Logger.Debugf("Error: %s", err)
				return err
			}
			Logger.Debugf("Markdown file '%s' created successfully from stdin.", mdParam)
			Logger.Info(green("✔ ") + " Markdown Created from stdin...") // User feedback

		} else { // Handle unexpected arguments
			err = fmt.Errorf("unexpected argument: %s. Use '-' to read from stdin or no arguments to run plan", args[0])
			Logger.Debugf("Error: %s", err)
			return err
		}

		// --- Final Check (adjusted based on mode) ---
		Logger.Debug("[LOG 10] Reached final check.")
		var filesToCheck []tpFile
		if len(args) == 0 { // Ran plan mode
			filesToCheck = []tpFile{{planFileValidated, "Plan"}, {mdParam, "Markdown"}}
		} else if args[0] == "-" { // Stdin mode
			filesToCheck = []tpFile{{mdParam, "Markdown"}}
		}

		// Perform the check only if there are files expected
		if len(filesToCheck) > 0 {
			err = existsOrCreated(filesToCheck)
			if err != nil {
				Logger.Debugf("Error: File verification failed: %s", err)
				// Provide a more specific error message
				return fmt.Errorf("output file verification failed (%s): %w", err.Error(), err)
			}
		}

		Logger.Debug("✔ Processing complete.")
		Logger.Debug("[LOG 11] RunE finished successfully.")
		return nil // Success!
	},
}
