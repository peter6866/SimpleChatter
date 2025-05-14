package handler

import (
	"net/http"

	"github.com/peter6866/SimpleChatter/apps/im/api/internal/logic"
	"github.com/peter6866/SimpleChatter/apps/im/api/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/im/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Get chat log by user
func getChatLogHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChatLogReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetChatLogLogic(r.Context(), svcCtx)
		resp, err := l.GetChatLog(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
