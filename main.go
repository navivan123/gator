package main

import _ "github.com/lib/pq"

import (
    "internal/config"
    "fmt"
    "os"
    "internal/database"
    "database/sql"
    "time"
)


func main() {

    // time program
    start := time.Now()
    defer tt(start)

    // Make CLI prettier by separating from prompt lines
    fmt.Println()

    // Read config from file containing current user and db url
    cfg, err := config.Read()
    if err != nil {
        return
    }

    // Create instance of commands struct
    coms := commands{ commandList: map[string]func(*state, command) error {} }

    // Register a handler function for each command
    coms.register("login",    handlerLogin)
    coms.register("register", handlerRegister)
    coms.register("reset",    handlerReset)
    coms.register("users",    handlerUsers)
    coms.register("agg",      handlerAgg)
    coms.register("addfeed",  handlerAddFeed)
    coms.register("feeds",    handlerFeedsWithName)

    // Open connection to database
    db, err := sql.Open("postgres", cfg.DBUrl)
    if err != nil {
        return
    }
    dbQueries := database.New(db)

    // Initialize state, for storing config and queries to be used by commands
    cState := state{ cfgState: &cfg, dbState: dbQueries }
    
    // Get args
    args := os.Args
    if len(args) < 2 {
        fmt.Println("Not enough arguments provided.  Exiting...")
        return
    }

    // Create command
    cName := args[1]
    cArgs := args[2:]
    com   := command{ name: cName, args: cArgs}

    // Run command
    err = coms.run(&cState, com)

    // Retrun with error if command fails
    if err != nil {
        fmt.Printf("Error while running command. Reason: %v\n", err)
        return
    }

}

func tt(start time.Time) {
    e := time.Since(start)
    fmt.Printf("Program took %v\n\n", e)
}
