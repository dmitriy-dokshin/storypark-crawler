package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/xerrors"

	"github.com/dmitriy-dokshin/storypark_crawler/constdef"
	"github.com/dmitriy-dokshin/storypark_crawler/httputil"
)

func RequestActivities(ctx context.Context, client *http.Client, until time.Time) (string, error) {
	exp := regexp.MustCompile("until=[0-9]+")
	reqStr := exp.ReplaceAllString(constdef.RequestActivity, fmt.Sprintf("until=%v", until.Unix()))
	req, err := httputil.ParseRequest(ctx, reqStr, nil)
	if err != nil {
		return "", xerrors.Errorf("unable to parse request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", xerrors.Errorf("unable to do request: %w", err)
	}

	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", xerrors.Errorf("unable to create gzip reader: %w", err)
	}

	res, err := io.ReadAll(reader)
	if err != nil {
		return "", xerrors.Errorf("unable to read compressed body: %w", err)
	}
	return string(res), nil
}

func main() {
	client := new(http.Client)
	activities, err := RequestActivities(context.Background(), client, time.Now())
	if err != nil {
		panic(err)
	}
	fmt.Println(activities)
}
