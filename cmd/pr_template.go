// SPDX-License-Identifier: MIT

package cmd

import (
	"github.com/spf13/viper"
)

const (
	// default name of PR template
	pr_template_name = "pull_request_template.md"
	// directory inside .github/ to support multiple PR templates
	pr_template_dirs = "PULL_REQUEST_TEMPLATE"
)

// getTemplateFromConfig checks for a binary specified via flag or config.
func getTemplateFromConfig() (string, error) {
	v := viper.IsSet("template")
	Logger.Debugf("Template is set: %v", v)
	viperTemplate := viper.GetString("template")
	if viperTemplate == "" {
		return "", nil // Not set
	}

	Logger.Debugf("Using template specified via flag or config: %s", viperTemplate)
	return viperTemplate, nil
}
