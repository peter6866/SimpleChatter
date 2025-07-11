// Code generated by goctl. DO NOT EDIT.
// goctl 1.8.2

package handler

import (
	"net/http"

	"github.com/peter6866/SimpleChatter/apps/im/api/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// Get chat log by user
				Method:  http.MethodGet,
				Path:    "/chatlog",
				Handler: getChatLogHandler(serverCtx),
			},
			{
				// Get chat log read records
				Method:  http.MethodGet,
				Path:    "/chatlog/readRecords",
				Handler: getChatLogReadRecordsHandler(serverCtx),
			},
			{
				// Get conversation
				Method:  http.MethodGet,
				Path:    "/conversation",
				Handler: getConversationsHandler(serverCtx),
			},
			{
				// Update conversation
				Method:  http.MethodPut,
				Path:    "/conversation",
				Handler: putConversationsHandler(serverCtx),
			},
			{
				// Set up conversation
				Method:  http.MethodPost,
				Path:    "/setup/conversation",
				Handler: setUpUserConversationHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.JwtAuth.AccessSecret),
		rest.WithPrefix("/v1/im"),
	)
}
