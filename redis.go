package main

import (
	"log"
	"net/url"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	redisPool *redis.Pool
)

func initRedis() {
	url, err := url.Parse(*targetURL)
	if err != nil {
		log.Fatal("Could not understand TARGET_URL paramer. Dying.")
	}
	host := url.Host
	password := ""
	if url.User != nil {
		password, _ = url.User.Password()
	}
	redisPool = newPool(host, password)
}

// yay for standard boilerplate, sigh
func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			log.Println("REDISPOOL: Dialing a new Redis connection...")
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			log.Println("REDISPOOL: testing on borrow")
			_, err := c.Do("PING")
			return err
		},
	}
}

// This is the same update script used in emojitrack-feeder
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
