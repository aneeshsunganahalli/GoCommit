package StoreUtils

import (
	"os"
	"path/filepath"
	"runtime"
)

func CheckConfig(configPath string) bool {

	_, err := os.Stat(configPath)
	if err != nil || os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateConfigFile(configPath string) error {

	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}

	return nil

}

func GetConfigPath() (string, error) {

	appName := "commit-msg"

	switch runtime.GOOS {

	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}

		return filepath.Join(localAppData, appName, "config.json"), nil

	case "darwin":

		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(home, "Library", "Application Support", appName, "config.json"), nil

	default:

		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}

			configHome = filepath.Join(home, ".config")
		}

		return filepath.Join(configHome, appName, "config.json"), nil
	}

}
