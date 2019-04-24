package srp

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"sync"
)

type RegistryMsgType int

const (
	MsgRegisterReq RegistryMsgType = iota
	MsgRegisterResp
	MsgServicePushReq
	MsgServicePushCancelReq
	MsgServicePushResp
	MsgCloseReq
)

type MsgRegistry struct {
	Type RegistryMsgType
	// MsgRegisterReq
	Service string
	Addr    string
	// MsgServiceListResp
	PushStart bool
	PushEnd   bool
	Services  map[string][]string
	// resp
	Success bool
	Error   error
}

// Registry
type Registry struct {
	Addr         string
	Services     map[string][]string
	servicesLock sync.Mutex
	pushList     map[*websocket.Conn]int
	pushListLock sync.Mutex
}

func NewRegistry(addr string) *Registry {
	return &Registry{
		Addr:     addr,
		Services: make(map[string][]string),
	}
}

func (r *Registry) addPushList(conn *websocket.Conn) {
	r.pushListLock.Lock()
	if r.pushList == nil {
		r.pushList = make(map[*websocket.Conn]int)
	}
	r.pushList[conn] = 0
	r.pushListLock.Unlock()
}

func (r *Registry) removePushList(conn *websocket.Conn) {
	r.pushListLock.Lock()
	if r.pushList != nil {
		r.pushList[conn] = -1
	}
	r.pushListLock.Unlock()
}

func (r *Registry) push() {
	for conn, cnt := range r.pushList {
		msg := &MsgRegistry{Type: MsgServicePushResp, Services: r.Services}
		if cnt == -1 {
			msg.PushEnd = true
			delete(r.pushList, conn)
		} else {
			if cnt == 0 {
				msg.PushStart = true
			}
			r.pushList[conn]++
		}
		check(conn.WriteJSON(msg), "push")
	}
}

func (r *Registry) addService(registerMsg MsgRegistry) {
	r.servicesLock.Lock()
	addrs := r.Services[registerMsg.Service]
	exists := false
	for _, addr := range addrs {
		if addr == registerMsg.Addr {
			exists = true
		}
	}
	if !exists {
		r.Services[registerMsg.Service] = append(addrs, registerMsg.Addr)
		r.push()
	}
	r.servicesLock.Unlock()
}

func (r *Registry) removeService(registerMsg MsgRegistry) {
	r.servicesLock.Lock()
	addrs := r.Services[registerMsg.Service]
	j := 0
	for _, addr := range addrs {
		if addr != registerMsg.Addr {
			addrs[j] = addr
			j++
		}
	}
	r.Services[registerMsg.Service] = addrs[:j]
	r.servicesLock.Unlock()
}

func (r *Registry) closeConn(conn *websocket.Conn, registerMsg MsgRegistry) bool {
	r.removeService(registerMsg)
	_ = conn.Close()
}

func (r *Registry) handleIncomingConn(conn *websocket.Conn) {
	var msg MsgRegistry
	var err error
	// now limit a connection can only register one service at a time
	var registerMsg MsgRegistry
	for {
		err = conn.ReadJSON(&msg)
		if check(err) {
			r.closeConn(conn, registerMsg)
			return
		}
		switch msg.Type {
		case MsgCloseReq:
			_ = conn.Close()
		case MsgRegisterReq:
			if msg.Addr == "" {
				msg.Addr = conn.RemoteAddr().String()
			}
			registerMsg = msg
			r.removeService(registerMsg)
			r.addService(registerMsg)
			err = conn.WriteJSON(&MsgRegistry{Type: MsgRegisterResp, Success: true})
		case MsgServicePushReq:
			r.addPushList(conn)
		case MsgServicePushCancelReq:
			r.removePushList(conn)
		}
		if check(err) {
			r.closeConn(conn, registerMsg)
			return
		}
	}
}

func (r *Registry) Run() error {
	upgrader := websocket.Upgrader{}
	s := gin.New()
	s.NoRoute(func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if check(err, "upgrade") {
			return
		}
		log("connect:", conn.RemoteAddr())
		r.handleIncomingConn(conn)
	})
	log("listen:", r.Addr)
	return s.Run(r.Addr)
}

// RemoteRegistry
type RemoteRegistry struct {
	RegistryAddr string
	Conn         *websocket.Conn
	Services     map[string][]string
	recv         sync.Mutex
}

func (rr *RemoteRegistry) Connect() (err error) {
	d := new(websocket.Dialer)
	rr.Conn, _, err = d.Dial(rr.RegistryAddr, nil)
	return
}

func (rr *RemoteRegistry) handleIncomingMsg(msg *MsgRegistry) error {
	switch msg.Type {
	case MsgRegisterResp:
		if !msg.Success {
			return msg.Error
		}
	case MsgServicePushResp:
		if !msg.Success {
			return msg.Error
		}
		if msg.PushStart {
			rr.recv.Lock()
		} else if msg.PushEnd {
			rr.recv.Unlock()
		}
		rr.Services = msg.Services
	default:
		return errors.New("registry msg type not supported")
	}
	return nil
}

func (rr *RemoteRegistry) call(msg *MsgRegistry, requireReply bool) (err error) {
	if requireReply {
		rr.recv.Lock()
	}
	err = rr.Conn.WriteJSON(*msg)
	if err != nil {
		return
	}
	if requireReply {
		var result MsgRegistry
		err = rr.Conn.ReadJSON(&result)
		if err != nil {
			return
		}
		rr.recv.Unlock()
		return rr.handleIncomingMsg(&result)
	}
	return
}

func (rr *RemoteRegistry) Disconnect() {
	if rr.Conn != nil {
		_ = rr.call(&MsgRegistry{Type: MsgCloseReq}, false)
		_ = rr.Conn.Close()
		rr.Conn = nil
	}
}

func (rr *RemoteRegistry) RegisterAddr(name string, addr string) error {
	return rr.call(&MsgRegistry{Type: MsgRegisterReq, Service: name, Addr: addr}, true)
}

func (rr *RemoteRegistry) Register(name string) error {
	return rr.RegisterAddr(name, "")
}

func (rr *RemoteRegistry) Subscribe(handler func()) error {
	return rr.call(&MsgRegistry{Type: MsgServicePushReq}, false)
}

func (rr *RemoteRegistry) Unsubscribe() error {
	return rr.call(&MsgRegistry{Type: MsgServicePushCancelReq}, false)
}

func (rr *RemoteRegistry) Query(name string) (addrs []string) {
	if rr.Services == nil {
		return
	}
	return rr.Services[name]
}