// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

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
