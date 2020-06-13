package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/icrowley/fake"
)

//go:generate go run ./scripts/generate_snapshot.go -- snapshot_data.go

type emojiRanking struct {
	char, id, name string
	score          int
}

// a semi realistic scale of where tweet IDs currently are at
var globalTweetID = 769706198425825280

// pick a random ID for an existing emoji in counter
func randomEmoji() *emojiRanking {
	return &emojiRankings[rand.Intn(len(emojiRankings))]
}

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

func randomTweetForEmoji(e emojiRanking) EnsmallenedTweet {
	globalTweetID += 42

	return EnsmallenedTweet{
		ID:              strconv.Itoa(globalTweetID),
		Text:            fmt.Sprintf("%s %s", fake.Sentence(), e.char),
		ScreenName:      fake.UserName(),
		Name:            fake.FullName(),
		Links:           []string{},
		ProfileImageURL: "https://abs.twimg.com/sticky/default_profile_images/default_profile_normal.png",
		CreatedAt:       time.Now(),
	}
}
