package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/peter6866/SimpleChatter/apps/user/models"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"
	"github.com/peter6866/SimpleChatter/pkg/ctxdata"
	"github.com/peter6866/SimpleChatter/pkg/encrypt"
	"github.com/peter6866/SimpleChatter/pkg/wuid"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	ErrPhoneIsRegistered = errors.New("phone is already registered")
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	// todo: add your logic here and delete this line

	// Check if the user already exists
	userEntity, err := l.svcCtx.UsersModel.FindByPhone(l.ctx, in.Phone)
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	if userEntity != nil {
		return nil, ErrPhoneIsRegistered
	}

	// Create a new user
	userEntity = &models.Users{
		Id:       wuid.GenUid(l.svcCtx.Config.Mysql.DataSource),
		Avatar:   in.Avatar,
		Nickname: in.Nickname,
		Phone:    in.Phone,
		Sex: sql.NullInt64{
			Int64: int64(in.Sex),
			Valid: true,
		},
	}

	if len(in.Password) > 0 {
		genPassword, err := encrypt.GenPasswordHash([]byte(in.Password))
		if err != nil {
			return nil, err
		}

		userEntity.Password = sql.NullString{
			String: string(genPassword),
			Valid:  true,
		}
	}

	_, err = l.svcCtx.UsersModel.Insert(l.ctx, userEntity)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	token, err := ctxdata.GetJwtToken(
		l.svcCtx.Config.Jwt.AccessSecret,
		now,
		now+l.svcCtx.Config.Jwt.AccessExpire,
		userEntity.Id,
	)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Token:  token,
		Expire: now + l.svcCtx.Config.Jwt.AccessExpire,
	}, nil
}
