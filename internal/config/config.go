package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BotToken      string
	TempDir       string
	SofficePath   string
	PythonPath    string
	TesseractPath string
	AdminUserID   int64
}

// loadDotEnv reads .env file if present in working directory
func loadDotEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
	}
}

func Load() *Config {
	loadDotEnv()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		token = "8864347724:AAGFX8tBfX7SRKX6jDhX7x13Wo66P7RiMwo"
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
	if adminID == 0 {
		adminID = 8958346579 // Default Admin ID
	}

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
