package session

import (
	"github.com/MetaDataLab/web3-console-backend/internal/pkg/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
)

var store *redis.Store

func Init() error {

	// 初始化基于redis的存储引擎
	// 参数说明：
	//    第1个参数 - redis最大的空闲连接数
	//    第2个参数 - 数通信协议tcp或者udp
	//    第3个参数 - redis地址, 格式，host:port
	//    第4个参数 - redis密码
	//    第5个参数 - session加密密钥
	s, err := redis.NewStore(10, "tcp", config.GConf.RedisConfig.Address, config.GConf.RedisConfig.Password, []byte("secret"))
	if err != nil {
		return err
	}
	s.Options(sessions.Options{
		MaxAge: config.GConf.RedisConfig.MaxAge, // 设置过期时间
	})

	store = &s
	return nil
}

func GetStore() sessions.Store {
	return *store
}
