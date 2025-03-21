// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"strings"

	md "github.com/nao1215/markdown"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	planMd      *os.File
	planBody    *md.Markdown
	planDetails string
	planStr     string
	sb          strings.Builder
	sbDetails   strings.Builder
	sbPlan      string
	err         error
)

type SyntaxHighlight string

const (
	SyntaxHighlightTerraform SyntaxHighlight = "terraform"
)

func createMarkdown(mdParam, planStr string) (*os.File, string, error) {
	// Let's check to see if planStr is empty.
	// Don't need to create a Markdown file with no plan output.
	if len(planStr) == 0 {
		Logger.Debug("Plan Output is Empty.")
	} else {
		planMd, err = os.Create(mdParam)
		if err != nil {
			Logger.Errorf("failed to create Markdown: %s\n", err)
		}

		// Close the file when we're done with it
		defer func(planMd *os.File) {
			err := planMd.Close()
			if err != nil {
				Logger.Error(err)
			}
		}(planMd)

		// This has the plan wrapped in a code block in Markdown
		planBody = md.NewMarkdown(os.Stdout).
			CodeBlocks(
				md.SyntaxHighlight(SyntaxHighlightTerraform), planStr,
			)
		if err != nil {
			Logger.Errorf("error generating plan Markdown: %s\n", err)
		}

		// NewMarkdown returns io.Writer
		fmt.Fprintf(&sb, "\n%s\n", planBody)

		// This turns NewMarkdown io.Writer into a String, which .Details expects
		sbPlan = sb.String()

		// This block of terribleness creates a string of $Binary Plan
		titleCaseConverter = cases.Title(language.English)
		sbDetails.WriteString(titleCaseConverter.String(binary))
		sbDetails.WriteString(" ")
		sbDetails.WriteString("Plan")
		planDetails = sbDetails.String()
		// This is what creates the final document (`mdoutfile`) plmd here could possibly be os.Stdout one day
		mderr := md.NewMarkdown(planMd).Details(planDetails, sbPlan).Build()
		if mderr != nil {
			Logger.Errorf(
				"error generating %s markdown file, error: %s", mdParam,
				err,
			)
		}

		// planMd doesn't have a new line at eof, we need to give it one because Markdown
		file, err := os.OpenFile(
			"./"+mdParam, os.O_APPEND|os.O_WRONLY, 0o644, //nolint:mnd
		)
		if err != nil {
			Logger.Errorf("Unable to create Markdown: %s\n", err)
		}

		// Add new line
		_, err = file.WriteString("\n\n")
		if err != nil {
			Logger.Error(err)
		}

		// Close file
		err = file.Close()
		if err != nil {
			return nil, "", err
		}
	}
	return planMd, mdParam, err
}
