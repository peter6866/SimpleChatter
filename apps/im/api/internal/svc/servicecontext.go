package svc

import (
	"github.com/peter6866/SimpleChatter/apps/im/api/internal/config"
	"github.com/peter6866/SimpleChatter/apps/im/rpc/imclient"
	"github.com/peter6866/SimpleChatter/apps/social/rpc/socialclient"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/userClient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	imclient.Im
	userClient.User
	socialclient.Social
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,

		Im:     imclient.NewIm(zrpc.MustNewClient(c.ImRpc)),
		User:   userClient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		Social: socialclient.NewSocial(zrpc.MustNewClient(c.SocialRpc)),
	}
}
