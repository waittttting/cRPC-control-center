package server

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/waittttting/cRPC-common/tcp"
)

type redisOp int

const (
	redisOpSAddServerIp = 1
	redisOpSRemServerIp = 2
)

func (ccs *ControlCenterServer)redisOp(message *tcp.Message, conn *tcp.Connection, op redisOp) error {

	client := redis.NewClient(&redis.Options{
		Addr:     ccs.config.Redis.Host,
		Password: ccs.config.Redis.Pwd,
		DB:       ccs.config.Redis.Index,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s:%s", message.Header.ServerName, message.Header.ServerVersion)

	switch op {
	case redisOpSAddServerIp:
		logrus.Infof("register info: key: %v, ip: %v", key, conn.IP)
		_, err = client.SAdd(key, conn.IP).Result()
	case redisOpSRemServerIp:
		logrus.Infof("delete info: key: %v, ip: %v", key, conn.IP)
		_, err = client.SRem(key, conn.IP).Result()
	}
	if err != nil {
		return err
	}
	logrus.Infof("redis client set, key: %v, value: %v", key, client.SMembers(key))
	client.Close()
	return nil
}
