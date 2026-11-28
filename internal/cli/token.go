package cli

import (
	"os"
	"path/filepath"
)

// tokenPath возвращает путь к файлу с токеном.
func tokenPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gophkeeper", "token"), nil
}

// SaveToken сохраняет токен в файл с правами 0600.
func SaveToken(t string) error {
	p, err := tokenPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(t), 0o600)
}

// LoadToken читает токен; возвращает "" если файла нет.
func LoadToken() string {
	p, err := tokenPath()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return ""
	}
	return string(b)
}
