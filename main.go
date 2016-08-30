package main // import "github.com/mroth/emojitrack-fakefeeder"

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/icrowley/fake"
)

var targetURL = flag.String("target", "redis://localhost:6379", "URI for redis target")
var rate = flag.Int("rate", 250, "number of updates per second to generate")
var verbose = flag.Bool("v", false, "verbose log all updates to stdout")

// Placeholder for snapshot score values
// var initialScores = map[string]int{}

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

func (t *EnsmallenedTweet) Encoded() []byte {
	b, _ := json.Marshal(t)
	return b
}

// pick a random ID for an existing emoji in counter
func randomEmoji() *emojiRanking {
	return &emojiRankings[rand.Intn(len(emojiRankings))]
}

func randomUpdateForEmoji(e emojiRanking) EnsmallenedTweet {
	globalTweetID += 42

	return EnsmallenedTweet{
		ID:              strconv.Itoa(globalTweetID),
		Text:            fmt.Sprintf("%s %s", fake.Sentence(), e.char),
		ScreenName:      fake.UserName(),
		Name:            fake.FullName(),
		Links:           []string{},
		ProfileImageURL: "https://pbs.twimg.com/profile_images/724054682579161088/3GgLeR65_normal.jpg",
		CreatedAt:       time.Now(),
	}
}

func main() {
	flag.Parse()

	// echo startup status
	log.Printf("Starting up. TARGET_URL: %v (rate: %d/sec.)\n", *targetURL, *rate)

	// paranoid safety check: refuse to go anywhere near a production DB
	if strings.Contains(*targetURL, "rediscloud") {
		log.Fatal("Are you certain you aren't trying to hit a prod db?")
	}

	// try to connect to Redis, otherwise die with error
	initRedis()
	c := redisPool.Get()
	defer c.Close()
	if _, err := c.Do("PING"); err != nil {
		log.Fatal("Redis connection failed: ", err)
	}

	// populate initial state internally
	// for _, er := range emojiRankings {
	// 	initialScores[er.id] = er.score
	// }

	// reset Redis main counter to that state
	log.Println("Setting up initial state...")
	for _, er := range emojiRankings {
		c.Do("ZADD", "emojitrack_score", er.score, er.id)
		// generate 10 random tweets for each existing ID
		t := randomUpdateForEmoji(er)
		updateScript.Do(c, er.id, t.Encoded())
	}

	// start feeding redis random updates
	period := time.Second / time.Duration(*rate)
	log.Println("Now starting to send fake updates every", period)
	for {
		time.Sleep(period)

		e := *randomEmoji()
		t := randomUpdateForEmoji(e)
		updateScript.Do(c, e.id, t.Encoded())

		if *verbose {
			log.Println("Sent fake update for ", e)
		}
	}
}
