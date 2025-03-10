// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	accessible bool
	binary     string
	cfgFile    string
	configDir  string
	createFile bool
	homeDir    string
	mdFile     string
	planFile   string
)

type ConfigParams struct {
	Binary   string `toml:"binary" comment:"binary: (type: string) The name of the binary, either 'tofu' or 'terraform'. Must exist on your $PATH."`
	PlanFile string `toml:"planFile" comment:"planFile: (type: string) The name of the plan file created by 'gh tp'."`
	MdFile   string `toml:"mdFile" comment:"mdFile: (type: string) The name of the Markdown file created by 'gh tp'."`
	Verbose  bool   `toml:"verbose" comment:"verbose: (type: bool) Enable Verbose Logging. Default is false."`
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Simple terminal-based form to generate a config file for tp",
	Long: `A CLI prompt-based form with some suggested values to generate a config file for tp
'gh tp init' with no flags or arguments will instantiate the prompt-based form. File will be created
in the location selected. Order of lookups is:
    '$HOME/.config/.tp.toml'
	'$HOME/.tp.toml'
	'.tp.toml' (Project's Root).`,
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, configDir, cwd, err := getDirectories()
		if err != nil {
			log.Fatalf("Error: %s", err)
		}

		// Should we run in accessible mode?
		accessible, _ = strconv.ParseBool(os.Getenv("ACCESSIBLE"))

		form := huh.NewForm(
			huh.NewGroup(

				huh.NewSelect[string]().
					Title("Where would you like to save your .tp.toml config file?").
					Options(
						huh.NewOption("Home Config Directory: "+configDir+"/.tp.toml",
							configDir).Selected(true),
						huh.NewOption("Home Directory: "+homeDir+"/.tp.toml", homeDir),
						huh.NewOption("Project Root:"+".tp.toml", cwd),
					).Value(&cfgFile),

				// It could make sense some day to do a `gh tp init --binary`
				huh.NewSelect[string]().
					Title("Choose your binary").
					Options(
						huh.NewOption("OpenTofu", "tofu"),
						huh.NewOption("Terraform", "terraform").Selected(true),
					).Value(&binary),

				huh.NewInput().
					Title("What do you want the name of your plan's output file to be? ").
					Placeholder("example: tpplan.out tp.out tp.plan plan.out out.plan ...").
					Suggestions([]string{"tpplan.out", "tp.out", "tp.plan", "plan.out", "out.plan"}).
					Value(&planFile).
					Validate(func(pf string) error {
						if pf == "" {
							return errors.New("This field is requiured. Please enter what your plan's output file should be named.") //nolint:stylecheck
						}
						return nil
					}),

				huh.NewInput().
					Title("What do you want the name of your Markdown file to be?  ").
					Suggestions([]string{"tpplan.md", "tp.md", "plan.md"}).
					Placeholder("example: tpplan.md tp.md plan.md ...").
					Value(&mdFile).
					Validate(func(md string) error {
						if md == "" {
							return errors.New("This field is requiured. Please enter what your Markdown file should be named.") //nolint:stylecheck
						}
						return nil
					}),

				huh.NewConfirm().
					Title("Create file[Y], or write to stdout[N]?").
					Value(&createFile),
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
			log.Warn(runerr)
			os.Exit(1)
		}

		conf := ConfigParams{
			Binary:   binary,
			PlanFile: planFile,
			MdFile:   mdFile,
			Verbose:  false,
		}

		config, err := genConfig(conf)
		if err != nil {
			log.Fatal(err)
		}

		if createFile {
			err = os.WriteFile(cfgFile+"/.tp.toml", config, 0o600) //nolint:mnd    // https://go.dev/ref/spec#Integer_literals
			if err != nil {
				log.Fatalf("Error writing Config file: %s", err)
			}
		} else {
			fmt.Println(string(config))
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
