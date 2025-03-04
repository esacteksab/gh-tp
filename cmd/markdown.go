// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	md "github.com/nao1215/markdown"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	planMd      *os.File
	planBody    *md.Markdown
	planDetails string
	planPath    string
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
	planMd, err = os.Create(mdParam)
	if err != nil {
		log.Errorf("failed to create Markdown: %s\n", err)
	}
	// Close the file when we're done with it
	defer planMd.Close()

	// This has the plan wrapped in a code block in Markdown

	planBody = md.NewMarkdown(os.Stdout).
		CodeBlocks(md.SyntaxHighlight(SyntaxHighlightTerraform), planStr)
	if err != nil {
		log.Errorf("error generating plan Markdown: %s\n", err)
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
		log.Errorf("error generating %s markdown file, error: %s", mdParam, err)
	}

	// Checking to see if Markdown file was created.
	if _, err := os.Stat(mdParam); err == nil {
		log.Debugf("Markdown file %s was created.", mdParam)
		fmt.Fprintf(color.Output, "%s%s\n", bold(green("✔")), "  Markdown Created...")
	} else if errors.Is(err, os.ErrNotExist) {
		//
		log.Errorf("Markdown file %s was not created.", mdParam)
		fmt.Fprintf(color.Output, "%s%s\n", bold(red("✕")), "  Failed to Create Markdown...")
	} else {
		// I'm only human. NFC how you got here. I hope to never have to find out.
		log.Errorf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
	}

	return planMd, mdParam, mderr
}
