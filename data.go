package fakefeeder

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/icrowley/fake"
)

//go:generate go run ./scripts/generate_snapshot.go -- snapshot_data.go

// Ranking defines the score for a single emoji glyph from the Emojitracker API.
type Ranking struct {
	Char  string `json:"char"`
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

// a semi realistic scale of where tweet IDs currently are at
var globalTweetID = 769706198425825280

// EnsmallenedTweet matches the structure used by emojitrack-feeder for sending
// out bandwidth efficient tweets.
type EnsmallenedTweet struct {
	ID              string    `json:"id"`
	Text            string    `json:"text"`
	ScreenName      string    `json:"screen_name"`
	Name            string    `json:"name"`
	Links           []string  `json:"links"`
	ProfileImageURL string    `json:"profile_image_url"`
	CreatedAt       time.Time `json:"created_at"`
}

// MustEncode returns the marshalled JSON representation of t, or panics if it
// cannot be encoded for some reason.
func (t *EnsmallenedTweet) MustEncode() []byte {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return b
}

func randomTweetForEmoji(r Ranking) EnsmallenedTweet {
	globalTweetID += 42

	return EnsmallenedTweet{
		ID:              strconv.Itoa(globalTweetID),
		Text:            fmt.Sprintf("%s %s", fake.Sentence(), r.Char),
		ScreenName:      fake.UserName(),
		Name:            fake.FullName(),
		Links:           []string{},
		ProfileImageURL: "https://abs.twimg.com/sticky/default_profile_images/default_profile_normal.png",
		CreatedAt:       time.Now(),
	}
}
