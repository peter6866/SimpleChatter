package logic

import (
	"path/filepath"

	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/config"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
)

var svcCtx *svc.ServiceContext

func init() {
	var c config.Config
	conf.MustLoad(filepath.Join("../../etc/dev/user.yaml"), &c)
	svcCtx = svc.NewServiceContext(c)
}
