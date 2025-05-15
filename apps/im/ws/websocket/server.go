package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

// AckType represents the acknowledgment type for message handling
type AckType int

const (
	// NoAck indicates no acknowledgment is required
	NoAck AckType = iota
	// OnlyAck indicates a simple acknowledgment is required
	OnlyAck
	// RigorAck indicates a rigorous acknowledgment with timeout and retry mechanism is required
	RigorAck
)

// ToString converts AckType to its string representation
func (t AckType) ToString() string {
	switch t {
	case OnlyAck:
		return "OnlyAck"
	case RigorAck:
		return "RigorAck"
	}

	return "NoAck"
}

// Server represents a WebSocket server instance
// It manages WebSocket connections, message routing, and user authentication
type Server struct {
	sync.RWMutex

	*threading.TaskRunner

	opt            *serverOption
	authentication Authentication

	routes map[string]HandlerFunc
	addr   string
	patten string

	connToUser map[*Conn]string
	userToConn map[string]*Conn

	upgrader websocket.Upgrader
	logx.Logger
}

// NewServer creates a new WebSocket server instance
// Parameters:
//   - addr: The address to listen on
//   - opts: Optional server configurations
//
// Returns:
//   - *Server: A new server instance
func NewServer(addr string, opts ...ServerOptions) *Server {
	opt := newServerOptions(opts...)

	return &Server{
		routes:   make(map[string]HandlerFunc),
		addr:     addr,
		patten:   opt.patten,
		opt:      &opt,
		upgrader: websocket.Upgrader{},

		authentication: opt.Authentication,

		connToUser: make(map[*Conn]string),
		userToConn: make(map[string]*Conn),

		Logger:     logx.WithContext(context.Background()),
		TaskRunner: threading.NewTaskRunner(opt.concurrency),
	}
}

// ServerWs handles incoming WebSocket connection requests
// It performs authentication and sets up the connection
func (s *Server) ServerWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			s.Errorf("server handler ws recover err %v", r)
		}
	}()

	conn := NewConn(s, w, r)
	if conn == nil {
		return
	}
	//conn, err := s.upgrader.Upgrade(w, r, nil)
	//if err != nil {
	//	s.Errorf("upgrade err %v", err)
	//	return
	//}

	if !s.authentication.Auth(w, r) {
		s.Send(&Message{FrameType: FrameData, Data: "not authorized"}, conn)
		conn.Close()
		return
	}

	// Register the new connection
	s.addConn(conn, r)

	go s.handlerConn(conn)
}

// handlerConn manages the lifecycle of a WebSocket connection
// It handles message reading, acknowledgment processing, and connection cleanup
func (s *Server) handlerConn(conn *Conn) {
	uids := s.GetUsers(conn)
	conn.Uid = uids[0]

	// Start message processing goroutines
	go s.handlerWrite(conn)

	if s.isAck(nil) {
		go s.readAck(conn)
	}

	for {
		// Read incoming messages
		_, msg, err := conn.ReadMessage()
		if err != nil {
			s.Errorf("websocket conn read message err %v", err)
			s.Close(conn)
			return
		}
		// Parse message
		var message Message
		if err = json.Unmarshal(msg, &message); err != nil {
			s.Errorf("json unmarshal err %v, msg %v", err, string(msg))
			s.Close(conn)
			return
		}

		// Process message based on acknowledgment requirements
		if s.isAck(&message) {
			s.Infof("conn message read ack msg %v", message)
			conn.appendMsgMq(&message)
		} else {
			conn.message <- &message
		}
	}
}

// isAck determines if a message requires acknowledgment
// Parameters:
//   - message: The message to check
//
// Returns:
//   - bool: True if acknowledgment is required
func (s *Server) isAck(message *Message) bool {
	if message == nil {
		return s.opt.ack != NoAck
	}
	return s.opt.ack != NoAck && message.FrameType != FrameNoAck
}

// readAck handles message acknowledgment processing
// It manages different acknowledgment strategies (OnlyAck and RigorAck)
func (s *Server) readAck(conn *Conn) {
	for {
		select {
		case <-conn.done:
			s.Infof("close message ack uid %v ", conn.Uid)
			return
		default:
		}

		// Process message queue
		conn.messageMu.Lock()
		if len(conn.readMessage) == 0 {
			conn.messageMu.Unlock()
			time.Sleep(100 * time.Microsecond)
			continue
		}

		message := conn.readMessage[0]

		// Handle different acknowledgment types
		switch s.opt.ack {
		case OnlyAck:
			// Simple acknowledgment: send ack and process message
			s.Send(&Message{
				FrameType: FrameAck,
				Id:        message.Id,
				AckSeq:    message.AckSeq + 1,
			}, conn)
			conn.readMessage = conn.readMessage[1:]
			conn.messageMu.Unlock()
			conn.message <- message

		case RigorAck:
			// Rigorous acknowledgment with timeout and retry
			if message.AckSeq == 0 {
				// Initial acknowledgment
				conn.readMessage[0].AckSeq++
				conn.readMessage[0].AckTime = time.Now()
				s.Send(&Message{
					FrameType: FrameAck,
					Id:        message.Id,
					AckSeq:    message.AckSeq,
				}, conn)
				s.Infof("message ack RigorAck send mid %v, seq %v , time%v", message.Id, message.AckSeq,
					message.AckTime)
				conn.messageMu.Unlock()
				continue
			}

			// Verify client acknowledgment
			msgSeq := conn.readMessageSeq[message.Id]
			if msgSeq.AckSeq > message.AckSeq {
				// Acknowledgment received
				conn.readMessage = conn.readMessage[1:]
				conn.messageMu.Unlock()
				conn.message <- message
				s.Infof("message ack RigorAck success mid %v", message.Id)
				continue
			}

			// Check acknowledgment timeout
			val := s.opt.ackTimeout - time.Since(message.AckTime)
			if !message.AckTime.IsZero() && val <= 0 {
				// Timeout reached, remove message
				delete(conn.readMessageSeq, message.Id)
				conn.readMessage = conn.readMessage[1:]
				conn.messageMu.Unlock()
				continue
			}
			// Retry acknowledgment
			conn.messageMu.Unlock()
			s.Send(&Message{
				FrameType: FrameAck,
				Id:        message.Id,
				AckSeq:    message.AckSeq,
			}, conn)
			time.Sleep(3 * time.Second)
		}
	}
}

// handlerWrite processes outgoing messages
// It handles different message types and routes them to appropriate handlers
func (s *Server) handlerWrite(conn *Conn) {
	for {
		select {
		case <-conn.done:
			return
		case message := <-conn.message:
			switch message.FrameType {
			case FramePing:
				s.Send(&Message{FrameType: FramePing}, conn)
			case FrameData:
				// Route message to appropriate handler
				if handler, ok := s.routes[message.Method]; ok {
					handler(s, conn, message)
				} else {
					s.Send(&Message{FrameType: FrameData, Data: fmt.Sprintf("Method %v does not exist, please check", message.Method)}, conn)
				}
			}

			if s.isAck(message) {
				conn.messageMu.Lock()
				delete(conn.readMessageSeq, message.Id)
				conn.messageMu.Unlock()
			}
		}
	}
}

// addConn registers a new connection and manages user connection mapping
func (s *Server) addConn(conn *Conn, req *http.Request) {
	uid := s.authentication.UserId(req)

	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	// Handle existing connection for user
	if c := s.userToConn[uid]; c != nil {
		c.Close()
	}

	s.connToUser[conn] = uid
	s.userToConn[uid] = conn
}

// GetConn retrieves a connection for a specific user
func (s *Server) GetConn(uid string) *Conn {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	return s.userToConn[uid]
}

// GetConns retrieves connections for multiple users
func (s *Server) GetConns(uids ...string) []*Conn {
	if len(uids) == 0 {
		return nil
	}

	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	res := make([]*Conn, 0, len(uids))
	for _, uid := range uids {
		res = append(res, s.userToConn[uid])
	}
	return res
}

// GetUsers retrieves user IDs associated with connections
func (s *Server) GetUsers(conns ...*Conn) []string {
	s.RWMutex.RLock()
	defer s.RWMutex.RUnlock()

	var res []string
	if len(conns) == 0 {
		// Get all users
		res = make([]string, 0, len(s.connToUser))
		for _, uid := range s.connToUser {
			res = append(res, uid)
		}
	} else {
		// Get specific users
		res = make([]string, 0, len(conns))
		for _, conn := range conns {
			res = append(res, s.connToUser[conn])
		}
	}

	return res
}

// Close terminates a connection and cleans up associated resources
func (s *Server) Close(conn *Conn) {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()

	uid := s.connToUser[conn]
	if uid == "" {
		return
	}

	delete(s.connToUser, conn)
	delete(s.userToConn, uid)

	conn.Close()
}

// SendByUserId sends a message to specific users by their IDs
func (s *Server) SendByUserId(msg interface{}, sendIds ...string) error {
	if len(sendIds) == 0 {
		return nil
	}

	return s.Send(msg, s.GetConns(sendIds...)...)
}

// Send transmits a message to specified connections
func (s *Server) Send(msg interface{}, conns ...*Conn) error {
	if len(conns) == 0 {
		return nil
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return err
		}
	}

	return nil
}

// AddRoutes registers new message handlers
func (s *Server) AddRoutes(rs []Route) {
	for _, r := range rs {
		s.routes[r.Method] = r.Handler
	}
}

// Start initializes the WebSocket server and begins listening for connections
func (s *Server) Start() {
	http.HandleFunc(s.patten, s.ServerWs)
	s.Info(http.ListenAndServe(s.addr, nil))
}

// Stop gracefully shuts down the WebSocket server
func (s *Server) Stop() {
	fmt.Println("Stop service")
}
