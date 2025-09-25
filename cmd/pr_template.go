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
	defaultPRTemplateDirs = "PULL_REQUEST_TEMPLATE"
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
		// e.g. .pull_request_template.md, .github/pull_request_template.md or docs/pull_request_template.md
		path := filepath.Join(dir, defaultPRTemplateName)
		if doesExist(path) {
			PRTemplatePaths = append(PRTemplatePaths, path)
			Logger.Debugf("Found PR Template path: %s", path)
		}
		// e.g. .github/PULL_REQUEST_TEMPLATE/, PULL_REQUEST_TEMPLATE/ or docs/PULL_REQUEST_TEMPLATE/
		nestedPathTemplates := filepath.Join(dir, defaultPRTemplateDirs)
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

func createWithTemplate(
	validatedFilename string,
	templateFile []byte,
	planMdFile *os.File,
) (string, error) {
	// let's add some padding to the top between the existing template body and the Terraform plan
	templateStr := string(templateFile)
	if !strings.HasSuffix(templateStr, "\n\n\n") {
		templateStr = strings.TrimRight(templateStr, "\n") + "\n\n"
	}
	// read planMdFile for it's contents
	planMdBytes, err := os.ReadFile(planMdFile.Name())
	if err != nil {
		Logger.Errorf("Unable to read Markdown file: %s", err)
		return validatedFilename, fmt.Errorf("failed to read markdown file: %w", err)
	}
	combined := append([]byte(templateStr), planMdBytes...)
	err = os.WriteFile(planMdFile.Name(), combined, 0o600) //nolint:mnd
	if err != nil {
		Logger.Errorf("failed to write combined template and markdown: %s", err)
		return validatedFilename, fmt.Errorf(
			"failed to write combined template and markdown: %w",
			err,
		)
	}
	return validatedFilename, nil
}

// getTemplateFromConfig checks for a binary specified via flag or config.
func getTemplateFromConfig() (string, error) {
	v := viper.IsSet("templateFile")
	Logger.Debugf("Template is set: %v", v)
	viperTemplate := viper.GetString("templateFile")
	if viperTemplate == "" {
		return "", nil // Not set
	}
	Logger.Debugf("Using template specified via flag or config: %s", viperTemplate)
	return viperTemplate, nil
}
