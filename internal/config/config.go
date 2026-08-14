package config

import (
	"os"
	"strconv"
)

type Config struct {
	BotToken      string
	TempDir       string
	SofficePath   string
	PythonPath    string
	TesseractPath string
	AdminUserID   int64
}

func Load() *Config {
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		token = "8885263927:AAExdabNfuKcrr9hG44-B_8Rf5mXv64i-xo"
	}

	tempDir := os.Getenv("TEMP_DIR")
	if tempDir == "" {
		tempDir = "/tmp/ilovepdf"
	}

	soffice := os.Getenv("SOFFICE_PATH")
	if soffice == "" {
		soffice = "/usr/local/bin/soffice"
	}

	python := os.Getenv("PYTHON_PATH")
	if python == "" {
		python = "/usr/local/bin/python3"
	}

	tesseract := os.Getenv("TESSERACT_PATH")
	if tesseract == "" {
		tesseract = "/opt/homebrew/bin/tesseract"
	}

	adminIDStr := os.Getenv("ADMIN_USER_ID")
	adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)

	_ = os.MkdirAll(tempDir, 0755)

	return &Config{
		BotToken:      token,
		TempDir:       tempDir,
		SofficePath:   soffice,
		PythonPath:    python,
		TesseractPath: tesseract,
		AdminUserID:   adminID,
	}
}
