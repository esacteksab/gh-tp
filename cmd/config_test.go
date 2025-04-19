// SPDX-License-Identifier: MIT
package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGenConfig(t *testing.T) {
	// Setup test struct
	conf := ConfigParams{
		Binary:   "terraform",
		PlanFile: "test.out",
		MdFile:   "test.md",
		Verbose:  false,
	}

	// Call the function
	data, err := genConfig(conf)

	// Assert results
	require.NoError(t, err)
	require.NotNil(t, data)
	require.Contains(t, string(data), "terraform")
	require.Contains(t, string(data), "test.out")
	require.Contains(t, string(data), "test.md")
	require.Contains(t, string(data), "false")
}

func TestCreateConfig_ValidationPlanAndMdAreNotTheSame(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	testCases := []struct {
		name      string
		binary    string
		planFile  string
		mdFile    string
		expectErr bool
	}{
		{
			name:      "Different file passes validation",
			binary:    "terraform",
			planFile:  "plan.out",
			mdFile:    "plan.md",
			expectErr: false,
		},
		{
			name:      "Same file fails validation",
			binary:    "terraform",
			planFile:  "plan.out",
			mdFile:    "plan.out",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := ConfigParams{
				Binary:   tc.binary,
				PlanFile: tc.planFile,
				MdFile:   tc.mdFile,
				Verbose:  false,
			}

			// Validate the struct
			err := validate.Struct(conf)

			if tc.expectErr {
				require.Error(
					t,
					err,
					"Should return validation error when planFile and mdFile are the same",
				)
				if err != nil {
					validationErrs, ok := err.(validator.ValidationErrors)
					require.True(t, ok, "Should be validator.ValidationErrors")

					// check if any validation error is for the nefield constraint
					found := false

					for _, valErr := range validationErrs {
						if valErr.Tag() == "nefield" {
							found = true
							break
						}
					}
					require.True(t, found, "Should have 'nefield' validation error")
				} else {
					require.NoError(t, err, "Should not return an error when planFile and mdFile are unique")
				}
			}
		})
	}
}

func TestCreateConfig_ValidationPlanFileRequired(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	testCases := []struct {
		name      string
		binary    string
		planFile  string
		mdFile    string
		expectErr bool
	}{
		{
			name:      "Plan file is defined",
			binary:    "terraform",
			planFile:  "plan.out",
			mdFile:    "plan.md",
			expectErr: false,
		},
		{
			name:      "Plan file is not defined",
			binary:    "terraform",
			mdFile:    "plan.out",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := ConfigParams{
				Binary:   tc.binary,
				PlanFile: tc.planFile,
				MdFile:   tc.mdFile,
				Verbose:  false,
			}

			// Validate the struct
			err := validate.Struct(conf)

			if tc.expectErr {
				require.Error(t, err, "Should return validation error when planFile does not exist")
				if err != nil {
					validationErrs, ok := err.(validator.ValidationErrors)
					require.True(t, ok, "Should be validator.ValidationErrors")

					// check if any validation error is for the required constraint
					found := false

					for _, valErr := range validationErrs {
						if valErr.Tag() == "required" {
							found = true
							break
						}
					}
					require.True(t, found, "Should have 'required' validation error")
				} else {
					require.NoError(t, err, "Should not return an error when planFile exists")
				}
			}
		})
	}
}

func TestCreateConfig_ValidationMdFileRequired(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	testCases := []struct {
		name      string
		binary    string
		planFile  string
		mdFile    string
		expectErr bool
	}{
		{
			name:      "Markdown file is defined",
			binary:    "terraform",
			planFile:  "plan.out",
			mdFile:    "plan.md",
			expectErr: false,
		},
		{
			name:      "Markdown file is not defined",
			binary:    "terraform",
			planFile:  "plan.out",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := ConfigParams{
				Binary:   tc.binary,
				PlanFile: tc.planFile,
				MdFile:   tc.mdFile,
				Verbose:  false,
			}

			// Validate the struct
			err := validate.Struct(conf)

			if tc.expectErr {
				require.Error(t, err, "Should return validation error when planFile does not exist")
				if err != nil {
					validationErrs, ok := err.(validator.ValidationErrors)
					require.True(t, ok, "Should be validator.ValidationErrors")

					// check if any validation error is for the required constraint
					found := false

					for _, valErr := range validationErrs {
						if valErr.Tag() == "required" {
							found = true
							break
						}
					}
					require.True(t, found, "Should have 'required' validation error")
				} else {
					require.NoError(t, err, "Should not return an error when planFile exists")
				}
			}
		})
	}
}

func TestCreateConfig_ValidationExpectedBinary(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	testCases := []struct {
		name      string
		binary    string
		planFile  string
		mdFile    string
		expectErr bool
	}{
		{
			name:      "Terraform file is defined",
			binary:    "terraform",
			planFile:  "plan.out",
			mdFile:    "plan.md",
			expectErr: false,
		},
		{
			name:      "OpenTofu file is defined",
			binary:    "tofu",
			planFile:  "plan.out",
			mdFile:    "plan.md",
			expectErr: false,
		},
		{
			name:      "Neither Terraform or Tofu are defined",
			binary:    "fukd",
			planFile:  "fukd.out",
			mdFile:    "fukd.md",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf := ConfigParams{
				Binary:   tc.binary,
				PlanFile: tc.planFile,
				MdFile:   tc.mdFile,
				Verbose:  false,
			}

			// Validate the struct
			err := validate.Struct(conf)

			if tc.expectErr {
				require.Error(
					t,
					err,
					"Should return validation error when tofu or terraform is not the binary",
				)
				if err != nil {
					validationErrs, ok := err.(validator.ValidationErrors)
					require.True(t, ok, "Should be validator.ValidationErrors")

					// check if any validation error is for the oneof constraint
					found := false

					for _, valErr := range validationErrs {
						if valErr.Tag() == "oneof" {
							found = true
							break
						}
					}
					require.True(t, found, "Should have 'oneof' validation error")
				} else {
					require.NoError(t, err, "Should not return an error when tofu or terraform is the binary")
				}
			}
		})
	}
}

func TestCreateConfig_ValidationVerboseIsABool(t *testing.T) {
	// Setup validation
	validate := validator.New(validator.WithRequiredStructEnabled())

	// Test validation through direct tag validation
	// This allows us to test the constraint without fighting Go's type system

	err := validate.Var("not-a-boolean", "boolean")

	// A non-boolean value should cause a validation error
	require.Error(t, err, "Should return validation error when value is not a boolean")
	if err != nil {
		validationErrs, ok := err.(validator.ValidationErrors)
		require.True(t, ok, "Should be validator.ValidationErrors")

		// Check the tag of the validation error
		require.Equal(
			t,
			"boolean",
			validationErrs[0].Tag(),
			"Should have 'boolean' validation error",
		)
	}

	// Testing with actual boolean values (should pass)
	errTrue := validate.Var(false, "boolean")
	require.NoError(t, errTrue, "True value should pass boolean validation")

	errFalse := validate.Var(false, "boolean")
	require.NoError(t, errFalse, "False value should pass boolean validation")
}

// Make sure your mock types embed mock.Mock
type MockFileChecker struct {
	mock.Mock // This provides all the On(), Called(), etc. methods
}

func (m *MockFileChecker) DoesExist(cfgFile string) bool {
	args := m.Called(cfgFile)
	return args.Bool(0)
}

type MockUserPrompt struct {
	mock.Mock // This provides all the On(), Called(), etc. methods
}

func (m *MockUserPrompt) AskOverwrite(configExists bool) (bool, error) {
	args := m.Called(configExists)
	return args.Bool(0), args.Error(1)
}

func TestCreateOrOverwriteWithMock(t *testing.T) {
	// Setup - initialize the logger if needed
	if Logger == nil {
		// Initialize your logger here
		Logger = log.NewWithOptions(os.Stderr, log.Options{
			Level:           log.InfoLevel,
			ReportCaller:    true,
			ReportTimestamp: true,
		})
	}

	tempFile := "test-config"

	// Test case 1: Config doesn't exist, user chooses to create
	t.Run("ConfigDoesNotExist_UserCreates", func(t *testing.T) {
		// Create mocks
		mockFileChecker := new(MockFileChecker)
		mockUserPrompt := new(MockUserPrompt)

		// Set expectations
		mockFileChecker.On("DoesExist", tempFile).Return(false)
		mockUserPrompt.On("AskOverwrite", false).Return(true, nil)

		// Call the function with mocks
		configExists, createFile, err := createOrOverwrite(
			tempFile,
			mockFileChecker,
			mockUserPrompt,
		)

		// Assert results
		require.NoError(t, err)
		require.False(t, configExists)
		require.True(t, createFile)

		// Verify that the expectations were met
		mockFileChecker.AssertExpectations(t)
		mockUserPrompt.AssertExpectations(t)
	})
}

type MockFormRunner struct {
	createFilePtr *bool
	userSelection bool // What the "user" selected
	err           error
}

func (m *MockFormRunner) Run() error {
	// Set the createFile pointer to simulate user input
	if m.createFilePtr != nil {
		*m.createFilePtr = m.userSelection
	}
	return m.err
}

func TestCreateConfig(t *testing.T) {
	// Setup test logger if needed
	if Logger == nil {
		Logger = log.NewWithOptions(os.Stderr, log.Options{
			Level: log.InfoLevel,
		})
	}

	originalFactory := formRunnerFactory

	// Create temp directory for test files
	tmpDir := t.TempDir()
	cfgFile := filepath.Join(tmpDir, ".tp.toml")
	cfgMdFile := "test.md"
	cfgPlanFile := "test.out"
	cfgBinary := "terraform"

	defer os.Remove(".tp.toml")
	defer os.Remove("test.md")
	defer os.Remove("test.out")

	// Test case 1: New config file, creation successful
	t.Run("NewConfigCreationSuccess", func(t *testing.T) {
		formRunnerFactory = func(title string, createFile *bool, accessible bool) FormRunner {
			// Set the createFile value to true, simulating "Yes" selection
			*createFile = true
			return &MockFormRunner{err: nil}
		}

		// Restore original factory after test
		defer func() {
			formRunnerFactory = originalFactory
		}()
		// Mock the file checker and user prompt
		// Store original values
		originalFileChecker := defaultFileChecker
		originalUserPrompt := defaultUserPrompt

		// Create mocks
		mockFileChecker := new(MockFileChecker)
		mockUserPrompt := new(MockUserPrompt)

		// Replace globals with mocks
		defaultFileChecker = mockFileChecker
		defaultUserPrompt = mockUserPrompt

		// Restore after test
		defer func() {
			defaultFileChecker = originalFileChecker
			defaultUserPrompt = originalUserPrompt
		}()

		// Set up mock expectations with Anything matcher
		// Set up exact mock expectations
		mockFileChecker.On("DoesExist", cfgFile).Return(false)
		mockUserPrompt.On("AskOverwrite", false).Return(true, nil)

		// Call the function
		err := createConfig(cfgBinary, cfgFile, cfgMdFile, cfgPlanFile)

		// Debug - print actual calls
		// t.Logf("Mock file checker calls: %v", mockFileChecker.Calls)
		// t.Logf("Mock user prompt calls: %v", mockUserPrompt.Calls)

		// Assert results
		require.NoError(t, err)
		mockFileChecker.AssertExpectations(t)
		mockUserPrompt.AssertExpectations(t)
	})
}
