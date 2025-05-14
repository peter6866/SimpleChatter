package logic

import (
	"context"

	"github.com/jinzhu/copier"
	"github.com/peter6866/SimpleChatter/apps/im/immodels"
	"github.com/peter6866/SimpleChatter/apps/im/rpc/im"
	"github.com/peter6866/SimpleChatter/apps/im/rpc/internal/svc"
	"github.com/peter6866/SimpleChatter/pkg/xerr"
	"github.com/pkg/errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get Conversations
func (l *GetConversationsLogic) GetConversations(in *im.GetConversationsReq) (*im.GetConversationsResp, error) {
	// find conversations by user id
	data, err := l.svcCtx.ConversationsModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		if err == immodels.ErrNotFound {
			return &im.GetConversationsResp{}, nil
		}
		return nil, errors.Wrapf(xerr.NewDBErr(), "ConversationsModel.FindByUserId err %v, req %v", err, in.UserId)
	}
	var res im.GetConversationsResp
	copier.Copy(&res, &data)

	// find conversation by conversation id
	ids := make([]string, 0, len(data.ConversationList))
	for _, conversation := range data.ConversationList {
		ids = append(ids, conversation.ConversationId)
	}
	conversations, err := l.svcCtx.ConversationModel.ListByConversationIds(l.ctx, ids)
	if err != nil {
		return nil, errors.Wrapf(xerr.NewDBErr(), "ConversationModel.ListByConversationIds err %v, req %v", err, ids)
	}

	// calculate unread message
	for _, conversation := range conversations {
		if _, ok := res.ConversationList[conversation.ConversationId]; !ok {
			continue
		}
		// total number of unread messages
		total := res.ConversationList[conversation.ConversationId].Total
		if total < int32(conversation.Total) {

			res.ConversationList[conversation.ConversationId].Total = int32(conversation.Total)
			res.ConversationList[conversation.ConversationId].ToRead = int32(conversation.Total) - total
			res.ConversationList[conversation.ConversationId].IsShow = true
		}
	}

	return &res, nil
}
