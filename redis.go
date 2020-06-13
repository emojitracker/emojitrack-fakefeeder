package main

import (
	"errors"
	"net/url"
	"time"

	"github.com/gomodule/redigo/redis"
)

// parse target redis URL to extract host and password (redigo doesn't natively
// understand URI format), use that for instantiating newPool().
func targetPool(targetURL string) (*redis.Pool, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, errors.New("could not parse target URL")
	}
	host := url.Host
	password := ""
	if url.User != nil {
		password, _ = url.User.Password()
	}
	return newPool(host, password), nil
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			// log.Println("REDISPOOL: Dialing a new Redis connection...")
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
		// TestOnBorrow: func(c redis.Conn, t time.Time) error {
		// 	log.Println("REDISPOOL: testing on borrow")
		// 	_, err := c.Do("PING")
		// 	return err
		// },
	}
}
