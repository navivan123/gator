package config

import (
    "encoding/json"
    "os"
)

func (cfg *Config) SetUser(username string) error {
    cfg.CurrentUserName = username
    err := write(*cfg)
    return err
}

func write(cfg Config) error {
    configPath, err := getConfigFilePath()
    if err != nil {
        return err
    }

    data, err := json.Marshal(cfg)
    if err != nil {
        return err
    }

    err = os.WriteFile(configPath, data, 0600)
    return err
}
