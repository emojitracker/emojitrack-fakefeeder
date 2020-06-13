package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/mroth/weightedrand"
)

// Feeder will generate probable random data to a redis instance to emulate the
// behavior of emojitrack-feeder without needing an super high-speed internet
// connection, Twitter Platform Partner status, and gobs of processing power.
//
// It is primarily useful for feeding data to emulate realtime behavior hacking
// on the other components of Emojitracker within a docker network.
type Feeder struct {
	rp            *redis.Pool
	seed          []emojiRanking
	chooseFunc    func() emojiRanking
	VerboseLogger logger // override to enable verbose update logging
}

type logger interface {
	Printf(string, ...interface{})
}

// NewFeeder generates a Feeder utilizing a configured redis.Pool p, using seed
// data. The weight parameter determines whether the updates it generates will
// be probablistically weighted based on past scores rather than uniform random
// distribution.
//
// Once NewFeeder is initialized, it will automatically seed the initial data,
// but note it will not start sending realtime updates until Start() is invoked
// or manual updates are sent via the Update() command.
//
// NewFeeder will return an error if it was unable to properly seed the redis
// instance in any fashion.
func NewFeeder(p *redis.Pool, seed []emojiRanking, weight bool) (*Feeder, error) {
	f := Feeder{
		rp:         p,
		seed:       seed,
		chooseFunc: buildChooseFunc(seed, weight),
	}

	err := f.init()
	return &f, err
}

func (f *Feeder) init() error {
	c := f.rp.Get()
	defer c.Close()

	err := f.seedScores(c)
	if err != nil {
		return fmt.Errorf("could not seed initial scores: %w", err)
	}
	err = f.seedTweets(c)
	if err != nil {
		return fmt.Errorf("could not seed initial tweets: %w", err)
	}
	err = updateScript.Load(c)
	if err != nil {
		return fmt.Errorf("could not load Lua update script: %w", err)
	}

	return nil
}

// constants of redis commands to avoid potential runtime errors from typos :-)
const (
	rEXEC  = "EXEC"
	rLPUSH = "LPUSH"
	rMULTI = "MULTI"
	rZADD  = "ZADD"

	rScoreKey       = "emojitrack_score"
	rTweetKeyPrefix = "emojitrack_tweets_"
)

// seedScores sets all scores in redis.Conn c to match snapshot data.
func (f *Feeder) seedScores(c redis.Conn) error {
	c.Send(rMULTI)
	for _, r := range f.seed {
		err := c.Send(rZADD, rScoreKey, r.score, r.id)
		if err != nil {
			return err
		}
	}
	_, err := c.Do(rEXEC)
	return err
}

// seedTweets generate 10 initial random tweets for each existing emoji, such
// that initial buffer for historical window is filled.
func (f *Feeder) seedTweets(c redis.Conn) error {
	c.Send(rMULTI)
	for _, r := range f.seed {
		for i := 0; i < 10; i++ {
			t := randomTweetForEmoji(r)
			tKey := rTweetKeyPrefix + r.id
			err := c.Send(rLPUSH, tKey, t.MustEncode())
			if err != nil {
				return err
			}
		}
	}
	_, err := c.Do(rEXEC)
	return err
}

func buildChooseFunc(seed []emojiRanking, weighted bool) func() emojiRanking {
	// non-weighted, simple random choice
	if !weighted {
		cf := func() emojiRanking {
			return seed[rand.Intn(len(seed))]
		}
		return cf
	}

	// weighted random distribution (using github.com/mroth/weightedrand)
	choices := make([]weightedrand.Choice, 0, len(seed))
	for _, r := range seed {
		c := weightedrand.Choice{Item: r, Weight: uint(r.score)}
		choices = append(choices, c)
	}
	chooser := weightedrand.NewChooser(choices...)
	cf := func() emojiRanking {
		return chooser.Pick().(emojiRanking)
	}
	return cf
}

// Update sends a single random update to the configured redis instance.
func (f *Feeder) Update() error {
	c := f.rp.Get()
	defer c.Close()

	emoji := f.chooseFunc()
	tweet := randomTweetForEmoji(emoji)
	payload := tweet.MustEncode()
	if err := updateScript.SendHash(c, emoji.id, payload); err != nil {
		return err
	}
	c.Flush()
	if f.VerboseLogger != nil {
		f.VerboseLogger.Printf("sent fake update for %v", emoji)
	}
	_, err := c.Receive() // blocks for response
	return err
}

// Start begins a background goroutine which calls f.Update() every
// time.Duration d. If the provided context is cancelled for any reason, it will
// safely cleanup and exit.
//
// The returned error chan will report any errors occuring during update. It has
// a minor buffer to allow a small grace period for consumption, but note that
// by design it will drop new errors on the floor if they are not being
// consumed, in order to allow the error chan to be safely ignored (e.g. no
// manual draining needed) without blocking updates.
func (f *Feeder) Start(ctx context.Context, d time.Duration) <-chan error {
	errC := make(chan error, 8)
	go func() {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		defer close(errC)

		for {
			select {
			case <-ctx.Done():
				errC <- ctx.Err()
				return
			case <-ticker.C:
				if err := f.Update(); err != nil {
					// In this use case, dropping error on the floor is the
					// desired behavior when the reader gets behind, because the
					// updates are periodic.
					select {
					case errC <- err:
					default:
					}
				}
			}
		}
	}()
	return errC
}

// This is the exact same update script used in emojitrack-feeder.
var updateScript = redis.NewScript(0, `
-- Updates the server whenever a new emoji is seen in a tweet
--
-- Putting this in a script enables us to save some bandwidth by not
-- transmitting any redundant data to the server, as we can calculate the
-- appropriate key names there and re-use data that goes to multiple
-- destinations.

local uid      = ARGV[1]   -- unified codepoint ID
local tinyjson = ARGV[2]   -- json blob representing the ensmallened tweet

-- increment the score in a sorted set
redis.call('ZINCRBY', 'emojitrack_score', 1, uid)

-- stream the fact that the score was updated
redis.call('PUBLISH', 'stream.score_updates', uid)

-- for each emoji char, store the most recent 10 tweets in a list
local tweet_details_key = "emojitrack_tweets_" .. uid
redis.call('LPUSH', tweet_details_key, tinyjson)
redis.call('LTRIM', tweet_details_key, 0, 9)

-- also stream all tweet updates to named streams by char
local stream_details_key = "stream.tweet_updates." .. uid
redis.call('PUBLISH', stream_details_key, tinyjson)

-- return ok status
return 1
`)
