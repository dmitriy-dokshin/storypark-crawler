package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dmitriy-dokshin/storypark_crawler/parser"
)

func main() {
	ctx := context.Background()
	client := new(http.Client)
	activities, err := parser.LoadActivities(ctx, client, time.Now())
	if err != nil {
		panic(err)
	}

	for _, activity := range activities {
		story, err := parser.LoadStory(ctx, client, activity.PostID)
		if err != nil {
			panic(err)
		}
		storyJson, err := json.Marshal(story)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(storyJson))
	}

}
