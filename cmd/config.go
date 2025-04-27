// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml/v2"
)

const TpDir = "gh-tp"

const ConfigName = ".tp.toml"

// Global variables used throughout the configuration management system
var (
	accessible         bool                             // Flag to enable accessibility mode for UI interactions
	localNow           string                           // Timestamp string used for backup file naming
	title              string                           // Title for user prompt UI
	defaultFileChecker FileChecker = &RealFileChecker{} // Default implementation of FileChecker interface
	defaultUserPrompt  UserPrompt  = &RealUserPrompt{}  // Default implementation of UserPrompt interface
)

// ConfigFile represents the configuration file structure with its location and parameters
type ConfigFile struct {
	Name   string       // Name of the configuration file
	Path   string       // Full path to the configuration file
	Params ConfigParams // Configuration parameters stored in the file
}

// ConfigParams contains all configurable parameters for the application
// with validation rules and comments for documentation
type ConfigParams struct {
	Binary   string `toml:"binary"   comment:"binary: (type: string) The name of the binary, expect either 'tofu' or 'terraform'. Must exist on your $PATH." validate:"oneof=terraform tofu"`
	PlanFile string `toml:"planFile" comment:"planFile: (type: string) The name of the plan file created by 'gh tp'."                                        validate:"required"`
	MdFile   string `toml:"mdFile"   comment:"mdFile: (type: string) The name of the Markdown file created by 'gh tp'."                                      validate:"required,nefield=PlanFile"`
	Verbose  bool   `toml:"verbose"  comment:"verbose: (type: bool) Enable Verbose Logging. Default is false."                                               validate:"boolean"`
}

// genConfig marshals the configuration parameters into TOML format
//
// This function converts the ConfigParams struct to a TOML byte array
// that can be written to a configuration file.
//
// Parameters:
//
//	conf - The configuration parameters to marshal
//
// Returns:
//
//	data - Byte array containing the marshalled TOML data
//	err - Any error encountered during marshalling, or nil on success
func genConfig(conf ConfigParams) (data []byte, err error) {
	data, err = toml.Marshal(conf)
	if err != nil {
		Logger.Fatalf("Failed marshalling TOML: %s", err)
		return nil, err
	}
	return data, err
}

// FileChecker is an interface for checking file existence
// This allows for dependency injection and easier testing
type FileChecker interface {
	DoesExist(cfgFile string) bool
}

// UserPrompt is an interface for handling user interactions
// This allows for dependency injection and easier testing
type UserPrompt interface {
	AskOverwrite(configExists bool) (createFile bool, err error)
}

// RealFileChecker implements the FileChecker interface for production use
type RealFileChecker struct{}

// DoesExist checks if a file or directory exists at the specified path
//
// This method uses the DoesExist function to determine if a path exists
// in the filesystem.
//
// Parameters:
//
//	cfgFile - The file path to check for existence
//
// Returns:
//
//	bool - true if the path exists, false otherwise
func (r *RealFileChecker) DoesExist(cfgFile string) bool {
	return doesExist(cfgFile)
}

// RealUserPrompt implements the UserPrompt interface for production use
type RealUserPrompt struct{}

// AskOverwrite prompts the user about creating or overwriting a configuration file
//
// This method uses the query function to ask the user whether to create a new
// configuration file or overwrite an existing one.
//
// Parameters:
//
//	configExists - Whether the configuration file already exists
//
// Returns:
//
//	bool - User's decision (true to create/overwrite, false otherwise)
//	error - Any error encountered during user interaction
func (r *RealUserPrompt) AskOverwrite(configExists bool) (bool, error) {
	// Delegates to the query function for actual prompting logic
	return query(configExists)
}

// createOrOverwrite determines if a config file exists and asks the user
// whether to create or overwrite it
//
// This function uses dependency injection for easier testing by accepting
// FileChecker and UserPrompt interfaces.
//
// Parameters:
//
//	cfgFile - The path to the configuration file
//	fileChecker - Implementation of FileChecker to check file existence
//	userPrompt - Implementation of UserPrompt to handle user interaction
//
// Returns:
//
//	configExists - Whether the configuration file already exists
//	createFile - User's decision (true to create/overwrite, false otherwise)
//	err - Any error encountered during the process
func createOrOverwrite(
	cfgFile string,
	fileChecker FileChecker,
	userPrompt UserPrompt,
) (configExists, createFile bool, err error) {
	configExists = fileChecker.DoesExist(cfgFile)
	Logger.Debugf("Using config: %s", cfgFile+ConfigName)
	createFile, err = userPrompt.AskOverwrite(configExists)
	if err != nil {
		Logger.Error(err)
		return false, false, err
	}
	return configExists, createFile, err
}

// FormRunner is an interface for running UI forms
// This allows for dependency injection and easier testing of UI components
type FormRunner interface {
	Run() error
}

// HuhFormRunner implements the FormRunner interface using the huh library
type HuhFormRunner struct {
	title      string // Title displayed in the form
	createFile *bool  // Pointer to store user's selection
	accessible bool   // Whether to enable accessibility features
}

// Run displays a confirmation form to the user and captures their response
//
// This method creates and runs a huh form with a confirmation prompt
// that allows the user to choose yes/no.
//
// Returns:
//
//	error - Any error encountered while running the form
func (h *HuhFormRunner) Run() error {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(h.title).
				Affirmative("Yes").
				Negative("No").
				Value(h.createFile),
		),
	).WithTheme(huh.ThemeBase16()).
		WithAccessible(h.accessible)
	return form.Run()
}

// Factory function for creating FormRunner instances
// Makes it easier to mock during testing
var formRunnerFactory = func(title string, createFile *bool, accessible bool) FormRunner {
	return &HuhFormRunner{
		title:      title,
		createFile: createFile,
		accessible: accessible,
	}
}

// query prompts the user whether to create or overwrite a configuration file
//
// This function checks if accessibility mode is enabled and displays an
// appropriate confirmation prompt to the user based on whether the
// configuration file already exists.
//
// Parameters:
//
//	configExists - Whether the configuration file already exists
//
// Returns:
//
//	createFile - User's decision (true to create/overwrite, false otherwise)
//	err - Any error encountered during user interaction
func query(configExists bool) (createFile bool, err error) {
	// Check if we should run in accessible mode by reading environment variable
	accessible, _ = strconv.ParseBool(os.Getenv("ACCESSIBLE"))

	// Set appropriate title based on whether config exists
	title = "Create new file?"
	if configExists {
		title = "Overwrite existing config file?"
	}

	// Create and run the form
	formRunner := formRunnerFactory(title, &createFile, accessible)
	err = formRunner.Run()
	if err != nil {
		Logger.Error(err)
	}

	return createFile, err
}

// createConfig creates or updates a configuration file with the provided parameters
//
// This function handles the entire configuration creation process:
// 1. Checking if config file exists and asking for user confirmation
// 2. Validating configuration parameters
// 3. Generating TOML configuration
// 4. Creating necessary directories
// 5. Creating backup of existing config if needed
// 6. Writing the new configuration file
//
// Parameters:
//
//	cfgBinary - The binary to use (terraform or tofu)
//	cfgFile - The path to the configuration file
//	cfgMdFile - The name of the markdown file
//	cfgPlanFile - The name of the plan file
//
// Returns:
//
//	error - Any error encountered during the configuration process
func createConfig(cfgBinary, cfgFile, cfgMdFile, cfgPlanFile string) error {
	// Check if config exists and ask user if they want to create/overwrite
	configExists, createFile, err := createOrOverwrite(
		cfgFile,
		defaultFileChecker,
		defaultUserPrompt,
	)
	if err != nil {
		Logger.Error(err)
		return err
	}

	// Set up config file structure
	configFile := ConfigFile{}
	configFile.Path = cfgFile
	configDir = filepath.Dir(cfgFile)

	// Create configuration with provided parameters
	conf := ConfigParams{
		Binary:   cfgBinary,
		PlanFile: cfgPlanFile,
		MdFile:   cfgMdFile,
		Verbose:  false, // Default to non-verbose mode
	}

	err = validateConfig(conf)
	if err != nil {
		Logger.Error(err)
	}

	Logger.Debug("Config is valid")

	// Generate TOML configuration
	config, err := genConfig(conf)
	if err != nil {
		Logger.Error(err)
		return err
	}

	// If user said 'Yes' in AskOverWrite()
	if createFile {
		// Create config directory if $XDG_CONFIG_HOME is chosen and `gh-tp` doesn't exist
		if !doesExist(configDir) {
			if err = os.MkdirAll(
				configDir, 0o750, //nolint:mnd
			); err != nil {
				Logger.Fatal(err)
				return err
			}
		}

		if !configExists {
			// Create new config file if it doesn't exist
			Logger.Debugf(
				"Inside configExists and 'config' is: %s", string(config),
			)
			err = os.WriteFile(
				configFile.Path, config, 0o600, //nolint:mnd
			)
			if err != nil {
				Logger.Fatalf("Error writing Config file: %s", err)
				return err
			}
		} else if configExists {
			// When overwriting existing config, create backup first
			Logger.Debugf("Config is: \n%s\n", string(config))

			// Create timestamp for backup file name
			// #117 This could be moved to BackupFile() I think
			localNow = time.Now().Local().Format("200601021504")
			existingConfigFile := configFile.Path
			bkupConfigFile := configFile.Path + "-" + localNow

			// Create backup of existing config
			err := BackupFile(existingConfigFile, bkupConfigFile)
			if err != nil {
				Logger.Fatal(err)
				return err
			}
			// This could prossibly go in #117
			Logger.Infof("Backup file %s created", bkupConfigFile)

			// Write new config file
			err = os.WriteFile(
				configFile.Path, config, 0o600, //nolint:mnd
			)
			if err != nil {
				Logger.Errorf("Error writing Config file: %s", err)
				return err
			}
		}
		Logger.Infof("Config file %s created", configFile.Path)
	} else if !createFile {
		// If user chose not to create file, just display the config
		Logger.Info(string(config))
	}
	return err
}

func validateConfig(conf ConfigParams) error {
	// Initialize validator with required struct validation
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Register custom tag name function to use field names in validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Name
	})

	// Validate the configuration against defined validation rules
	err := validate.Struct(conf)
	if err != nil {
		var validationErrors []string
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("Field: %s, Error: %s, Param: %s",
					err.Field(), err.Tag(), err.Param()),
			)
		}
		return fmt.Errorf("validation failed: %s", strings.Join(validationErrors, "; "))
	}

	return nil
}
