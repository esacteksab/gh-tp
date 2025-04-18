// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/go-playground/validator/v10"
	"github.com/pelletier/go-toml/v2"
)

var (
	accessible   bool
	binary       string
	configExists bool
	createFile   bool
	localNow     string
	title        string
)

type ConfigFile struct {
	Name   string
	Path   string
	Params ConfigParams
}

type ConfigParams struct {
	Binary   string `toml:"binary"   comment:"binary: (type: string) The name of the binary, expect either 'tofu' or 'terraform'. Must exist on your $PATH." validate:"oneof=terraform tofu"`
	PlanFile string `toml:"planFile" comment:"planFile: (type: string) The name of the plan file created by 'gh tp'."                                        validate:"required"`
	MdFile   string `toml:"mdFile"   comment:"mdFile: (type: string) The name of the Markdown file created by 'gh tp'."                                      validate:"required,nefield=PlanFile"`
	Verbose  bool   `toml:"verbose"  comment:"verbose: (type: bool) Enable Verbose Logging. Default is false."                                               validate:"boolean"`
}

func genConfig(conf ConfigParams) (data []byte, err error) {
	data, err = toml.Marshal(conf)
	if err != nil {
		Logger.Fatalf("Failed marshalling TOML: %s", err)
	}
	return data, err
}

// Checks the existence of a config file.
// If one exists, asks to overwrite it, otherwise creates it.
func createOrOverwrite(cfgFile string) (configExists, createFile bool, err error) {
	configExists = doesExist(cfgFile)
	Logger.Debugf("Using config: %s", cfgFile+ConfigName)
	createFile, err = query(configExists)
	if err != nil {
		Logger.Error(err)
		return false, false, err
	}

	return configExists, createFile, err
}

func query(configExists bool) (createFile bool, err error) {
	// Should we run in accessible mode?
	accessible, _ = strconv.ParseBool(os.Getenv("ACCESSIBLE"))

	title = "Create new file?"
	if configExists {
		title = "Overwrite existing config file?"
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes").
				Negative("No").
				Value(&createFile),
		),
	).WithTheme(huh.ThemeBase16()).
		WithAccessible(accessible)

	err = form.Run()
	if err != nil {
		Logger.Error(err)
	}

	return createFile, err
}

func createConfig(cfgBinary, cfgFile, cfgMdFile, cfgPlanFile string) error {
	configExists, createFile, err = createOrOverwrite(cfgFile)
	if err != nil {
		Logger.Error(err)
		return err
	}
	configFile := ConfigFile{}
	configFile.Path = cfgFile
	configDir = filepath.Dir(cfgFile)

	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return fld.Name
	})

	conf := ConfigParams{
		Binary:   cfgBinary,
		PlanFile: cfgPlanFile,
		MdFile:   cfgMdFile,
		Verbose:  false,
	}

	err := validate.Struct(conf)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			Logger.Errorf(" Field: %s, Error: %s, Param: %s\n", err.Field(), err.Tag(), err.Param())
			return err
		}
	}

	Logger.Debug("Config is valid")

	config, err := genConfig(conf)
	if err != nil {
		Logger.Error(err)
		return err
	}

	if createFile {
		// configFile.Path may be os.UserConfigDir + TpDir -- It may not exist
		// If it doesn't, we need to create the directory, prior to trying to
		// create the file
		configDirExists := doesExist(configDir)
		if !configDirExists {
			if err = os.MkdirAll(
				configDir, 0o750, //nolint:mnd
			); err != nil {
				Logger.Fatal(err)
				return err
			}
		}

		if !configExists {
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
			Logger.Debugf("Config is: \n%s\n", string(config))

			localNow = time.Now().Local().Format("200601021504")
			existingConfigFile := configFile.Path
			bkupConfigFile := configFile.Path + "-" + localNow
			// Create Backup
			err := backupFile(existingConfigFile, bkupConfigFile)
			if err != nil {
				Logger.Fatal(err)
				return err
			}
			Logger.Infof("Backup file %s created", bkupConfigFile)
			// Create New File
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
		Logger.Info(string(config))
	}
	return err
}
