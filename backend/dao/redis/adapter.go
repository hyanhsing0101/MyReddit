package redis

import goredis "github.com/go-redis/redis"

type Client = goredis.Client
type Options = goredis.Options
type Z = goredis.Z

func NewClient(opt *Options) *Client {
	return goredis.NewClient(opt)
}

func IsNil(err error) bool {
	return err == goredis.Nil
}
