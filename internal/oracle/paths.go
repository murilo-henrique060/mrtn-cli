package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// DefaultRepoURL is the official public repository containing test cases and expected outputs.
	DefaultRepoURL = "https://github.com/murilo-henrique060/mrtn-oracle.git"

	appDirName    = "mrtn"
	oracleDirName = "oracle"
)

// DataDir returns the designated OS-specific directory for persistent mrtn application data.
// Users can override this location with the MRTN_DATA_DIR environment variable.
func DataDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("MRTN_DATA_DIR")); override != "" {
		return filepath.Clean(override), nil
	}

	switch runtime.GOOS {
	case "windows":
		if localAppData := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); localAppData != "" {
			return filepath.Join(localAppData, appDirName), nil
		}
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return "", fmt.Errorf("não foi possível localizar o diretório de dados do usuário: %w", err)
		}
		return filepath.Join(cacheDir, appDirName), nil

	case "darwin":
		configDir, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("não foi possível localizar o diretório de dados do usuário: %w", err)
		}
		return filepath.Join(configDir, appDirName), nil

	default:
		if xdgData := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); xdgData != "" {
			return filepath.Join(xdgData, appDirName), nil
		}
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("não foi possível localizar o diretório home do usuário: %w", err)
		}
		return filepath.Join(homeDir, ".local", "share", appDirName), nil
	}
}

// OracleDir returns the directory path where the mrtn-oracle repository is stored locally.
func OracleDir() (string, error) {
	base, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, oracleDirName), nil
}
