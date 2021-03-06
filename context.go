package kingim

import (
	"google.golang.org/protobuf/proto"
	"kingim/logger"
	"kingim/wire"
	"kingim/wire/pkt"
	"sync"
)

type Session interface {
	GetChannelId() string
	GetGateId() string
	GetAccount() string
	GetZone() string
	GetIsp() string
	GetRemoteIP() string
	GetDevice() string
	GetApp() string
	GetTags() []string
}

type Context interface {
	Dispather
	SessionStorage
	Header() *pkt.Header
	ReadBody(val proto.Message) error
	Session() Session
	RespWithError(status pkt.Status, err error) error
	Resp(status pkt.Status, body proto.Message) error
	Dispatch(body proto.Message, recvs ...*Location) error
}

type HandlerFunc func(ctx Context)

type HandlersChain []HandlerFunc

type ContextImpl struct {
	sync.Mutex
	Dispather
	SessionStorage
	handlers HandlersChain
	index int
	request *pkt.LogicPkt
	session Session
}

func BuildContext() Context {
	return &ContextImpl{}
}

func (c*ContextImpl) Next() {
	if c.index >= len(c.handlers) {
		return
	}
	f := c.handlers[c.index]
	f(c)
	c.index++
}

func (c*ContextImpl) Header() *pkt.Header {
	return &c.request.Header
}
func (c*ContextImpl) ReadBody(val proto.Message) error {
	return c.request.ReadBody(val)
}
func (c*ContextImpl) RespWithError(status pkt.Status, err error) error {
	return c.Resp(status, &pkt.ErrorResp{Message: err.Error()})
}

// 给发送方回应一条信息
func (c*ContextImpl) Resp(status pkt.Status, body proto.Message) error {
	packet := pkt.NewFrom(&c.request.Header)
	packet.Status = status
	packet.WriteBody(body)
	packet.Flag = pkt.Flag_Response
	logger.Debugf("<-- Resp to %s command:%s  status: %v body: %s\", c.Session().GetAccount(), &c.request.Header, status, body")
	err := c.Push(c.Session().GetGateId(), []string{c.Session().GetChannelId()},packet)
	if err != nil {
		logger.Error(err)
	}
	return err
}

func (c*ContextImpl) Dispatch(body proto.Message, recvs...*Location) error {
	if len(recvs) == 0 {
		return nil
	}
	group := make(map[string][]string)
	packet := pkt.NewFrom(&c.request.Header)
	packet.WriteBody(body)
	packet.Flag = pkt.Flag_Response
	logger.Debugf("<-- Dispatch to %d users command:%s", len(recvs), &c.request.Header)
	for _,recv := range recvs {
		if recv.ChannelId == c.Session().GetChannelId() {
			continue
		}
		// 判断是否已存在
		if _,ok := group[recv.GateId]; !ok {
			group[recv.GateId] = make([]string, 0)
		}
		// 把来组相同网关的ChannelID 组合在一个数组
		group[recv.GateId] = append(group[recv.GateId], recv.ChannelId)
	}
	// 更具 网关把信息推送出去
	for gateway, ids := range group {
		err := c.Push(gateway, ids, packet)
		if err != nil {
			logger.Error(err)
		}
		return err
	}
	return nil
}
func (c*ContextImpl) Session() Session {
	if c.session == nil {
		server, _ := c.request.GetMeta(wire.MetaDestServer)
		c.session = &pkt.Session{
			ChannelId: c.request.ChannelId,
			GateId: server.(string),
			Tags: []string{"AutoGenerated"},
		}
	}
	return c.session
}
func (c*ContextImpl) reset() {
	c.request = nil
	c.index = 0
	c.handlers = nil
	c.session = nil
}
