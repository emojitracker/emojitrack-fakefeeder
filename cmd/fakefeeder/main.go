package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	fakefeeder "github.com/emojitracker/emojitrack-fakefeeder"
)

var (
	targetURL = flag.String("target", "redis://localhost:6379", "URI for redis target")
	rate      = flag.Uint("rate", 250, "number of updates per second to generate")
	weighted  = flag.Bool("weight", true, "weight random update probability based on history")
	verbose   = flag.Bool("v", false, "verbose log all feeder updates")
)

func main() {
	flag.Parse()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	logger.Printf("Starting up. Target: %v (rate: %d/sec.)\n", *targetURL, *rate)

	// paranoid safety check: refuse to go anywhere near a production DB
	if strings.Contains(*targetURL, "rediscloud") {
		logger.Fatal("Are you certain you aren't trying to hit a prod db?")
	}

	// otherwise, set up the redis pool
	pool, err := targetPool(*targetURL)
	if err != nil {
		logger.Fatal(err)
	}

	// don't forget to seed random!
	rand.Seed(time.Now().UnixNano())

	// set up feeder with initial state in redis
	logger.Println("Setting up initial feeder state...")
	feeder, err := fakefeeder.NewFeeder(pool, fakefeeder.Snapshot(), *weighted)
	if err != nil {
		logger.Fatal(err)
	}
	if *verbose {
		feeder.VerboseLogger = logger
	}

	// start feeding redis random updates
	period := time.Second / time.Duration(*rate)
	logger.Printf("Sending fake updates every %v (%v/sec)", period, *rate)
	errChan := feeder.Start(context.Background(), period)
	for err := range errChan {
		logger.Println("ERROR:", err)
	}
}
