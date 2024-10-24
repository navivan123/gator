package main

import (
    "errors"
    "fmt"
    "internal/database"
    "github.com/google/uuid"
    "time"
    "context"
    "log"
    "strings"
    "database/sql"
    "strconv"
)

var ErrorParsingTime = errors.New("Error: Unable to parse time from argument")
var ErrorParsingInt  = errors.New("Error: Unable to parse int from argument")

var ErrorSettingUser = errors.New("Error: User unable to be set")

var EmptyArgList  = errors.New("Error: No argument for command that takes arguments")
var NotEnoughArgs = errors.New("Error: Not enough arguments for command that takes multiple arguments")

var ErrorRunningHandle = errors.New("Error: Unable to run command")
var NoCommandExists    = errors.New("Error: Unable to find command")

var ErrorRegisterUser    = errors.New("Error: Failure registering user in user table")
var ErrorGettingUser     = errors.New("Error: Failure to get user from user table (user probably does not exist)")
var ErrorGettingUsers    = errors.New("Error: Failure to get all users from user table")
var ErrorGettingUsername = errors.New("Error: Failure to get username based on uuid from user table")

var ErrorFetchingFeed  = errors.New("Error: Failure to fetch feed from web")
var ErrorCreatingFeed  = errors.New("Error: Failure to create feed in feed table")
var ErrorGettingFeeds  = errors.New("Error: Failure to get all feeds from feed table")
var ErrorGettingFeed   = errors.New("Error: Failure to get feed from feed table")

var ErrorGettingNextFeed      = errors.New("Error: Failure to get next feed from feed table")
var ErrorMarkingFeedAsFetched = errors.New("Error: Failure to mark feed as fetched")

var ErrorCreatingFeedFollows    = errors.New("Error: Failure to create feed follow table")
var ErrorGettingUserFeedFollows = errors.New("Error: Failure to get feed follows from follow table using username")

var ErrorGettingPosts = errors.New("Error: Failure to get posts")

var ErrorDeletingFeeds        = errors.New("Error: Failure to truncate feeds table")
var ErrorDeletingUsers        = errors.New("Error: Failure to truncate users table")
var ErrorDeletingFeedFollows  = errors.New("Error: Failure to truncate feed follows table")

// Handlers
func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) < 1 {
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
    if len(cmd.args) < 1 {
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

    err = s.dbState.DeleteFeedFollows(context.Background())
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorDeletingFeedFollows, err)
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
    if len(cmd.args) < 1 {
        fmt.Println("usage: agg <time_between_requests>")
        return EmptyArgList
    }

    timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorParsingTime, err)
    }

    fmt.Println("Collecting feeds every", cmd.args[0])

    ticker := time.NewTicker(timeBetweenRequests)
    for ; ; <-ticker.C {
        feed, err := s.dbState.GetNextFeedToFetch(context.Background())
        if err != nil {
            log.Printf("%v | Reason: %v\n", ErrorGettingNextFeed, err)
            continue
        }

        _, err = s.dbState.MarkFeedFetched(context.Background(), feed.ID)
        if err != nil {
            return fmt.Errorf("%v | Reason: %v", ErrorMarkingFeedAsFetched, err)
        }

        rss, err := fetchFeed(context.Background(), feed.Url)
        if err != nil {
            return fmt.Errorf("%v | Reason: %v", ErrorFetchingFeed, err)
        }

        for _, item := range rss.Channel.Item {
            publishedAt := sql.NullTime{}
            if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
                publishedAt = sql.NullTime{ Time:  t, Valid: true, }
            }
            
            _, err = s.dbState.CreatePost(context.Background(), database.CreatePostParams{ ID:    uuid.New(), CreatedAt:   time.Now(), UpdatedAt:   time.Now(), FeedID: feed.ID,
                                                                                           Title: item.Title, Url:         item.Link,  PublishedAt: publishedAt,
                                                                                           Description: sql.NullString{ String: item.Description, Valid:  true, }, })
            if err != nil {
                if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
                    continue
                }
                log.Printf("Couldn't create post: %v", err)
                continue
            }
        }
        log.Printf("Feed %s collected, %v posts found", feed.Name, len(rss.Channel.Item))
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

    _, err = s.dbState.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ ID:     uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), 
                                                                                               UserID: user.ID,    FeedID:    feed.ID, })
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorCreatingFeedFollows, err)
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

func handlerFollow(s *state, cmd command) error {
    if len(cmd.args) < 1 {
        fmt.Println("usage: follow <url>")
        return NotEnoughArgs
    }

    user, err := s.dbState.GetUser(context.Background(), s.cfgState.CurrentUserName)
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
    }

    
    feed, err := s.dbState.GetFeedUrl(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingFeed, err)
    }

    feedFollow, err := s.dbState.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ ID:     uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), 
                                                                                                         UserID: user.ID,    FeedID:    feed.ID, })
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorCreatingFeedFollows, err)
    }

    fmt.Println("FeedFollow created successfully:")
    fmt.Println("Feed Name:", feedFollow.FeedName)
    fmt.Println("User:", s.cfgState.CurrentUserName)

    fmt.Println()
    fmt.Println("=====================================")

    return nil
}

func handlerFollowing(s *state, cmd command) error {
    feedFollows, err := s.dbState.GetFeedFollowsForUser(context.Background(), s.cfgState.CurrentUserName)
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUserFeedFollows, err)
    }

    fmt.Printf("FeedFollows for user %v received successfully:\n", s.cfgState.CurrentUserName)
    for _, feedFollow := range feedFollows {
        fmt.Println("Name:", feedFollow.FeedName)
    }

    fmt.Println()
    fmt.Println("=====================================")
    
    return nil
}

func handlerUnfollow(s *state, cmd command) error {
    if len(cmd.args) < 1 {
        fmt.Println("usage: unfollow <url>")
        return NotEnoughArgs
    }

    user, err := s.dbState.GetUser(context.Background(), s.cfgState.CurrentUserName)
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
    }

    feed, err := s.dbState.GetFeedUrl(context.Background(), cmd.args[0])
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingFeed, err)
    }

    err = s.dbState.DeleteFeedFollowsForUserUrl(context.Background(), database.DeleteFeedFollowsForUserUrlParams{ UserID: user.ID, FeedID: feed.ID })
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUserFeedFollows, err)
    }

    return nil
}

func handlerBrowse(s *state, cmd command) error {
    limit := 2
    var err error
    if len(cmd.args) > 0 {
        limit, err = strconv.Atoi(cmd.args[0])
        if err != nil {
            return fmt.Errorf("%v | Reason: %v", ErrorParsingInt, err)
        }
    }

    user, err := s.dbState.GetUser(context.Background(), s.cfgState.CurrentUserName)
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
    }

    posts, err := s.dbState.GetPostsForUser(context.Background(), database.GetPostsForUserParams{ UserID: user.ID, Limit: int32(limit) })
    if err != nil {
        return fmt.Errorf("%v | Reason: %v", ErrorGettingPosts, err)
    }

    for _, post := range posts {
        fmt.Println("Title:        ", post.Title)
        fmt.Println("Feed Name:    ", post.FeedName)
        fmt.Println("Url:          ", post.Url)
        fmt.Println("Created at:   ", post.CreatedAt)
        fmt.Println("Updated at:   ", post.UpdatedAt)
        fmt.Println("Published at: ", post.PublishedAt)
        fmt.Println("Description:  ")
        fmt.Println(post.Description)
        fmt.Println()
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

// Middleware (eugh)
//func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
//    
//    user, err := s.dbState.GetUser(context.Background(), s.cfgState.CurrentUserName)
//    if err != nil {
//        return fmt.Errorf("%v | Reason: %v", ErrorGettingUser, err)
//   }
//  
//}
