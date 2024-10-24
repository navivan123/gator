# gator
## Requirements
### Go
You need Go 1.23.2 installed
### Postgres
You need Postgres 16+ installed
### Config
- A file ".gatorconfig.json" must be present in the home directory (ex: ~/.gatorconfig.json).
- This file must contain json equivalent to this: { "db_url": <Database URL string>, "current_user_name": <username> } \
  For initialization purposes, db_url needs to be set. current_user_name will be set after you log in for the first time.

## Install
Use go install github.com/navivan123/gator to install the gator command.

## Run
- gator help
  - lists help menu of below commands
- gator register <username>
  - register username on the app
- gator login <username>
  - login on the app with username
- gator users
  - lists all users on app. indicates which one is currently logged in
- gator addfeed <name> <url>
  - Add feed with name and url to database and automatically subscribes the user
- gator follow <url>
  - Finds feed with url argument to subscribe user to the feed
- gator unfollow <url>
  - Finds feed with url argument to unsubscribe user to the feed
- gator following
  - Lists feeds that user is currently subscribed to
- gator feeds
  - Lists all feeds names, urls, and users that created them
- gator agg <time>
  - Aggregates feeds from base rss links that user is currently subscribed to and stores them in the database
  - Indicate the amount of time between link fetches as time.  (Ex: gator agg 1m, gator agg 5s, gator agg 1h)
- gator browse <limit>
  - Browse Aggregate feeds that user collected with the agg command
  - By default returns 2.  Optionally use a number indicating how many feeds you would like to receive
