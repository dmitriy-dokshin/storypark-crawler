package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"golang.org/x/xerrors"

	"github.com/dmitriy-dokshin/storypark_crawler/parser"
)

func main() {
	ctx := context.Background()
	logger := slog.Default()
	err := run(ctx, logger)
	if err != nil {
		logger.InfoContext(ctx, "run failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	client := new(http.Client)

	//from := time.Now().Add(-10 * 365 * 24 * time.Hour)
	from, err := time.Parse(time.RFC3339, "2024-12-20T11:51:07+08:00")
	if err != nil {
		return xerrors.Errorf("failed to parse from time: %w", err)
	}
	//until := time.Now()
	until, err := time.Parse(time.RFC3339, "2025-02-03T13:33:41.000+08:00")
	if err != nil {
		return xerrors.Errorf("failed to parse until time: %w", err)
	}
	for {
		logger.InfoContext(ctx, "loading activities", "until", until)
		activities, err := parser.LoadActivities(ctx, logger, client, until)
		if err != nil {
			return xerrors.Errorf("failed to load activities: %w", err)
		}
		if len(activities) == 0 {
			logger.InfoContext(ctx, "no more activities found")
			return nil
		}

		for _, activity := range activities {
			until = activity.UpdatedAt
			logger := logger.With("post_id", activity.PostID, "until", until)

			if until.Compare(from) <= 0 {
				logger.InfoContext(ctx, "specified from time reached", "from", from)
				return nil
			}

			logger.InfoContext(ctx, "loading story")
			story, err := parser.LoadStory(ctx, logger, client, activity.PostID)
			if err != nil {
				return xerrors.Errorf("failed to load story: %w", err)
			}

			if len(story.Media) == 0 {
				logger.InfoContext(ctx, "no media in the story")
				continue
			}

			title := story.Title
			if title == "" {
				title = "(no title)"
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return xerrors.Errorf("failed to get user home dir: %w", err)
			}
			storyPath := fmt.Sprintf("%s/Downloads/stories/%s - %s - %s", homeDir, activity.UpdatedAt.Format(time.RFC3339), story.ID, title)
			err = os.Mkdir(storyPath, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				return xerrors.Errorf("failed to mkdir %s: %w", storyPath, err)
			}

			for _, media := range story.Media {
				logger := logger.With("media_id", media.ID)
				logger.InfoContext(ctx, "loading media")
				err := parser.LoadMedia(ctx, logger, client, media, storyPath)
				if err != nil {
					return xerrors.Errorf("failed to load media %q: %w", media.OriginalUrl, err)
				}
			}
		}
	}
}
