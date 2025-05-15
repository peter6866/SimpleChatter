package msgTransfer

import (
	"sync"
	"time"

	"github.com/peter6866/SimpleChatter/apps/im/ws/ws"
	"github.com/peter6866/SimpleChatter/pkg/constants"
	"github.com/zeromicro/go-zero/core/logx"
)

// groupMsgRead handles batch processing of group message read status updates
type groupMsgRead struct {
	mu             sync.Mutex
	conversationId string
	push           *ws.Push
	pushCh         chan *ws.Push
	count          int

	pushTime time.Time
	done     chan struct{}
}

// newGroupMsgRead creates a new instance of groupMsgRead for handling group message read status
func newGroupMsgRead(push *ws.Push, pushCh chan *ws.Push) *groupMsgRead {
	m := &groupMsgRead{
		conversationId: push.ConversationId,
		push:           push,
		pushCh:         pushCh,
		count:          1,
		pushTime:       time.Now(),
		done:           make(chan struct{}),
	}

	go m.transfer()
	return m
}

// mergePush combines multiple read status updates into a single batch
func (m *groupMsgRead) mergePush(push *ws.Push) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.count++
	for msgId, read := range push.ReadRecords {
		m.push.ReadRecords[msgId] = read
	}
}

// transfer processes and forwards batched read status updates
func (m *groupMsgRead) transfer() {
	timer := time.NewTimer(GroupMsgReadRecordDelayTime / 2)
	defer timer.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-timer.C:
			m.mu.Lock()

			pushTime := m.pushTime
			val := GroupMsgReadRecordDelayTime - time.Since(pushTime)
			push := m.push
			logx.Infof("Timer triggered at %v, remaining time: %v", time.Now(), val)

			// Continue waiting if conditions for batch processing are not met
			if val > 0 && m.count < GroupMsgReadRecordDelayCount || push == nil {
				if val > 0 {
					timer.Reset(val)
				}

				m.mu.Unlock()
				continue
			}

			// Reset state and forward the batched updates
			m.pushTime = time.Now()
			m.push = nil
			m.count = 0
			timer.Reset(GroupMsgReadRecordDelayTime / 2)
			m.mu.Unlock()

			logx.Infof("Forwarding batched read status updates: %v", push)
			m.pushCh <- push
		default:
			m.mu.Lock()

			// Process immediately if batch size threshold is reached
			if m.count >= GroupMsgReadRecordDelayCount {
				push := m.push
				m.push = nil
				m.count = 0
				m.mu.Unlock()

				logx.Infof("Forwarding read status updates due to batch size threshold: %v", push)
				m.pushCh <- push
				continue
			}

			// Handle idle state
			if m.isIdle() {
				m.mu.Unlock()
				m.pushCh <- &ws.Push{
					ChatType:       constants.GroupChatType,
					ConversationId: m.conversationId,
				}
				continue
			}
			m.mu.Unlock()

			// Implement backoff delay
			tempDelay := GroupMsgReadRecordDelayTime / 4
			if tempDelay > time.Second {
				tempDelay = time.Second
			}
			time.Sleep(tempDelay)
		}
	}
}

// IsIdle checks if the group message read processor is in an idle state
func (m *groupMsgRead) IsIdle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isIdle()
}

// isIdle determines if the processor is idle based on time and state conditions
func (m *groupMsgRead) isIdle() bool {
	pushTime := m.pushTime
	val := GroupMsgReadRecordDelayTime*2 - time.Since(pushTime)

	if val <= 0 && m.push == nil && m.count == 0 {
		return true
	}

	return false
}

// clear resets the group message read processor state
func (m *groupMsgRead) clear() {
	select {
	case <-m.done:
	default:
		close(m.done)
	}

	m.push = nil
}
