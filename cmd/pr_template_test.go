// SPDX-License-Identifier: MIT

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// TestCreateWithTemplate verifies that template content is prepended to
// generated plan markdown with normalized spacing.
func TestCreateWithTemplate(t *testing.T) {
	template := []byte("## Header")
	plan := []byte("<details><summary>Terraform plan</summary>\n\ncontent\n")

	got := createWithTemplate(template, plan)
	want := "## Header\n\n<details><summary>Terraform plan</summary>\n\ncontent\n"

	if string(got) != want {
		t.Fatalf("createWithTemplate() mismatch\nwant:\n%s\n\ngot:\n%s", want, string(got))
	}
}

// TestGetTemplateFromConfig verifies explicit template handling and discovery
// behavior from GitHub default template locations.
func TestGetTemplateFromConfig(t *testing.T) {
	if Logger == nil {
		Logger = log.NewWithOptions(os.Stderr, log.Options{Level: log.InfoLevel})
	}

	tests := []struct {
		name        string
		setup       func(t *testing.T, tempDir string)
		want        string
		wantErr     bool
		wantErrPart string
	}{
		{
			name: "uses explicit template path from config",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				templatePath := filepath.Join(tempDir, "explicit-template.md")
				if err := os.WriteFile(templatePath, []byte("template"), 0o600); err != nil {
					t.Fatalf("failed to create explicit template: %v", err)
				}
				viper.Set("templateFile", templatePath)
			},
			wantErr: false,
		},
		{
			name: "returns error for explicit missing template",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				viper.Set("templateFile", filepath.Join(tempDir, "missing.md"))
			},
			wantErr:     true,
			wantErrPart: "template file does not exist",
		},
		{
			name: "discovers single default template",
			setup: func(t *testing.T, _ string) {
				t.Helper()
				viper.Set("useTemplate", true)
				if err := os.MkdirAll(".github", 0o755); err != nil {
					t.Fatalf("failed to create .github: %v", err)
				}
				if err := os.WriteFile(
					filepath.Join(".github", "pull_request_template.md"),
					[]byte("template"),
					0o600,
				); err != nil {
					t.Fatalf("failed to create discovered template: %v", err)
				}
			},
			want:    filepath.Join(".github", "pull_request_template.md"),
			wantErr: false,
		},
		{
			name: "errors when multiple default templates are discovered",
			setup: func(t *testing.T, _ string) {
				t.Helper()
				viper.Set("useTemplate", true)
				if err := os.WriteFile("pull_request_template.md", []byte("root"), 0o600); err != nil {
					t.Fatalf("failed to create root template: %v", err)
				}
				if err := os.MkdirAll(".github", 0o755); err != nil {
					t.Fatalf("failed to create .github: %v", err)
				}
				if err := os.WriteFile(
					filepath.Join(".github", "pull_request_template.md"),
					[]byte("github"),
					0o600,
				); err != nil {
					t.Fatalf("failed to create .github template: %v", err)
				}
			},
			wantErr:     true,
			wantErrPart: "multiple pull request templates discovered",
		},
		{
			name: "does not discover default template when useTemplate is false",
			setup: func(t *testing.T, _ string) {
				t.Helper()
				if err := os.MkdirAll(".github", 0o755); err != nil {
					t.Fatalf("failed to create .github: %v", err)
				}
				if err := os.WriteFile(
					filepath.Join(".github", "pull_request_template.md"),
					[]byte("template"),
					0o600,
				); err != nil {
					t.Fatalf("failed to create discovered template: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name: "uses explicit template when useTemplate is false",
			setup: func(t *testing.T, tempDir string) {
				t.Helper()
				templatePath := filepath.Join(tempDir, "explicit-template.md")
				if err := os.WriteFile(templatePath, []byte("template"), 0o600); err != nil {
					t.Fatalf("failed to create explicit template: %v", err)
				}
				viper.Set("useTemplate", false)
				viper.Set("templateFile", templatePath)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use an isolated working directory per test case.
			tempDir := t.TempDir()
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get cwd: %v", err)
			}
			if err := os.Chdir(tempDir); err != nil {
				t.Fatalf("failed to chdir: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chdir(cwd)
			})

			// Reset viper between cases to avoid key leakage.
			viper.Reset()
			t.Cleanup(viper.Reset)

			tt.setup(t, tempDir)

			got, err := getTemplateFromConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("getTemplateFromConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrPart != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrPart) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErrPart)
				}
			}
			if tt.want != "" && got != tt.want {
				t.Fatalf("getTemplateFromConfig() = %q, want %q", got, tt.want)
			}
			if tt.want == "" && !tt.wantErr && got == "" {
				return
			}
			if !tt.wantErr && got == "" {
				t.Fatalf("expected discovered template path, got empty")
			}
		})
	}
}
