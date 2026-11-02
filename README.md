## Requirements: Go and Postgres installed

#### A command-line RSS feed aggregator built with Go and Postgres.

### Installation

Install Gator with:
```
go install github.com/SittingBox/blog_agg@latest
```
You will want to create a file named

.gatorconfig.json

in your home directory ```~/.gatorconfig.json```

In the file you will put:
```
{
  "db_url": "postgres://username:password@localhost:5432/database?sslmode=disable"
}
```
Inside the .gatorconfig.json file, but you need to replace the ```username:password``` with your own postgres username and password

## Database Schema

Create your Postgres database, then run the SQL files in `sql/schema/` in numerical order:

```
psql -d users.sql -f sql/schema/001_users.sql
psql -d feeds.sql -f sql/schema/002_feeds.sql
psql -d feeds_follow.sql -f sql/schema/003_feed_follows.sql
psql -d feeds_last_fetched.sql -f sql/schema/004_feed_lastfetched.sql
psql -d create_posts.sql -f sql/schema/005_posts.sql
```

### Commands

With that all complete you should be able to use the following commands:

Register a user:
``` gator register <name> ```

Add an RSS feed:
``` gator addfeed <url> ```

Aggregate feeds every 30 seconds (This will continue to run until you use ctrl + c to cancel the command):
``` gator agg 30s ```

Browse posts:
```gator browse [limit] ```

