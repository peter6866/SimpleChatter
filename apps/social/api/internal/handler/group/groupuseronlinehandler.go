package group

import (
	"net/http"

	"github.com/peter6866/SimpleChatter/apps/social/api/internal/logic/group"
	"github.com/peter6866/SimpleChatter/apps/social/api/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// group user online
func GroupUserOnlineHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupUserOnlineReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewGroupUserOnlineLogic(r.Context(), svcCtx)
		resp, err := l.GroupUserOnline(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
