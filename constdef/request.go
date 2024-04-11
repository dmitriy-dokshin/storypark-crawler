package constdef

import (
	_ "embed"
)

//go:embed request_activity.txt
var RequestActivity string

//go:embed request_story.txt
var RequestStory string
