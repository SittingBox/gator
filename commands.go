package main

import (
	"context"

	"strconv"

	"errors"

	"database/sql"

	"time"

	"fmt"

	"github.com/SittingBox/blog_agg/internal/database"

	"github.com/lib/pq"

	"github.com/google/uuid"
)

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

/*func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("username argument is required")
	}

	userLogin, err := s.db.GetUser(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("user not found in database: %w", err)
	}

	err = s.configFile.SetUser(userLogin.Name)
	if err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	fmt.Printf("User set to: %s\n", userLogin.Name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("username argument is required")
	}

	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
	})
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}

	err = handlerLogin(s, command{name: "login", arguments: []string{newUser.Name}})
	if err != nil {
		return fmt.Errorf("failed to log in user: %w", err)
	}

	fmt.Printf("User created and logged in: %s\n", newUser.Name)
	fmt.Printf("User data: ID= %s, CreatedAt= %s, UpdatedAt= %s, Name= %s\n", newUser.ID, newUser.CreatedAt, newUser.UpdatedAt, newUser.Name)

	return nil
}*/

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset users: %w", err)
	}

	err = s.db.ResetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset feeds: %w", err)
	}

	fmt.Println("All tables have been reset.")
	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to retrieve users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	fmt.Println("Registered Users:")
	for _, user := range users {
		if user.Name != s.configFile.CurrentUserName {
			fmt.Printf("* %s\n", user.Name)
		} else {
			fmt.Printf("* %s (current)\n", user.Name)
		}
	}
	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to retrieve feeds: %w", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	fmt.Println("Registered Feeds:")
	for _, feed := range feeds {
		feedCreator, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("failed to retrieve feed creator: %w", err)
		}
		fmt.Printf("* Name: %s, URL: %s, UserID: %s\n", feed.Name, feed.Url, feedCreator.Name)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("time between requests argument is required")
	}

	timeBetweenReqs, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("failed to parse time between requests: %w", err)
	}

	fmt.Printf("Fetching feed every %s...\n", timeBetweenReqs)

	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		handlerScrapFeeds(s)
	}

	return nil
}

func handlerScrapFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		fmt.Printf("failed to retrieve feed: %s", err)
		return
	}

	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		fmt.Printf("failed to mark feed as fetched: %s", err)
		return
	}

	Feed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		fmt.Printf("failed to fetch feed: %s", err)
		return
	}

	for _, item := range Feed.Channel.Item {

		var pqErr *pq.Error

		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		publishedAt := sql.NullTime{Time: pubDate, Valid: err == nil}

		descrip := sql.NullString{String: item.Description, Valid: item.Description != ""}

		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: descrip,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				fmt.Printf("Post already exists: %s\n", item.Link)
			} else {
				fmt.Printf("failed to create post: %s\n", err)
			}
			continue
		}
		if err != nil {
			fmt.Printf("failed to create post: %s\n", err)
			continue
		}
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int

	if len(cmd.arguments) < 1 {
		limit = 2
	} else {
		parsed, err := strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = parsed
	}

	posts, err := s.db.GetPostForUser(context.Background(), database.GetPostForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to retrieve posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found.")
		return nil
	}

	fmt.Println("Posts:")

	for _, post := range posts {
		fmt.Printf("- %s\n", post.Title)
		fmt.Printf("  URL: %s\n", post.Url)
		if post.Description.Valid {
			fmt.Printf("  Description: %s\n", post.Description.String)
		}
		if post.PublishedAt.Valid {
			fmt.Printf("  Published At: %s\n", post.PublishedAt.Time.Format(time.RFC1123))
		}
	}
	return nil
}

/*func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("feed name and URL arguments are required in that order")
	}

	currentUser, err := s.db.GetUser(context.Background(), s.configFile.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	feedName := cmd.arguments[0]
	feedUrl := cmd.arguments[1]

	newFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow the newly added feed: %w", err)
	}

	fmt.Printf("Feed added successfully: Name= %s, Url= %s, UserID= %s\n", newFeed.Name, newFeed.Url, newFeed.UserID)

	return nil
}

func handlerFollow(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("Feed URL argument is required")
	}

	feedUrl := cmd.arguments[0]

	feed, err := s.db.GetFeedFromURL(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	currentUser, err := s.db.GetUser(context.Background(), s.configFile.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	followFeed, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}

	fmt.Printf("Successfully followed feed:\n Feed name: %s, User name: %s\n", followFeed.FeedName, followFeed.UserName)

	return nil
}

func handlerFollowingFeeds(s *state, cmd command) error {
	currentUser, err := s.db.GetUser(context.Background(), s.configFile.CurrentUserName)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("failed to get following feeds: %w", err)
	}

	fmt.Println("Following Feeds:")
	for _, follow := range feedFollows {
		fmt.Printf("- Feed: %s", follow.FeedName)
	}

	return nil
}*/

func (c *commands) register(name string, handler func(*state, command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = handler
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}
	return handler(s, cmd)
}
