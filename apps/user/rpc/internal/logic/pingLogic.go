package logic

import (
	"context"

	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PingLogic) Ping(in *user.Request) (*user.Response, error) {
	return &user.Response{
		Pong: "Pong",
	}, nil
}
