package file

import (
	"fmt"
	"os"
	"path/filepath"
)

type fileService struct {
	path string
}

func NewWriter(folder string) fileService {
	return fileService{path: folder}
}

func (fs fileService) Write(data, fName string) (rErr error) {
	n := filepath.Join(fs.path, fName)
	file, err := os.OpenFile(n, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("file %v is not available: %w", n, err)
	}

	defer func() {
		err = file.Close()
		if err != nil {
			rErr = fmt.Errorf("failed to close file %v: %w", n, err)
		}
	}()

	_, err = file.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file %v: %w", n, err)
	}

	return nil
}

func (fs fileService) Read(fName string) (string, error) {
	n := filepath.Join(fs.path, fName)
	d, err := os.ReadFile(n)
	if err != nil {
		return "", fmt.Errorf("failed to read the file '%v': %w", n, err)
	}

	return string(d), nil
}
