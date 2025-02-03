package parser

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"

	"golang.org/x/xerrors"

	"github.com/dmitriy-dokshin/storypark_crawler/constdef"
	"github.com/dmitriy-dokshin/storypark_crawler/httputil"
)

type Story struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Media []*Media `json:"media"`
}

type Media struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	ContentType string `json:"content_type"`
	OriginalUrl string `json:"original_url"`
	ResizedUrl  string `json:"resized_url"`
}

type StoryResponse struct {
	Story *Story `json:"story"`
}

func LoadStory(
	ctx context.Context,
	logger *slog.Logger,
	client *http.Client,
	id string,
) (*Story, error) {
	exp := regexp.MustCompile("stories/[0-9]+")
	reqStr := exp.ReplaceAllString(constdef.RequestStory, fmt.Sprintf("stories/%v", id))
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

	storyJson, err := io.ReadAll(reader)
	if err != nil {
		return nil, xerrors.Errorf("unable to read story json: %w", err)
	}
	logger.DebugContext(ctx, "story json read", "json", string(storyJson))

	var storyResponse *StoryResponse
	err = json.Unmarshal(storyJson, &storyResponse)
	if err != nil {
		return nil, xerrors.Errorf("unable to unmarshal story json: %w", err)
	}
	return storyResponse.Story, nil
}
