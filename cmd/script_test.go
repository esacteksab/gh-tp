// SPDX-License-Identifier: MIT

package cmd_test

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestRoot(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir: "../testdata/script",
		Setup: func(env *testscript.Env) error {
			if os.Getenv("GOCOVERDIR") != "" {
				env.Vars = append(env.Vars, "GOCOVERDIR="+os.Getenv("GOCOVERDIR"))
			}
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){},
	})
}
