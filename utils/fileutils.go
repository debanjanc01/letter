package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func ListFilesWithFullPath(directoryPath string) ([]string, error) {
	files := make([]string, 0)

	fileInfos, err := os.ReadDir(directoryPath)
	if err != nil {
		return nil, err
	}

	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue // Skip directories
		}
		if strings.HasSuffix(fileInfo.Name(), "postman_collection.json") {
			files = append(files, filepath.Join(directoryPath, fileInfo.Name()))
		}
	}

	return files, nil
}

func GetCurrRunningPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	binaryDir := filepath.Dir(executable)

	return binaryDir, nil
}
