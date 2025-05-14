package logic

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/peter6866/SimpleChatter/apps/user/models"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"
	"github.com/peter6866/SimpleChatter/pkg/ctxdata"
	"github.com/peter6866/SimpleChatter/pkg/encrypt"
	"github.com/peter6866/SimpleChatter/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrPhoneNotRegistered = xerr.New(xerr.SERVER_COMMON_ERROR, "phone not registered")
	ErrUserPwdError       = xerr.New(xerr.SERVER_COMMON_ERROR, "user password error")
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user.LoginReq) (*user.LoginResp, error) {
	// Check if the user already exists
	userEntity, err := l.svcCtx.UsersModel.FindByPhone(l.ctx, in.Phone)
	if err != nil {
		if err == models.ErrNotFound {
			return nil, errors.WithStack(ErrPhoneNotRegistered)
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "find user by phone failed, err: %v, req %v", err, in.Phone)
	}

	if !encrypt.ValidatePasswordHash(in.Password, userEntity.Password.String) {
		return nil, errors.WithStack(ErrUserPwdError)
	}

	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		now,
		now+l.svcCtx.Config.Jwt.AccessExpire,
		userEntity.Id,
	)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "get jwt token failed, err: %v", err)
	}

	return &user.LoginResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
