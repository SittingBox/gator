package main

import (
	"context"

	"fmt"

	"time"

	"github.com/google/uuid"

	"github.com/SittingBox/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		currentUser, err := s.db.GetUser(context.Background(), s.configFile.CurrentUserName)
		if err != nil {
			return fmt.Errorf("failed to get current user: %w", err)
		}

		return handler(s, cmd, currentUser)
	}
}

func handlerLogin(s *state, cmd command) error {
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
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("feed name and URL arguments are required in that order")
	}

	feedName := cmd.arguments[0]
	feedUrl := cmd.arguments[1]

	newFeed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to add feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow the newly added feed: %w", err)
	}

	fmt.Printf("Feed added successfully: Name= %s, Url= %s, UserID= %s\n", newFeed.Name, newFeed.Url, newFeed.UserID)

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("Feed URL argument is required")
	}

	feedUrl := cmd.arguments[0]

	feed, err := s.db.GetFeedFromURL(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	followFeed, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to follow feed: %w", err)
	}

	fmt.Printf("Successfully followed feed:\n Feed name: %s, User name: %s\n", followFeed.FeedName, followFeed.UserName)

	return nil
}

func handlerFollowingFeeds(s *state, cmd command, user database.User) error {

	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get following feeds: %w", err)
	}

	fmt.Println("Following Feeds: ")
	for _, follow := range feedFollows {
		fmt.Printf("- Feed: %s\n", follow.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("Feed URL argument is required")
	}

	FeedFromURL, err := s.db.GetFeedFromURL(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	err = s.db.RemoveFollowedFeed(context.Background(), database.RemoveFollowedFeedParams{
		UserID: user.ID,
		FeedID: FeedFromURL.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to unfollow feed: %w", err)
	}

	return nil
}
