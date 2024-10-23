package main

import (
    "errors"
    "fmt"
    "internal/database"
    "github.com/google/uuid"
    "time"
    "context"
)

var EmptyArgList       = errors.New("Error: No argument for command that takes arguments")
var ErrorSettingUser   = errors.New("Error: User unable to be set")
var ErrorRunningHandle = errors.New("Error: Unable to run command")
var NoCommandExists    = errors.New("Error: Unable to find command")
var ErrorRegisterUser  = errors.New("Error: Failure registering user in database")

// Handlers
func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return EmptyArgList
    }
    _, err := s.dbState.GetUser(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("Error getting user from database (probably does not exist) | %v", err)
    }

    err = s.cfgState.SetUser(cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorSettingUser, err)
    }

    fmt.Printf("User %v has been set.\n", cmd.args[0])
    return nil
}

func handlerRegister(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        return EmptyArgList
    }

    user, err := s.dbState.CreateUser(context.Background(), database.CreateUserParams{ ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]})
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorRegisterUser, err)
    }
    err = s.cfgState.SetUser(cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorSettingUser, err)
    }

    fmt.Printf("User %v has been set.  ID: %v | CreatedAt: %v | UpdatedAt: %v | Name: %v\n", cmd.args[0], user.ID, user.CreatedAt, user.UpdatedAt, user.Name)
    return nil
    
}

func handlerReset(s *state, cmd command) error {
    err := s.dbState.Delete(context.Background())
    return err
}

func handlerUsers(s *state, cmd command) error {
    users, err := s.dbState.GetUsers(context.Background())
    if err != nil {
        return err
    }

    for _, user := range users {
        fmt.Printf("* %v", user)
        if user == s.cfgState.CurrentUserName {
            fmt.Printf(" (current)")
        }
        fmt.Printf("\n")
    }

    return nil
}

// Methods
func (c *commands) register(name string, f func(*state, command) error) {
    c.commandList[name] = f 
}

func (c *commands) run(s *state, cmd command) error {
    fmt.Println(cmd.name)
    comFunc, ok := c.commandList[cmd.name]
    if !ok {
        return NoCommandExists
    }

    err := comFunc(s, cmd)


    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorRunningHandle, err)
    }

    return nil
}
