package redis

import (
	"Go-web-server/setting"
	"fmt"
	"github.com/go-redis/redis"
)

var Rdb *redis.Client

// Init 初始化链接
func Init(config *setting.RedisConfig) error{
	Rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d",config.Host,config.Post),
		Password: config.Password,
		DB: config.Db,
		PoolSize: config.PoolSize,
	})
	_, err := Rdb.Ping().Result()
	return err
}