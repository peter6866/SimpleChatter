package group

import (
	"net/http"

	"github.com/peter6866/SimpleChatter/apps/social/api/internal/logic/group"
	"github.com/peter6866/SimpleChatter/apps/social/api/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/social/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// apply to join group
func GroupPutInHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GroupPutInRep
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := group.NewGroupPutInLogic(r.Context(), svcCtx)
		resp, err := l.GroupPutIn(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
