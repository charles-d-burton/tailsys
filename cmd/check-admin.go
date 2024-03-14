//go:build !windows
// +build !windows

package cmd

import "os"

// Check if admin
func Check() bool {
	return os.Getuid() == 0
}
