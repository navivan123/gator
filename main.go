package main

import _ "github.com/lib/pq"

import (
    "internal/config"
//  "fmt"
    "log"
    "os"
    "internal/database"
    "database/sql"
)


func main() {

    cfg, err := config.Read()
    if err != nil {
        return
    }


    // Create instance of commands struct
    coms := commands{ commandList: map[string]func(*state, command) error {} }

    // Register a handler function for login
    coms.register("login", handlerLogin)
    coms.register("register", handlerRegister)
    coms.register("reset", handlerReset)
    coms.register("users", handlerUsers)

    db, err := sql.Open("postgres", cfg.DBUrl)
    if err != nil {
        return
    }

    dbQueries := database.New(db)

    cState := state{ cfgState: &cfg, dbState: dbQueries }
    
    // Get args
    args := os.Args
    if len(args) < 2 {
        log.Fatalf("Not enough arguments provided.  Exiting...")
    }

    // Create command
    cName := args[1]
    cArgs := args[2:]
    com   := command{ name: cName, args: cArgs}

    // Run command
    err = coms.run(&cState, com)

    if err != nil {
        log.Fatalf("Error while running command. Reason: %v", err)
    }

}
