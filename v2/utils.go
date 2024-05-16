package toolkit

import (
	"errors"
	"os"
	"path/filepath"
)

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func getAllFiles(path string) ([]string, error) {
	// list all the files in the folder
	filepaths := []string{}
	files, err := os.ReadDir(path)
	if err != nil {
		return []string{}, err
	}
	for _, file := range files {
		f := filepath.Join(path, file.Name())
		filepaths = append(filepaths, f)
	}
	return filepaths, nil
}
