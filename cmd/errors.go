// SPDX-License-Identifier: MIT

package cmd

import "errors"

// ErrInterrupted indicates that the operation was cancelled by the user (e.g., Ctrl+C).
var ErrInterrupted = errors.New("operation interrupted by user")
