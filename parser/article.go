package parser

import (
	"compress/gzip"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/xerrors"

	"github.com/dmitriy-dokshin/storypark_crawler/constdef"
	"github.com/dmitriy-dokshin/storypark_crawler/httputil"
)

type Article struct {
	PostID    string    `json:"post_id"`
	UpdatedAt time.Time `json:"updated_at"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
}

func LoadActivities(
	ctx context.Context,
	logger *slog.Logger,
	client *http.Client,
	until time.Time,
) ([]*Article, error) {
	exp := regexp.MustCompile("until=[0-9]+")
	reqStr := exp.ReplaceAllString(constdef.RequestActivity, fmt.Sprintf("until=%v", until.Unix()))
	req, err := httputil.ParseRequest(ctx, reqStr, nil)
	if err != nil {
		return nil, xerrors.Errorf("unable to parse request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, xerrors.Errorf("unable to do request: %w", err)
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, xerrors.Errorf("unable to create gzip reader: %w", err)
	}
	defer func(reader *gzip.Reader) {
		err := reader.Close()
		if err != nil {
			logger.ErrorContext(ctx, "unable to close gzip reader", "error", err)
		}
	}(reader)

	root, err := htmlquery.Parse(reader)
	if err != nil {
		return nil, xerrors.Errorf("unable to parse html: %w", err)
	}

	var articles []*Article
	articleNodes := htmlquery.Find(root, "//article")
	for _, articleNode := range articleNodes {
		postID := htmlquery.SelectAttr(articleNode, "data-post-id")
		if postID == "" {
			continue
		}

		updatedAtStr := htmlquery.SelectAttr(articleNode, "data-updated-at")
		if updatedAtStr == "" {
			continue
		}
		updatedAtSeconds, err := strconv.ParseInt(updatedAtStr, 10, 64)
		if err != nil {
			return nil, xerrors.Errorf("unable to parse updated at: %w", err)
		}
		updatedAt := time.Unix(updatedAtSeconds, 0)

		articleType := htmlquery.SelectAttr(articleNode, "data-type")

		titleNode := htmlquery.FindOne(articleNode, "//h1")
		var title string
		if titleNode != nil {
			title = titleNode.FirstChild.Data
		}

		articles = append(articles, &Article{
			PostID:    postID,
			UpdatedAt: updatedAt,
			Type:      articleType,
			Title:     title,
		})
	}
	return articles, nil
}
