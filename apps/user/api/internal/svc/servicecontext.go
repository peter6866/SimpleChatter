package svc

import (
	"github.com/peter6866/SimpleChatter/apps/user/api/internal/config"
	userclient "github.com/peter6866/SimpleChatter/apps/user/rpc/userClient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	userclient.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
		User:   userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
	}
}
