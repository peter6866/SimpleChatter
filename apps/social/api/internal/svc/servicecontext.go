package svc

import (
	"github.com/peter6866/SimpleChatter/apps/im/rpc/imclient"
	"github.com/peter6866/SimpleChatter/apps/social/api/internal/config"
	"github.com/peter6866/SimpleChatter/apps/social/rpc/socialclient"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/userClient"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	userClient.User
	socialclient.Social
	imclient.Im
	*redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		User:   userClient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
		Im:     imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
		Redis:  redis.MustNewRedis(c.Redisx),
	}
}
