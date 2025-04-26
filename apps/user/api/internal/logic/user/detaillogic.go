package user

import (
	"context"

	"github.com/jinzhu/copier"
	"github.com/peter6866/SimpleChatter/apps/user/api/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/user/api/internal/types"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"
	"github.com/peter6866/SimpleChatter/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type DetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get User info
func NewDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DetailLogic {
	return &DetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DetailLogic) Detail(req *types.UserInfoReq) (resp *types.UserInfoResp, err error) {
	uid := ctxdata.GetUId(l.ctx)

	userInfoResp, err := l.svcCtx.User.GetUserInfo(l.ctx, &user.GetUserInfoReq{
		Id: uid,
	})
	if err != nil {
		return nil, err
	}

	var res types.User
	copier.Copy(&res, userInfoResp.User)

	return &types.UserInfoResp{Info: res}, nil
}
