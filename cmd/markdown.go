// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"strings"

	md "github.com/nao1215/markdown"
)

// SyntaxHighlight represents the language identifier used for syntax
// highlighting in markdown.
type SyntaxHighlight string

const (
	// SyntaxHighlightTerraform is the syntax highlighting identifier for
	// Terraform/OpenTofu code.
	SyntaxHighlightTerraform SyntaxHighlight = "terraform"
)

// createMarkdown generates a GitHub Flavored Markdown document containing the
// Terraform/OpenTofu plan output.
//
// Parameters:
//
//	mdParam - The desired filename for the markdown document. MUST be a base filename without directory separators and using only allowed characters.
//	planStr - The human-readable plan output from createPlan() or stdin.
//	binaryName - The name of the binary used ("terraform" or "tofu") for the title.
//
// Returns:
//
//	string - The validated filename used.
//	error - Any error encountered during markdown generation or validation, or nil on success.
func createMarkdown(mdParam, planStr, binaryName string) (string, error) {
	// Use local variables
	var sbPlanBuilder strings.Builder

	Logger.Debugf(
		"createMarkdown called for binary: %s, output file parameter: %q",
		binaryName,
		mdParam,
	)

	// If we reach here, validatedFilename is considered safe and is just the filename.
	validatedFilename, err := validateFilePath(mdParam)
	if err != nil {
		return mdParam, err
	}

	if len(planStr) == 0 {
		Logger.Debugf(
			"Plan output is empty. Skipping Markdown file creation for %q.",
			validatedFilename,
		)
		// Return the validated path, indicating it wasn't processed, and no error.
		return validatedFilename, nil
	}

	// Prepare Markdown Content
	codeBlockMarkdown := md.NewMarkdown(&sbPlanBuilder)
	err = codeBlockMarkdown.CodeBlocks(
		md.SyntaxHighlight(SyntaxHighlightTerraform), planStr,
	).Build()
	if err != nil {
		Logger.Errorf("Internal error generating markdown code block: %v", err)
		return validatedFilename, fmt.Errorf("markdown generation failed (code block): %w", err)
	}
	sbPlan := sbPlanBuilder.String()

	title := ""
	switch strings.ToLower(binaryName) {
	case "tofu":
		title = "OpenTofu plan"
	case "terraform":
		title = "Terraform plan"
	default:
		title = "Plan Details"
		Logger.Warnf("Unknown binary name '%s', using default markdown title.", binaryName)
	}
	Logger.Debugf("Markdown details title: %s", title)

	Logger.Debugf("Attempting to create/write markdown file: %s", validatedFilename)

	// Use the validatedFilename directly - it's just the filename for the current dir.
	planMdFile, err := os.Create( //nolint:gosec // validateFilename is sanitized by validateFilePath
		validatedFilename,
	)
	if err != nil {
		Logger.Errorf("Failed to create markdown file '%s': %v", validatedFilename, err)
		return validatedFilename, fmt.Errorf(
			"failed to create markdown file %s: %w",
			validatedFilename,
			err,
		)
	}
	defer func() {
		if closeErr := planMdFile.Close(); closeErr != nil {
			Logger.Errorf("Error closing markdown file '%s': %v", validatedFilename, closeErr)
		} else {
			Logger.Debugf("Closed markdown file: %s", validatedFilename)
		}
	}()

	// Build final markdown directly into the file handle
	finalMarkdown := md.NewMarkdown(planMdFile)
	buildErr := finalMarkdown.Details(title, "\n"+sbPlan+"\n").Build()
	if buildErr != nil {
		Logger.Errorf(
			"Failed to write <details> block to markdown file '%s': %v",
			validatedFilename,
			buildErr,
		)
		return validatedFilename, fmt.Errorf(
			"failed to write markdown content to %s: %w",
			validatedFilename,
			buildErr,
		)
	}

	// Add final newline to mdFile
	_, err = planMdFile.WriteString("\n")
	if err != nil {
		Logger.Errorf(
			"Failed to write final newline to markdown file '%s': %v",
			validatedFilename,
			err,
		)
		return validatedFilename, fmt.Errorf(
			"failed write final newline to %s: %w",
			validatedFilename,
			err,
		)
	}

	Logger.Debugf("Successfully wrote markdown content to %s", validatedFilename)
	// Return the validatedFilename used and nil error on success
	return validatedFilename, nil
}
