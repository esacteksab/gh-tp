// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/pelletier/go-toml/v2"
)

var (
	accessible   bool
	binary       string
	cfgBinary    string
	cfgFile      string
	cfgMdFile    string
	cfgPlanFile  string
	configDir    string
	configExists bool
	createFile   bool
	homeDir      string
	title        string
)

const TpDir = "gh-tp"

const ConfigName = ".tp.toml"

const DefaultXDGConfigDirName = ".config"

type ConfigParams struct {
	Binary   string `toml:"binary" comment:"binary: (type: string) The name of the binary, expect either 'tofu' or 'terraform'. Must exist on your $PATH." validate:"oneof=terraform tofu"`
	PlanFile string `toml:"planFile" comment:"planFile: (type: string) The name of the plan file created by 'gh tp'." validate:"required"`
	MdFile   string `toml:"mdFile" comment:"mdFile: (type: string) The name of the Markdown file created by 'gh tp'." validate:"required"`
	Verbose  bool   `toml:"verbose" comment:"verbose: (type: bool) Enable Verbose Logging. Default is false." validate:"boolean"`
}

func genConfig(conf ConfigParams) (data []byte, err error) {
	data, err = toml.Marshal(conf)
	if err != nil {
		Logger.Fatalf("Failed marshalling TOML: %s", err)
	}
	return data, err
}

// Checks the existence of a config file
// If one already exists, asks to overwrite
// If one does not exist, asks to create
func createOrOverwrite(cfgFile string) (configExists, createFile bool) {
	// configName = ".tp.toml"
	configExists = doesExist(cfgFile + "/" + ConfigName)
	Logger.Debugf("Using config: %s", cfgFile+ConfigName)
	createFile, err := query(configExists)
	if err != nil {
		Logger.Error(err)
	}

	return configExists, createFile
}

func query(configExists bool) (createFile bool, err error) {
	// Should we run in accessible mode?
	accessible, _ = strconv.ParseBool(os.Getenv("ACCESSIBLE"))

	if !configExists {
		title = "Create new file?"
	} else if configExists {
		title = "Overwrite existing config file?"
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes").
				Negative("No").
				Value(&createFile),
		)).WithTheme(huh.ThemeBase16()).
		WithAccessible(accessible)

	err = form.Run()
	if err != nil {
		Logger.Error(err)
	}

	return createFile, err
}

func createConfig(cfgBinary, cfgFile, cfgMdFile, cfgPlanFile string) error {
	configExists, createFile = createOrOverwrite(cfgFile)

	conf := ConfigParams{
		Binary:   cfgBinary,
		PlanFile: cfgPlanFile,
		MdFile:   cfgMdFile,
		Verbose:  false,
	}

	config, err := genConfig(conf)
	if err != nil {
		Logger.Error(err)
	}

	if createFile {
		if !configExists {
			// Figure out how to get in here
			Logger.Debugf("Inside configExists and 'config' is: %s", string(config))
			// Tracking this in #69
			err = os.WriteFile(cfgFile+"/.tp.toml", config, 0o600) //nolint:mnd    // https://go.dev/ref/spec#Integer_literals
			if err != nil {
				Logger.Fatalf("Error writing Config file: %s", err)
			}
		} else if configExists {
			Logger.Debugf("Config is: \n%s\n", string(config))
			localNow := time.Now().Local().Format("200601021504")

			existingConfigFile := cfgFile + "/" + ConfigName
			bkupConfigFile := cfgFile + "/" + ConfigName + "-" + localNow
			// Create Backup
			err := backupFile(existingConfigFile, bkupConfigFile)
			if err != nil {
				Logger.Fatal(err)
			}
			Logger.Infof("Backup file %s created", bkupConfigFile)
			// Create New File
			err = os.WriteFile(cfgFile+"/.tp.toml", config, 0o600) //nolint:mnd    // https://go.dev/ref/spec#Integer_literals
			if err != nil {
				Logger.Errorf("Error writing Config file: %s", err)
			}
		}
		Logger.Infof("Config file %s/.tp.toml created", cfgFile)
	} else if !createFile {
		Logger.Info(string(config))
	}
	return err
}
