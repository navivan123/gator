package main

import (
    "internal/config"
    "internal/database"
)

type command struct {
    name   string
    args []string
}

type commands struct {
    commandList map[string]func(*state, command) error
}

type state struct {
    cfgState *config.Config
    dbState  *database.Queries
}

