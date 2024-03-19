package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// PathExists checks if a path exists or not.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// CheckExtension checks if the given file path has the given extension.
func CheckExtension(path string, extension string) bool {
	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}
	return filepath.Ext(path) == strings.ToLower(extension)
}
