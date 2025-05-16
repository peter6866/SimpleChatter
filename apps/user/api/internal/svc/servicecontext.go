package svc

import (
	"github.com/peter6866/SimpleChatter/apps/user/api/internal/config"
	userClient "github.com/peter6866/SimpleChatter/apps/user/rpc/userClient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	*redis.Redis
	userClient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		User:   userClient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Redis:  redis.MustNewRedis(c.Redisx),
	}
}
