// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/viper"
)

const (
	// https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/creating-a-pull-request-template-for-your-repository
	defaultPRTemplateName = "pull_request_template.md"
	// defaultPRTemplatesDir is the conventional directory for multiple PR templates.
	defaultPRTemplatesDir = "PULL_REQUEST_TEMPLATE"
)

// findPRTemplate looks in a handful of places based on the link above
// `pull_request_template.md` in the root of the repo or in `.github/` directory on in a `docs/` directory
// in a directory inside `.github` called `PULL_REQUEST_TEMPLATE` (e.g. .github/PULL_REQUEST_TEMPLATE/)
// or in the root of the repository or in the `docs/` directory (e.g. `docs/PULL_REQUEST_TEMPLATE`)
// returns: a slice of paths to be consumed by `init.go` eventually and an `err`
func findPRTemplate() (PRTemplatePaths []string, err error) {
	// cwd ("."), .github/ or docs/ in the root
	defaultPRTemplateRootDirs := []string{".", ".github", "docs"}
	PRTemplatePaths = []string{}

	for _, dir := range defaultPRTemplateRootDirs {
		dirEntries, readErr := os.ReadDir(dir)
		if readErr != nil {
			// Ignore missing directories; report other errors and continue scanning.
			if !os.IsNotExist(readErr) {
				Logger.Errorf(
					"Failed to read directory %s while searching for PR templates: %s",
					dir,
					readErr,
				)
			}
			continue
		}

		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}
			// GitHub treats PR template filenames case-insensitively.
			if strings.EqualFold(entry.Name(), defaultPRTemplateName) {
				path := filepath.Join(dir, entry.Name())
				PRTemplatePaths = append(PRTemplatePaths, path)
				Logger.Debugf("Found PR Template path: %s", path)
			}
		}

		// e.g. .github/PULL_REQUEST_TEMPLATE/, PULL_REQUEST_TEMPLATE/ or docs/PULL_REQUEST_TEMPLATE/
		nestedPathTemplates := filepath.Join(dir, defaultPRTemplatesDir)
		if doesExist(nestedPathTemplates) {
			PRTemplates, err := os.ReadDir(nestedPathTemplates)
			if err != nil {
				Logger.Errorf(
					"Unable to read the contents of the default PR template directory: %s",
					err,
				)
			}
			// read all files in PULL_REQUEST_TEMPLATE directory
			for _, template := range PRTemplates {
				// we only want files, no directories
				if !template.IsDir() {
					base := template.Name()
					templatePath := filepath.Join(nestedPathTemplates, base)
					if doesExist(templatePath) {
						Logger.Debugf("Found PR Template path: %s", templatePath)
						PRTemplatePaths = append(PRTemplatePaths, templatePath)
					}
				}
			}
		}
	}
	// sorting slice so we can remove duplicates with slices.Compact() below
	slices.Sort(PRTemplatePaths)
	return slices.Compact(PRTemplatePaths), nil
}

// createWithTemplate prefixes markdown plan content with a PR template body.
//
// Returns the combined template and plan markdown with proper spacing.
func createWithTemplate(templateFile, planMarkdown []byte) []byte {
	// let's add some padding to the top between the existing template body and the Terraform plan
	templateStr := string(templateFile)
	if !strings.HasSuffix(templateStr, "\n\n") {
		templateStr = strings.TrimRight(templateStr, "\n") + "\n\n"
	}
	return append([]byte(templateStr), planMarkdown...)
}

// getTemplateFromConfig checks for a template specified via flag or config,
// otherwise it falls back to default GitHub PR template locations.
// It returns the path to the selected template, or empty string if none is found,
// and an error if multiple templates are discovered or a specified template does not exist.
func getTemplateFromConfig() (string, error) {
	// Explicit template path has precedence over discovery.
	viperTemplate := strings.TrimSpace(viper.GetString("templateFile"))
	if viperTemplate != "" {
		if !doesExist(viperTemplate) {
			return "", fmt.Errorf("template file does not exist: %s", viperTemplate)
		}
		Logger.Debugf("Using template specified via flag or config: %s", viperTemplate)
		return viperTemplate, nil
	}

	// Discovery is opt-in for rollout safety.
	if !viper.GetBool("useTemplate") {
		Logger.Debug("Template discovery disabled; set useTemplate=true to enable default search")
		return "", nil
	}

	templates, err := findPRTemplate()
	if err != nil {
		return "", fmt.Errorf("failed to discover pull request templates: %w", err)
	}

	if len(templates) == 0 {
		Logger.Debug("No pull request template discovered in default locations")
		return "", nil
	}

	// Avoid silently choosing one when multiple templates are present.
	if len(templates) > 1 {
		return "", fmt.Errorf(
			"multiple pull request templates discovered (%s); set --templateFile to choose one",
			strings.Join(templates, ", "),
		)
	}

	Logger.Debugf("Using discovered pull request template: %s", templates[0])
	return templates[0], nil
}
