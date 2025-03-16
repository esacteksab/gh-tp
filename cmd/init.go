// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	accessible  bool
	binary      string
	cfgBinary   string
	cfgFile     string
	cfgMdFile   string
	cfgPlanFile string
	configDir   string
	createFile  bool
	configName  string
	homeDir     string
	noConfig    bool
)

type ConfigParams struct {
	Binary   string `toml:"binary" comment:"binary: (type: string) The name of the binary, expect either 'tofu' or 'terraform'. Must exist on your $PATH."`
	PlanFile string `toml:"planFile" comment:"planFile: (type: string) The name of the plan file created by 'gh tp'."`
	MdFile   string `toml:"mdFile" comment:"mdFile: (type: string) The name of the Markdown file created by 'gh tp'."`
	Verbose  bool   `toml:"verbose" comment:"verbose: (type: bool) Enable Verbose Logging. Default is false."`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:               "init",
	Aliases:           []string{"i"},
	SilenceUsage:      true,
	SilenceErrors:     true,
	Args:              cobra.NoArgs,
	ValidArgsFunction: cobra.NoFileCompletions,
	Short:             "A interactive prompt-based form to generate a config file for tp.",
	Long: heredoc.Doc(`
		An interactive prompt-based form with some suggested values to generate a config file for tp.
		File will be created in the one of the following locations:
		Order of lookups is:
			1. A .tp.toml file in your project's root
			2. $XDG_CONFIG_HOME/.tp.toml
			3. $HOME/.tp.toml)

		View docs at https://github.com/esacteksab/gh-tp for more information.`),
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, configDir, cwd, err := getDirectories()
		if err != nil {
			logger.Fatalf("Error: %s", err)
		}

		// Should we run in accessible mode?
		accessible, _ = strconv.ParseBool(os.Getenv("ACCESSIBLE"))

		form := huh.NewForm(
			huh.NewGroup(

				huh.NewSelect[string]().
					Title("Where would you like to save your .tp.toml config file?").
					Options(
						huh.NewOption("Project Root:"+".tp.toml", cwd).Selected(true),
						huh.NewOption("Home Config Directory: "+configDir+"/.tp.toml",
							configDir),
						huh.NewOption("Home Directory: "+homeDir+"/.tp.toml", homeDir),
					).Value(&cfgFile),

				// It could make sense some day to do a `gh tp init --binary`
				huh.NewSelect[string]().
					Title("Choose your binary").
					Options(
						huh.NewOption("OpenTofu", "tofu"),
						huh.NewOption("Terraform", "terraform").Selected(true),
					).Value(&cfgBinary),

				huh.NewInput().
					Title("What do you want the name of your plan's output file to be? ").
					Placeholder("example: tpplan.out tp.out tp.plan plan.out out.plan ...").
					Suggestions([]string{"tpplan.out", "tp.out", "tp.plan", "plan.out", "out.plan"}).
					Value(&cfgPlanFile).
					Validate(func(pf string) error {
						if pf == "" {
							//lint:ignore ST1005 It's a user-facing error message. I want pretty!
							return errors.New("This field is required. Please enter what your plan's output file should be named.") //nolint:stylecheck
						}
						return nil
					}),

				huh.NewInput().
					Title("What do you want the name of your Markdown file to be?  ").
					Suggestions([]string{"tpplan.md", "tp.md", "plan.md"}).
					Placeholder("example: tpplan.md tp.md plan.md ...").
					Value(&cfgMdFile).
					Validate(func(md string) error {
						if md == "" {
							//lint:ignore ST1005 It's a user-facing error message. I want pretty!
							return errors.New("This field is required. Please enter what your Markdown file should be named.") //nolint:stylecheck
						}
						return nil
					}),
			),
		).WithTheme(huh.ThemeBase16()).
			// Just in case https://raw.githubusercontent.com/charmbracelet/huh/refs/tags/v0.6.0/keymap.go
			// https://github.com/charmbracelet/huh/issues/73
			WithKeyMap(&huh.KeyMap{
				Quit: key.NewBinding(key.WithKeys("q", "esc"), key.WithHelp("q", "quit")),
				Input: huh.InputKeyMap{
					AcceptSuggestion: key.NewBinding(key.WithKeys("tab", "enter"), key.WithHelp("tab", "accept")),
					Prev:             key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
					Next:             key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "next")),
					Submit:           key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
				},
				Select: huh.SelectKeyMap{
					Prev:         key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
					Next:         key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "select")),
					Submit:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
					Up:           key.NewBinding(key.WithKeys("up", "k", "ctrl+k", "ctrl+p"), key.WithHelp("↑", "up")),
					Down:         key.NewBinding(key.WithKeys("down", "j", "ctrl+j", "ctrl+n"), key.WithHelp("↓", "down")),
					Left:         key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("←", "left"), key.WithDisabled()),
					Right:        key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("→", "right"), key.WithDisabled()),
					Filter:       key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
					SetFilter:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "set filter"), key.WithDisabled()),
					ClearFilter:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear filter"), key.WithDisabled()),
					HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "½ page up")),
					HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "½ page down")),
					GotoTop:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("g/home", "go to start")),
					GotoBottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("G/end", "go to end")),
				},
				Confirm: huh.ConfirmKeyMap{
					Prev:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
					Next:   key.NewBinding(key.WithKeys("enter", "tab"), key.WithHelp("enter", "next")),
					Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
					Toggle: key.NewBinding(key.WithKeys("h", "l", "right", "left"), key.WithHelp("←/→", "toggle")),
					Accept: key.NewBinding(key.WithKeys("y", "Y"), key.WithHelp("y", "Yes")),
					Reject: key.NewBinding(key.WithKeys("n", "N"), key.WithHelp("n", "No")),
				},
			}).WithShowHelp(true).WithShowErrors(true).WithAccessible(accessible)

		runerr := form.Run()
		if runerr != nil {
			logger.Warn(runerr)
			os.Exit(1)
		}

		noConfig, createFile = mkFile(cfgFile)

		conf := ConfigParams{
			Binary:   cfgBinary,
			PlanFile: cfgPlanFile,
			MdFile:   cfgMdFile,
			Verbose:  false,
		}

		config, err := genConfig(conf)
		if err != nil {
			logger.Fatal(err)
		}

		if createFile {
			if noConfig {
				// #69 Figure out how to get in here logger.Debugf("Inside noConfig and 'config' is: %s", string(config))
				err = os.WriteFile(cfgFile+"/.tp.toml", config, 0o600) //nolint:mnd    // https://go.dev/ref/spec#Integer_literals
				if err != nil {
					logger.Fatalf("Error writing Config file: %s", err)
				}

			} else if !noConfig {
				// #69 like above: logger.Debugf("Inside !noConfig and 'config' is: %s", string(config))
				// epoch as an int64
				e := time.Now().Unix()
				// epoch as string
				epoch := strconv.FormatInt(e, 10)

				existingConfigFile := cfgFile + "/" + configName
				bkupConfigFile := cfgFile + "/" + configName + "-bkup-" + epoch
				// Create Backup
				err := backupFile(existingConfigFile, bkupConfigFile)
				if err != nil {
					logger.Fatal(err)
				}
				// Create New File
				err = os.WriteFile(cfgFile+"/.tp.toml", config, 0o600) //nolint:mnd    // https://go.dev/ref/spec#Integer_literals
				if err != nil {
					logger.Fatalf("Error writing Config file: %s", err)
				}
				logger.Infof("Config file %s/.tp.toml created", cfgFile)
			}
		} else if !createFile {
			logger.Info(string(config))
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
