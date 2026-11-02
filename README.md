# Gator

## Requirements: Go and Postgres installed

#### A command-line RSS feed aggregator built with Go and Postgres.

### Installation

Install Gator with:
```sh
go install github.com/SittingBox/blog_agg@latest
```
You will want to create a file named

.gatorconfig.json

in your home directory ``` ~/.gatorconfig.json```

In the file you will put:
```
{
  "db_url": "postgres://username:password@localhost:5432/database?sslmode=disable"
}
```
Inside the .gatorconfig.json file, but you need to replace the ```username:password``` with your own postgres username and password

## Database Schema

- Your `psql -d ...` argument should be the **database name**, not a SQL filename. Use one database consistently:

```sh
psql -d <database-name> -f sql/schema/001_users.sql
psql -d <database-name> -f sql/schema/002_feeds.sql
psql -d <database-name> -f sql/schema/003_feed_follows.sql
psql -d <database-name> -f sql/schema/004_feed_lastfetched.sql
psql -d <database-name> -f sql/schema/005_posts.sql
```
Replace <database-name> with the Postgres database name you created. It should match the database name in your .gatorconfig.json db_url.

### Commands

With that all complete you should be able to use the following commands:

Register a user:
``` sh
gator register <name>
```
Add an RSS feed:
```sh
gator addfeed <url>
```

Aggregate feeds every 30 seconds (This will continue to run until you use ctrl + c to cancel the command):
```sh 
gator agg 30s
```

Browse posts:
```sh
gator browse [limit]
```

