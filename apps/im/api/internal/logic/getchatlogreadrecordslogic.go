package logic

import (
	"context"

	"github.com/peter6866/SimpleChatter/apps/im/api/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/im/api/internal/types"
	"github.com/peter6866/SimpleChatter/apps/im/rpc/im"
	"github.com/peter6866/SimpleChatter/apps/social/rpc/socialclient"
	"github.com/peter6866/SimpleChatter/apps/user/rpc/user"
	"github.com/peter6866/SimpleChatter/pkg/bitmap"
	"github.com/peter6866/SimpleChatter/pkg/constants"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatLogReadRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// Get chat log read records
func NewGetChatLogReadRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatLogReadRecordsLogic {
	return &GetChatLogReadRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChatLogReadRecordsLogic) GetChatLogReadRecords(req *types.GetChatLogReadRecordsReq) (resp *types.GetChatLogReadRecordsResp, err error) {
	chatlogs, err := l.svcCtx.Im.GetChatLog(l.ctx, &im.GetChatLogReq{
		MsgId: req.MsgId,
	})

	if err != nil || len(chatlogs.List) == 0 {
		return nil, err
	}

	var (
		chatlog = chatlogs.List[0]
		reads   = []string{chatlog.SendId}
		unreads []string
		ids     []string
	)

	switch constants.ChatType(chatlog.ChatType) {
	case constants.SingleChatType:
		if len(chatlog.ReadRecords) == 0 || chatlog.ReadRecords[0] == 0 {
			unreads = []string{chatlog.RecvId}
		} else {
			reads = append(reads, chatlog.RecvId)
		}
		ids = []string{chatlog.RecvId, chatlog.SendId}
	case constants.GroupChatType:
		groupUsers, err := l.svcCtx.Social.GroupUsers(l.ctx, &socialclient.GroupUsersReq{
			GroupId: chatlog.RecvId,
		})
		if err != nil {
			return nil, err
		}

		bitmaps := bitmap.Load(chatlog.ReadRecords)
		for _, members := range groupUsers.List {
			ids = append(ids, members.UserId)

			if members.UserId == chatlog.SendId {
				continue
			}

			if bitmaps.IsSet(members.UserId) {
				reads = append(reads, members.UserId)
			} else {
				unreads = append(unreads, members.UserId)
			}
		}
	}

	userEntitys, err := l.svcCtx.User.FindUser(l.ctx, &user.FindUserReq{
		Ids: ids,
	})
	if err != nil {
		return nil, err
	}
	userEntitySet := make(map[string]*user.UserEntity, len(userEntitys.User))
	for i, entity := range userEntitys.User {
		userEntitySet[entity.Id] = userEntitys.User[i]
	}

	for i, read := range reads {
		if u := userEntitySet[read]; u != nil {
			reads[i] = u.Phone
		}
	}
	for i, unread := range unreads {
		if u := userEntitySet[unread]; u != nil {
			unreads[i] = u.Phone
		}
	}

	return &types.GetChatLogReadRecordsResp{
		Reads:   reads,
		UnReads: unreads,
	}, nil
}
