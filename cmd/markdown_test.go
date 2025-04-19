// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func Test_createMarkdown(t *testing.T) {
	// Setup test logger (same as before)
	if Logger == nil {
		Logger = log.NewWithOptions(os.Stderr, log.Options{Level: log.InfoLevel})
	}

	baseTestDir, err := os.MkdirTemp("", "gh_tp_markdown_test_")
	if err != nil {
		t.Fatalf("Failed to create base test directory: %v", err)
	}
	defer os.RemoveAll(baseTestDir)

	type args struct {
		mdParam    string
		planStr    string
		binaryName string
	}
	tests := []struct {
		name        string
		args        args
		wantPath    string // Expected returned path
		wantErr     bool
		wantErrMsg  string   // Optional: Check for specific error message content
		wantContent []string // Keep this for checking file content on success
	}{
		{
			name: "empty plan",
			args: args{mdParam: "empty_plan.md", planStr: "", binaryName: "terraform"},
			// Should return the validated filename even if no file is written
			wantPath: "empty_plan.md",
			wantErr:  false,
		},
		{
			name: "no changes",
			args: args{
				mdParam:    "no_changes_plan.md",
				planStr:    `Plan content here.`, // Simplified content for example
				binaryName: "terraform",
			},
			wantPath: "no_changes_plan.md", // Expect simple filename
			wantErr:  false,
			wantContent: []string{
				"<details><summary>Terraform plan</summary>",
				"```terraform",
				"Plan content here.",
				"</details>",
			},
		},
		{
			name: "with changes - tofu",
			args: args{
				mdParam:    "changes_plan_tofu.md",
				planStr:    "+ resource \"test\"",
				binaryName: "tofu",
			},
			wantPath: "changes_plan_tofu.md", // Expect simple filename
			wantErr:  false,
			wantContent: []string{
				"<details><summary>OpenTofu plan</summary>",
				"```terraform",
				"+ resource",
				"</details>",
			},
		},
		// --- Validation Failure Cases ---
		{
			name: "invalid filename - contains slash",
			args: args{mdParam: "invalid/name.md", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input on validation failure
			wantPath:   "invalid/name.md",
			wantErr:    true,
			wantErrMsg: "must be a filename only",
		},
		{
			name: "invalid filename - empty string",
			args: args{mdParam: "", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input
			wantPath:   "",
			wantErr:    true,
			wantErrMsg: "filename cannot be empty",
		},
		{
			name: "invalid filename - contains null",
			args: args{mdParam: "invalid\x00name.md", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input
			wantPath: "invalid\x00name.md",
			wantErr:  true,
			// --- CORRECTED wantErrMsg ---
			// The null byte fails the character regex check first.
			wantErrMsg: "contains invalid characters",
		},
		{
			name: "invalid filename - is dot",
			args: args{mdParam: ".", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input
			wantPath:   ".",
			wantErr:    true,
			wantErrMsg: "must be a filename only",
		},
		{
			name: "invalid filename - contains invalid chars",
			args: args{mdParam: "bad<char>.md", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input
			wantPath:   "bad<char>.md",
			wantErr:    true,
			wantErrMsg: "contains invalid characters",
		},
		{
			name: "invalid filename - too long",
			args: args{mdParam: strings.Repeat("a", 300) + ".md", planStr: "plan", binaryName: "t"},
			// Should return the *original* invalid input
			wantPath:   strings.Repeat("a", 300) + ".md",
			wantErr:    true,
			wantErrMsg: "exceeds maximum length",
		},
	}

	for _, tt := range tests {
		// Test setup (create dir, cd, cleanup) remains the same
		testRunDir := filepath.Join(baseTestDir, tt.name)
		err := os.MkdirAll(testRunDir, 0o755)
		if err != nil {
			t.Fatalf("[%s] Failed to create test run directory %s: %v", tt.name, testRunDir, err)
		}
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("[%s] Failed to get cwd: %v", tt.name, err)
		}
		err = os.Chdir(testRunDir)
		if err != nil {
			t.Fatalf("[%s] Failed to change dir: %v", tt.name, err)
		}
		t.Cleanup(func() { os.Chdir(cwd) })

		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := createMarkdown(tt.args.mdParam, tt.args.planStr, tt.args.binaryName)

			// 1. Check error status
			if (err != nil) != tt.wantErr {
				t.Fatalf("createMarkdown() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 2. Check error message content if error was expected
			if tt.wantErr && tt.wantErrMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("Expected error containing %q, but got: %v", tt.wantErrMsg, err)
				}
			}

			// 3. Check returned path - This should *always* match wantPath now,
			//    regardless of error status, because the function returns either the
			//    validated path on success/skip, or the original invalid path on validation error.
			if gotPath != tt.wantPath {
				t.Errorf("createMarkdown() returned path = %q, want %q", gotPath, tt.wantPath)
			}

			// 4. Check file existence/content only if no error AND plan was not empty
			if !tt.wantErr && tt.args.planStr != "" {
				// File should exist - use gotPath (which is the validated filename)
				fileInfo, statErr := os.Stat(gotPath)
				if statErr != nil {
					t.Fatalf("Expected file %q to exist, but stat failed: %v", gotPath, statErr)
				}
				if fileInfo.IsDir() {
					t.Fatalf("Expected %q to be a file, but it's a directory", gotPath)
				}
				// Check content (same as before)
				if len(tt.wantContent) > 0 {
					contentBytes, readErr := os.ReadFile(gotPath)
					if readErr != nil {
						t.Fatalf("Failed to read file %q: %v", gotPath, readErr)
					}
					contentStr := string(contentBytes)
					for _, sub := range tt.wantContent {
						if !strings.Contains(contentStr, sub) {
							t.Errorf(
								"File %q: Expected content to contain %q, but didn't.\n--- Content ---\n%s\n---------------",
								gotPath,
								sub,
								contentStr,
							)
						}
					}
				}
			} else if !tt.wantErr && tt.args.planStr == "" {
				// Empty plan, no error: File should NOT exist
				if _, statErr := os.Stat(gotPath); statErr == nil {
					t.Errorf("Expected file %q *not* to exist for empty plan, but it does", gotPath)
					_ = os.Remove(gotPath) // Clean up
				} else if !os.IsNotExist(statErr) {
					t.Errorf("Unexpected error stating file %q for empty plan check: %v", gotPath, statErr)
				}
			}
			// No file check needed if tt.wantErr is true
		})
	}
}
