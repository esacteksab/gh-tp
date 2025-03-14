// SPDX-License-Identifier: MIT

package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

func checkFilesByExtension(dir string, exts []string) bool {
	var exists bool
	for _, v := range exts {
		files, err := filepath.Glob(filepath.Join(dir, "*"+v))
		if err != nil {
			exists = false
			return exists
		}
		if len(files) > 0 {
			exists = true
			return exists
		}
	}
	return exists
}

func existsOrCreated(files []tpFile) error {
	for _, v := range files {
		if _, err := os.Stat(v.Name); err == nil {
			logger.Debugf("%s file %s was created.", v.Purpose, v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n", bold(green("✔")), v.Purpose, " Created...")
		} else if errors.Is(err, os.ErrNotExist) {
			//
			logger.Errorf("Markdown file %s was not created.", v.Name)
			fmt.Fprintf(color.Output, "%s  %s%s\n", bold(red("✕")), v.Purpose, " Failed to Create ...")
		} else {
			// I'm only human. NFC how you got here. I hope to never have to find out.
			logger.Errorf("If you see this error message, please open a bug. Error Code: TPE003. Error: %s", err)
		}
	}
	return err
}
