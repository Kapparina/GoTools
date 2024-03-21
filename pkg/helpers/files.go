package helpers

import (
	"io"
	"log"
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

func MoveFile(src, dst string) (err error) {
	// Open original file.
	originalFile, err := os.Open(src)
	if err != nil {
		return err
	}

	// Create new file.
	newFile, err := os.Create(dst)
	if err != nil {
		err := originalFile.Close()
		if err != nil {
			return err
		}
		return err
	}

	// Copy the bytes to destination from source.
	_, err = io.Copy(newFile, originalFile)
	if err != nil {
		err := newFile.Close()
		if err != nil {
			return err
		}
		err = originalFile.Close()
		if err != nil {
			return err
		}
		return err
	}

	// Closes files
	err = newFile.Close()
	if err != nil {
		log.Printf("Failed to close new file: %v", err)
	}

	err = originalFile.Close()
	if err != nil {
		log.Printf("Failed to close original file: %v", err)
	}

	// Remove original file.
	err = os.Remove(src)
	if err != nil {
		return err
	}

	return nil
}
