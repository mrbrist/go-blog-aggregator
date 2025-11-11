package config

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
)

const configFileName = "/.gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func SetUser(cfg *Config, username string) error {
	cfg.CurrentUserName = username

	err1 := write(cfg)
	if err1 != nil {
		return err1
	}

	return nil
}

func Read() (*Config, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		return nil, err
	}

	var data Config
	if err := json.Unmarshal(bs, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + configFileName, nil
}

func write(cfg *Config) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
