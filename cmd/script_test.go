// SPDX-License-Identifier: MIT

package cmd_test

import (
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestRoot(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "../testdata/script",
	})
}
