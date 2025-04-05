// SPDX-License-Identifier: MIT
package cmd

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

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

	for _, tc := range testCases { //nolint:dupl
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

	for _, tc := range testCases { //nolint:dupl
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

	for _, tc := range testCases { //nolint:dupl
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

	for _, tc := range testCases { //nolint:dupl
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
