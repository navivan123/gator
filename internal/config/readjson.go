package config

import (
    "os"
    "errors"
    "encoding/json"
//    "fmt"
)

func Read() (Config, error) {

    configPath, err := getConfigFilePath()
    if err != nil {
        return Config{}, err
    }

    data, err := os.ReadFile(configPath)
    if errors.Is(err, os.ErrNotExist) {
        return Config{}, err
    }

    config := Config{}
    err = json.Unmarshal(data, &config)
    if err != nil {
        return Config{}, err
    }

    return config, nil
}

func getConfigFilePath() (string, error) {
    homePath, err := os.UserHomeDir()
    if err != nil {
        return "", err
    }
    return homePath + "/" + configFileName, nil
}
