## Requirements: Go and Postgres installed

### Installation
Install Gator with:
(For `go install`, look at your `go.mod`: its module name is the start of the install path.)
```
go install <your-module-path>
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

With that all complete you should be able to use the following commands:

Register a user:
``` gator register <name> ```

Add an RSS feed:
``` gator addfeed <url> ```

Aggregate feeds every 30 seconds (This will continue to run until you use ctrl + c to cancel the command):
``` gator agg 30s ```

Browse posts:
```gator browse [limit] ```

