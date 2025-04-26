package logic

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"github.com/peter6866/SimpleChatter/apps/user/models"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

var ErrUserNotFound = errors.New("user not found")

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *user.GetUserInfoReq) (*user.GetUserInfoResp, error) {
	// todo: add your logic here and delete this line

	userEntiy, err := l.svcCtx.UsersModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	var resp user.UserEntity
	copier.Copy(&resp, userEntiy)

	return &user.GetUserInfoResp{
		User: &resp,
	}, nil
}
