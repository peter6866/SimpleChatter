package push

import (
	"github.com/mitchellh/mapstructure"
	"github.com/peter6866/SimpleChatter/apps/im/ws/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/im/ws/websocket"
	"github.com/peter6866/SimpleChatter/apps/im/ws/ws"
)

func Push(svc *svc.ServiceContext) websocket.HandlerFunc {
	return func(srv *websocket.Server, conn *websocket.Conn, msg *websocket.Message) {
		var data ws.Push
		if err := mapstructure.Decode(msg.Data, &data); err != nil {
			srv.Send(websocket.NewErrMessage(err))
			return
		}

		// target to send
		rconn := srv.GetConn(data.RecvId)
		if rconn == nil {
			// target offline
			return
		}

		srv.Infof("push msg %v", data)

		srv.Send(websocket.NewMessage(data.SendId, &ws.Chat{
			ConversationId: data.ConversationId,
			ChatType:       data.ChatType,
			SendTime:       data.SendTime,
			Msg: ws.Msg{
				MType:   data.MType,
				Content: data.Content,
			},
		}), rconn)
	}
}
