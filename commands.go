package main

import (
    "errors"
    "fmt"
    "internal/database"
    "github.com/google/uuid"
    "time"
    "context"
)

var ErrorSettingUser   = errors.New("Error: User unable to be set")

var EmptyArgList       = errors.New("Error: No argument for command that takes arguments")
var NotEnoughArgs      = errors.New("Error: Not enough arguments for command that takes multiple arguments")

var ErrorRunningHandle = errors.New("Error: Unable to run command")
var NoCommandExists    = errors.New("Error: Unable to find command")

var ErrorRegisterUser    = errors.New("Error: Failure registering user in database")
var ErrorGettingUser     = errors.New("Error: Failure to get user from database (user probably does not exist)")
var ErrorGettingUsers    = errors.New("Error: Failure to get all users from database")
var ErrorGettingUsername = errors.New("Error: Failure to get username based on uuid from database")

var ErrorFetchingFeed  = errors.New("Error: Failure to fetch feed")
var ErrorCreatingFeed  = errors.New("Error: Failure to create feed")
var ErrorGettingFeeds  = errors.New("Error: Failure to get all feeds from database")

var ErrorDeletingFeeds  = errors.New("Error: Failure to truncate feeds table")
var ErrorDeletingUsers  = errors.New("Error: Failure to truncate users table")


// Handlers
func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 {
        fmt.Println("usage: login <name>")
        return EmptyArgList
    }
    _, err := s.dbState.GetUser(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
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
        fmt.Println("usage: register <name>")
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
    err := s.dbState.DeleteFeeds(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorDeletingFeeds, err)
    }

    err = s.dbState.DeleteUsers(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorDeletingUsers, err)
    }
    return nil
}

func handlerUsers(s *state, cmd command) error {
    users, err := s.dbState.GetUsers(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUsers, err)
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

func handlerAgg(s *state, cmd command) error {
    
    rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorFetchingFeed, err)
    }
    
    // Print RSS "Header" Data
    fmt.Println("Title:",       rss.Channel.Title)
    fmt.Println("Description:", rss.Channel.Description)
    fmt.Println("Link:",        rss.Channel.Link)

    // Print RSS Item Data
    for _, item := range rss.Channel.Item {
        fmt.Println("Title:",        item.Title)
        fmt.Println("Description:",  item.Description)
        fmt.Println("Link:",         item.Link)
        fmt.Println("Publish Date:", item.PubDate)
    }

    return nil
}

func handlerAddFeed(s *state, cmd command) error {
    if len(cmd.args) < 2 {
        fmt.Println("usage: addfeed <name> <url>")
        return NotEnoughArgs
    }

    user, err := s.dbState.GetUser(context.Background(), s.cfgState.CurrentUserName)
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
    }

    feed, err := s.dbState.CreateFeed(context.Background(), database.CreateFeedParams{ ID:   uuid.New(),  CreatedAt: time.Now(),  UpdatedAt: time.Now(), 
                                                                                   Name: cmd.args[0], Url:       cmd.args[1], UserID:    user.ID, })
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorCreatingFeed, err)
    }

    fmt.Println("Feed created successfully:")
    printFeed(feed)
    fmt.Println()
    fmt.Println("=====================================")

    return nil
    
}

func handlerFeeds(s *state, cmd command) error {
    feeds, err := s.dbState.GetFeeds(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingFeeds, err)
    }

    for _, feed := range feeds {
        fmt.Printf("Name: %v | URL: %v | Username: ", feed.Name, feed.Url)
        name, err := s.dbState.GetUserName(context.Background(), feed.UserID)
        if err != nil {
            return fmt.Errorf("%v | Reason: %v", ErrorGettingUsername, err)
        }
        fmt.Printf("%v\n", name)
    }

    return nil
}

func handlerFeedsWithName(s *state, cmd command) error {
    feeds, err := s.dbState.GetFeedsWithName(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingFeeds, err)
    }

    for _, feed := range feeds {
        fmt.Printf("Name: %v | URL: %v | Username: %v\n", feed.Name, feed.Url, feed.Username)
    }

    return nil
}

// Methods
func (c *commands) register(name string, f func(*state, command) error) {
    c.commandList[name] = f 
}

func (c *commands) run(s *state, cmd command) error {

    if cmd.name == "help" {
        fmt.Println("Command List: ")
        for name, _ := range c.commandList {
            fmt.Println("*", name)
        }
        return nil
    }

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

// Helpers
func printFeed(feed database.Feed) {
    fmt.Printf("* ID:            %s\n", feed.ID)
    fmt.Printf("* Created:       %v\n", feed.CreatedAt)
    fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
    fmt.Printf("* Name:          %s\n", feed.Name)
    fmt.Printf("* URL:           %s\n", feed.Url)
    fmt.Printf("* UserID:        %s\n", feed.UserID)
}
