package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

type LLMProvider struct {
	LLM string `json:"model"`
	APIKey string `json:"api_key"`
}

type Config struct {
	Default string `json:"default"`
	LLMProviders []LLMProvider `json:"models"`
}

func Save(LLMConfig LLMProvider) error {
	
	cfg := Config{
		LLMConfig.LLM,
		[]LLMProvider{LLMConfig},
	}


	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	isConfigExists := checkConfig(configPath)
	if !isConfigExists {
		err := createConfigFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
	}


	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error", err)
		log.Fatal(err)
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			fmt.Println("unmarshal error", err)
			log.Fatal(err)
		}
	}
	
	
	updated := false
	for i, p := range cfg.LLMProviders {
		if p.LLM == LLMConfig.LLM {
			cfg.LLMProviders[i] = LLMConfig
			updated = true
			break
		}
	}

	if !updated {
		cfg.LLMProviders = append(cfg.LLMProviders, LLMConfig)
	}

	cfg.Default = LLMConfig.LLM 

	file, err := os.OpenFile(configPath, os.O_CREATE| os.O_WRONLY| os.O_TRUNC, 0600)
	if err != nil {
		fmt.Println("error opening file")
	}
	defer file.Close()

	data, err = json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0600)
}


func checkConfig(configPath string) bool {

	_,err := os.Stat(configPath)
	if err != nil ||os.IsNotExist(err)  {
		return false
	}

	return true
}


func createConfigFile(configPath string) error {

	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return err
	}

	return nil

}

func getConfigPath() (string, error) {

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


func DefaultLLMKey() (*LLMProvider, error) {

	var cfg Config
	var useModel LLMProvider
	
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	isConfigExists := checkConfig(configPath)
	if !isConfigExists {
		return nil, errors.New("config file Not exists")
	}

	
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error", err)
		return nil, err
	}


	if len(data) > 0 {
		err = json.Unmarshal(data, &cfg)
		if err != nil {
			fmt.Println("unmarshal error", err)
			return nil, err
		}
	} else {
		return nil, errors.New("config file is empty")
	}

	defaultLLM := cfg.Default

	for i, p := range cfg.LLMProviders {
		if p.LLM == defaultLLM {
			useModel= cfg.LLMProviders[i]
			return &useModel, nil
		}
	}
	return nil, errors.New("not found default model in config")
}