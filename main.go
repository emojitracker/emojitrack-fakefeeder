package main

import (
	"flag"
	"log"
	"strings"
	"time"

	"github.com/mroth/weightedrand"
)

var (
	targetURL = flag.String("target", "redis://localhost:6379", "URI for redis target")
	rate      = flag.Int("rate", 250, "number of updates per second to generate")
	weighted  = flag.Bool("weight", true, "weight random update probability based on history")
	verbose   = flag.Bool("v", false, "verbose log all updates to stdout")
)

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

	// set up initial state in redis
	log.Println("Setting up initial state...")
	// reset counts to equal snapshot
	err := seedScores(c)
	if err != nil {
		log.Fatal("FATAL: could not seed initial scores: ", err)
	}
	// mock historic tweet sliding window for each emoji
	err = seedTweets(c)
	if err != nil {
		log.Fatal("FATAL: could not seed initial tweets: ", err)
	}

	// create weighted random choice data
	var emojiChoices []weightedrand.Choice
	for _, er := range emojiRankings {
		ec := weightedrand.Choice{Item: er, Weight: uint(er.score)}
		emojiChoices = append(emojiChoices, ec)
	}
	emojiChooser := weightedrand.NewChooser(emojiChoices...)

	// start feeding redis random updates
	period := time.Second / time.Duration(*rate)
	log.Println("Now starting to send fake updates every", period)
	_ = updateScript.Load(c)
	for {
		time.Sleep(period)

		var e emojiRanking
		if *weighted {
			e = emojiChooser.Pick().(emojiRanking)
		} else {
			e = *randomEmoji()
		}

		t := randomUpdateForEmoji(e)
		updateScript.SendHash(c, e.id, t.MustEncode())
		c.Flush()

		if *verbose {
			log.Println("Sent fake update for", e)
		}
	}
}
