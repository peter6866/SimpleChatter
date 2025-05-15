package msgTransfer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"sync"
	"time"

	"github.com/peter6866/SimpleChatter/apps/im/ws/ws"
	"github.com/peter6866/SimpleChatter/apps/task/mq/internal/svc"
	"github.com/peter6866/SimpleChatter/apps/task/mq/mq"
	"github.com/peter6866/SimpleChatter/pkg/bitmap"
	"github.com/peter6866/SimpleChatter/pkg/constants"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

// GroupMsgReadRecordDelayTime defines the delay time for processing group message read records
var GroupMsgReadRecordDelayTime = time.Second

// GroupMsgReadRecordDelayCount defines the threshold count for batch processing group message read records
var GroupMsgReadRecordDelayCount = 10

const (
	// GroupMsgReadHandlerAtTransfer indicates immediate processing of group message read records
	GroupMsgReadHandlerAtTransfer = iota
	// GroupMsgReadHandlerDelayTransfer indicates delayed processing of group message read records
	GroupMsgReadHandlerDelayTransfer
)

// MsgReadTransfer handles the transfer of message read status between services
type MsgReadTransfer struct {
	*baseMsgTransfer

	cache.Cache

	mu sync.Mutex

	// groupMsgs stores the group message read records for batch processing
	// key: conversationId
	groupMsgs map[string]*groupMsgRead
	// push channel for sending read status updates
	push chan *ws.Push
}

// NewMsgReadTransfer creates a new instance of MsgReadTransfer
func NewMsgReadTransfer(svc *svc.ServiceContext) kq.ConsumeHandler {
	m := &MsgReadTransfer{
		baseMsgTransfer: NewBaseMsgTransfer(svc),
		groupMsgs:       make(map[string]*groupMsgRead, 1),
		push:            make(chan *ws.Push, 1),
	}

	// Configure delay settings if not using immediate transfer
	if svc.Config.MsgReadHandler.GroupMsgReadHandler != GroupMsgReadHandlerAtTransfer {
		if svc.Config.MsgReadHandler.GroupMsgReadRecordDelayCount > 0 {
			GroupMsgReadRecordDelayCount = svc.Config.MsgReadHandler.GroupMsgReadRecordDelayCount
		}

		if svc.Config.MsgReadHandler.GroupMsgReadRecordDelayTime > 0 {
			GroupMsgReadRecordDelayTime = time.Duration(svc.Config.MsgReadHandler.GroupMsgReadRecordDelayTime) * time.Second
		}
	}

	go m.transfer()

	return m
}

// Consume processes incoming message read status updates
func (m *MsgReadTransfer) Consume(ctx context.Context, key, value string) error {
	m.Info("MsgReadTransfer ", value)

	var (
		data mq.MsgMarkRead
	)
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return err
	}

	// Update read status in chat logs
	readRecords, err := m.UpdateChatLogRead(ctx, &data)
	if err != nil {
		return err
	}

	// Prepare push notification
	push := &ws.Push{
		ConversationId: data.ConversationId,
		ChatType:       data.ChatType,
		SendId:         data.SendId,
		RecvId:         data.RecvId,
		ContentType:    constants.ContentMakeRead,
		ReadRecords:    readRecords,
	}

	switch data.ChatType {
	case constants.SingleChatType:
		// Direct push for single chat messages
		m.push <- push
	case constants.GroupChatType:
		// Check if immediate processing is enabled
		if m.svcCtx.Config.MsgReadHandler.GroupMsgReadHandler == GroupMsgReadHandlerAtTransfer {
			m.push <- push
		}

		m.mu.Lock()
		defer m.mu.Unlock()

		push.SendId = ""

		if _, ok := m.groupMsgs[push.ConversationId]; ok {
			m.Infof("Merging push for conversation %v", push.ConversationId)
			// Merge read status updates
			m.groupMsgs[push.ConversationId].mergePush(push)
		} else {
			m.Infof("Creating new group message read record for conversation %v", push.ConversationId)
			m.groupMsgs[push.ConversationId] = newGroupMsgRead(push, m.push)
		}
	}

	return nil
}

// UpdateChatLogRead updates the read status of chat logs
func (m *MsgReadTransfer) UpdateChatLogRead(ctx context.Context, data *mq.MsgMarkRead) (map[string]string, error) {
	res := make(map[string]string)

	chatLogs, err := m.svcCtx.ChatLogModel.ListByMsgIds(ctx, data.MsgIds)
	if err != nil {
		return nil, err
	}

	// Process read status updates
	for _, chatLog := range chatLogs {
		switch chatLog.ChatType {
		case constants.SingleChatType:
			chatLog.ReadRecords = []byte{1}
		case constants.GroupChatType:
			readRecords := bitmap.Load(chatLog.ReadRecords)
			readRecords.Set(data.SendId)
			chatLog.ReadRecords = readRecords.Export()
		}

		res[chatLog.ID.Hex()] = base64.StdEncoding.EncodeToString(chatLog.ReadRecords)

		err = m.svcCtx.ChatLogModel.UpdateMakeRead(ctx, chatLog.ID, chatLog.ReadRecords)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

// transfer processes and forwards read status updates
func (m *MsgReadTransfer) transfer() {
	for push := range m.push {
		if push.RecvId != "" || len(push.RecvIds) > 0 {
			if err := m.Transfer(context.Background(), push); err != nil {
				m.Errorf("Transfer error: %v for push: %v", err, push)
			}
		}

		if push.ChatType == constants.SingleChatType {
			continue
		}

		if m.svcCtx.Config.MsgReadHandler.GroupMsgReadHandler == GroupMsgReadHandlerAtTransfer {
			continue
		}

		// Clean up processed data
		m.mu.Lock()
		if _, ok := m.groupMsgs[push.ConversationId]; ok && m.groupMsgs[push.ConversationId].IsIdle() {
			m.groupMsgs[push.ConversationId].clear()
			delete(m.groupMsgs, push.ConversationId)
		}
		m.mu.Unlock()
	}
}
