package file

import (
	"fmt"
	"os"
	"path/filepath"
)

type writer struct {
	path string
}

func NewWriter(folder string) writer {
	return writer{path: folder}
}

func (w writer) Write(data, fName string) (rErr error) {
	n := filepath.Join(w.path, fName)
	file, err := os.OpenFile(n, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("file %v is not available: %w", n, err)
	}

	defer func() {
		err = file.Close()
		rErr = fmt.Errorf("failed to close file %v: %w", n, err)
	}()

	_, err = file.WriteString(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file %v: %w", n, err)
	}

	return nil
}
