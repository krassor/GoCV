package utils

import (
	"fmt"
	"os"
	"path"
)

// listDir return slice of files path in srcPath, excluding within directories.
// Return error if directory does not exist
func ListDir(srcPath string, filter func(fileName string) bool) (fileList []string, err error) {

	files, err := os.ReadDir(srcPath) //ioutil.ReadDir(srcPath)
	if err != nil {
		return nil, fmt.Errorf("Error read padth: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		} else {
			fileName := file.Name()
			if filter(fileName) {
				fileList = append(fileList, path.Join(srcPath, fileName))
			}

		}
	}

	if len(fileList) <= 0 {
		return nil, fmt.Errorf("No match in this dir")
	}
	return fileList, nil
}
