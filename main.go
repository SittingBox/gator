package main

import (
	"database/sql"

	"fmt"

	"os"

	"github.com/SittingBox/blog_agg/internal/config"

	"github.com/SittingBox/blog_agg/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db         *database.Queries
	configFile *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	newState := &state{
		db:         dbQueries,
		configFile: &cfg,
	}

	cmd := &commands{}
	cmd.register("login", handlerLogin)
	cmd.register("register", handlerRegister)
	cmd.register("reset", handlerReset)
	cmd.register("users", handlerListUsers)
	cmd.register("feeds", handlerListFeeds)
	cmd.register("agg", handlerAgg)
	cmd.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmd.register("follow", middlewareLoggedIn(handlerFollow))
	cmd.register("following", middlewareLoggedIn(handlerFollowingFeeds))
	cmd.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmd.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		fmt.Println("No command provided")
		os.Exit(1)
	}

	err = cmd.run(newState, command{
		name:      os.Args[1],
		arguments: os.Args[2:],
	})
	if err != nil {
		fmt.Println("Error running command:", err)
		os.Exit(1)
	}

}
