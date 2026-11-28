package cli

import (
	"encoding/base64"
	"fmt"
	"os"
)

// maxBinaryFileSize — максимальный размер файла для --file (10 МБ).
const maxBinaryFileSize = 10 << 20 // 10 MB

// readFileAsBase64 читает файл и возвращает base64-строку.
func readFileAsBase64(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("%s is a directory", path)
	}
	if info.Size() > maxBinaryFileSize {
		return "", fmt.Errorf("file %s is too large: %d bytes (max %d)", path, info.Size(), maxBinaryFileSize)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}
