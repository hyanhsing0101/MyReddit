package redis

import (
	"fmt"
	"myreddit/settings"
)

var rdb *Client

func Init(cfg *settings.RedisConfig) (err error) {
	rdb = NewClient(&Options{
		Addr: fmt.Sprintf("%s:%d",
			cfg.Host,
			cfg.Port,
		),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	_, err = rdb.Ping().Result()
	return
}

func Close() {
	_ = rdb.Close()
}
