// +build !windows

package cmd

import (
	"os"
	"path/filepath"
)

func removeDiskLayer(location, foldername string) error {
	return os.RemoveAll(filepath.Join(location, foldername))
}
