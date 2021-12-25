package client

import (
	"net"
	"runtime/debug"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/Mrs4s/MiraiGo/client/internal/network"
	"github.com/Mrs4s/MiraiGo/client/internal/oicq"
	"github.com/Mrs4s/MiraiGo/utils"
)

// ConnectionQualityInfo 客户端连接质量测试结果
// 延迟单位为 ms 如为 9999 则测试失败 测试方法为 TCP 连接测试
// 丢包测试方法为 ICMP. 总共发送 10 个包, 记录丢包数
type ConnectionQualityInfo struct {
	// ChatServerLatency 聊天服务器延迟
	ChatServerLatency int64
	// ChatServerPacketLoss 聊天服务器ICMP丢包数
	ChatServerPacketLoss int
	// LongMessageServerLatency 长消息服务器延迟. 涉及长消息以及合并转发消息下载
	LongMessageServerLatency int64
	// LongMessageServerResponseLatency 长消息服务器返回延迟
	LongMessageServerResponseLatency int64
	// SrvServerLatency Highway服务器延迟. 涉及媒体以及群文件上传
	SrvServerLatency int64
	// SrvServerPacketLoss Highway服务器ICMP丢包数.
	SrvServerPacketLoss int
}

func (c *QQClient) ConnectionQualityTest() *ConnectionQualityInfo {
	if !c.Online.Load() {
		return nil
	}
	r := &ConnectionQualityInfo{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		var err error

		if r.ChatServerLatency, err = qualityTest(c.servers[c.currServerIndex].String()); err != nil {
			c.Error("test chat server latency error: %v", err)
			r.ChatServerLatency = 9999
		}

		if addr, err := net.ResolveIPAddr("ip", "ssl.htdata.qq.com"); err == nil {
			if r.LongMessageServerLatency, err = qualityTest((&net.TCPAddr{IP: addr.IP, Port: 443}).String()); err != nil {
				c.Error("test long message server latency error: %v", err)
				r.LongMessageServerLatency = 9999
			}
		} else {
			c.Error("resolve long message server error: %v", err)
			r.LongMessageServerLatency = 9999
		}
		if c.highwaySession.AddrLength() > 0 {
			if r.SrvServerLatency, err = qualityTest(c.highwaySession.SsoAddr[0].String()); err != nil {
				c.Error("test srv server latency error: %v", err)
				r.SrvServerLatency = 9999
			}
		}
	}()
	go func() {
		defer wg.Done()
		res := utils.RunICMPPingLoop(&net.IPAddr{IP: c.servers[c.currServerIndex].IP}, 10)
		r.ChatServerPacketLoss = res.PacketsLoss
		if c.highwaySession.AddrLength() > 0 {
			res = utils.RunICMPPingLoop(&net.IPAddr{IP: c.highwaySession.SsoAddr[0].AsNetIP()}, 10)
			r.SrvServerPacketLoss = res.PacketsLoss
		}
	}()
	start := time.Now()
	if _, err := utils.HttpGetBytes("https://ssl.htdata.qq.com", ""); err == nil {
		r.LongMessageServerResponseLatency = time.Since(start).Milliseconds()
	} else {
		c.Error("test long message server response latency error: %v", err)
		r.LongMessageServerResponseLatency = 9999
	}
	wg.Wait()
	return r
}

func (c *QQClient) connectFastest() error {
	c.Disconnect()
	addr, err := c.TCP.ConnectFastest(c.servers)
	if err != nil {
		c.Disconnect()
		return err
	}
	c.Debug("connected to server: %v [fastest]", addr.String())
	c.transport.NetLoop(c.pktProc, c.transport.ReadRequest)
	c.retryTimes = 0
	c.ConnectTime = time.Now()
	return nil
}

// connect 连接到 QQClient.servers 中的服务器
func (c *QQClient) connect() error {
	return c.connectFastest() // 暂时?
	/*c.Info("connect to server: %v", c.servers[c.currServerIndex].String())
	err := c.TCP.Connect(c.servers[c.currServerIndex])
	c.currServerIndex++
	if c.currServerIndex == len(c.servers) {
		c.currServerIndex = 0
	}
	if err != nil {
		c.retryTimes++
		if c.retryTimes > len(c.servers) {
			return errors.New("All servers are unreachable")
		}
		c.Error("connect server error: %v", err)
		return err
	}
	c.once.Do(func() {
		c.OnGroupMessage(func(_ *QQClient, _ *message.GroupMessage) {
			c.stat.MessageReceived.Add(1)
			c.stat.LastMessageTime.Store(time.Now().Unix())
		})
		c.OnPrivateMessage(func(_ *QQClient, _ *message.PrivateMessage) {
			c.stat.MessageReceived.Add(1)
			c.stat.LastMessageTime.Store(time.Now().Unix())
		})
		c.OnTempMessage(func(_ *QQClient, _ *TempMessageEvent) {
			c.stat.MessageReceived.Add(1)
			c.stat.LastMessageTime.Store(time.Now().Unix())
		})
		c.onGroupMessageReceipt("internal", func(_ *QQClient, _ *groupMessageReceiptEvent) {
			c.stat.MessageSent.Add(1)
		})
		go c.netLoop()
	})
	c.retryTimes = 0
	c.ConnectTime = time.Now()
	return nil*/
}

func (c *QQClient) QuickReconnect() {
	c.quickReconnect() // TODO "用户请求快速重连"
}

// quickReconnect 快速重连
func (c *QQClient) quickReconnect() {
	c.Disconnect()
	time.Sleep(time.Millisecond * 200)
	if err := c.connect(); err != nil {
		c.Error("connect server error: %v", err)
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "快速重连失败"})
		return
	}
	if err := c.registerClient(); err != nil {
		c.Error("register client failed: %v", err)
		c.Disconnect()
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "register error"})
		return
	}
}

// Disconnect 中断连接, 不释放资源
func (c *QQClient) Disconnect() {
	c.Online.Store(false)
	c.TCP.Close()
}

func (c *QQClient) send(call *network.Call) {
	if call.Done == nil {
		call.Done = make(chan *network.Call, 3) // use buffered channel
	}
	seq := uint16(call.Request.SequenceID)
	c.pendingMu.Lock()
	c.pending[seq] = call
	c.pendingMu.Unlock()

	err := c.sendPacket(c.transport.PackPacket(call.Request))
	if err != nil {
		c.pendingMu.Lock()
		call = c.pending[seq]
		delete(c.pending, seq)
		call.Err = err
		call.Done <- call
		c.pendingMu.Unlock()
	}
}

func (c *QQClient) sendReq(req *network.Request) {
	c.send(&network.Call{Request: req})
}

func (c *QQClient) call(req *network.Request) (*network.Response, error) {
	call := &network.Call{
		Request: req,
		Done:    make(chan *network.Call, 3),
	}
	c.send(call)
	select {
	case <-call.Done:
		return call.Response, call.Err
	case <-time.After(time.Second * 15):
		return nil, errors.New("Packet timed out")
	}
}

func (c *QQClient) callAndDecode(req *network.Request, decoder func(*QQClient, *network.Response) (interface{}, error)) (interface{}, error) {
	resp, err := c.call(req)
	if err != nil {
		return nil, err
	}
	return decoder(c, resp)
}

// sendPacket 向服务器发送一个数据包
func (c *QQClient) sendPacket(pkt []byte) error {
	err := c.TCP.Write(pkt)
	if err != nil {
		c.stat.PacketLost.Add(1)
	} else {
		c.stat.PacketSent.Add(1)
	}
	return errors.Wrap(err, "Packet failed to sendPacket")
}

// waitPacket
// 等待一个或多个数据包解析, 优先级低于 sendAndWait
// 返回终止解析函数
func (c *QQClient) waitPacket(cmd string, f func(interface{}, error)) func() {
	c.waiters.Store(cmd, f)
	return func() {
		c.waiters.Delete(cmd)
	}
}

// waitPacketTimeoutSyncF
// 等待一个数据包解析, 优先级低于 sendAndWait
func (c *QQClient) waitPacketTimeoutSyncF(cmd string, timeout time.Duration, filter func(interface{}) bool) (r interface{}, e error) {
	notifyChan := make(chan bool)
	defer c.waitPacket(cmd, func(i interface{}, err error) {
		if filter(i) {
			r = i
			e = err
			notifyChan <- true
		}
	})()
	select {
	case <-notifyChan:
		return
	case <-time.After(timeout):
		return nil, errors.New("timeout")
	}
}

// plannedDisconnect 计划中断线事件
func (c *QQClient) plannedDisconnect(_ *network.TCPListener) {
	c.Debug("planned disconnect.")
	c.stat.DisconnectTimes.Add(1)
	c.Online.Store(false)
}

// unexpectedDisconnect 非预期断线事件
func (c *QQClient) unexpectedDisconnect(_ *network.TCPListener, e error) {
	c.Error("unexpected disconnect: %v", e)
	c.stat.DisconnectTimes.Add(1)
	c.Online.Store(false)
	if err := c.connect(); err != nil {
		c.Error("connect server error: %v", err)
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "connection dropped by server."})
		return
	}
	if err := c.registerClient(); err != nil {
		c.Error("register client failed: %v", err)
		c.Disconnect()
		c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "register error"})
		return
	}
}

func (c *QQClient) pktProc(req *network.Request, netErr error) {
	if netErr != nil {
		switch true {
		case errors.Is(netErr, network.ErrConnectionBroken):
			go c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: netErr.Error()})
			c.QuickReconnect()
		case errors.Is(netErr, network.ErrSessionExpired) || errors.Is(netErr, network.ErrPacketDropped):
			c.Disconnect()
			go c.dispatchDisconnectEvent(&ClientDisconnectedEvent{Message: "session expired"})
		}
		c.Error("parse incoming packet error: %v", netErr)
		return
	}

	if req.EncryptType == network.EncryptTypeEmptyKey {
		m, err := c.oicq.Unmarshal(req.Body)
		if err != nil {
			c.Error("decrypt payload error: %v", err)
			if errors.Is(err, oicq.ErrUnknownFlag) {
				go c.quickReconnect() // TODO "服务器发送未知响应"
			}
		}
		req.Body = m.Body
	}

	defer func() {
		if pan := recover(); pan != nil {
			c.Error("panic on decoder %v : %v\n%s", req.CommandName, pan, debug.Stack())
			c.Dump("packet decode error: %v - %v", req.Body, req.CommandName, pan)
		}
	}()

	c.Debug("rev resp: %v seq: %v", req.CommandName, req.SequenceID)
	c.stat.PacketReceived.Add(1)

	// snapshot of read call
	c.pendingMu.Lock()
	call := c.pending[uint16(req.SequenceID)]
	if call != nil {
		call.Response = &network.Response{
			SequenceID:  req.SequenceID,
			CommandName: req.CommandName,
			Body:        req.Body,
			Params:      call.Request.Params,
			// Request:     nil,
		}
	}
	c.pendingMu.Unlock()
	if call != nil && call.Request.CommandName == req.CommandName {
		select {
		case call.Done <- call:
		default:
			// we don't want blocking
		}
		return
	}

	if decoder, ok := decoders[req.CommandName]; ok {
		// found predefined decoder
		resp := network.Response{
			SequenceID:  req.SequenceID,
			CommandName: req.CommandName,
			Body:        req.Body,
			// Request:     nil,
		}
		decoded, err := decoder(c, &resp)
		if err != nil {
			c.Debug("decode req %v error: %+v", req.CommandName, err)
		}
		if f, ok := c.waiters.Load(req.CommandName); ok { // 在不存在handler的情况下触发wait
			f.(func(interface{}, error))(decoded, err)
		}
	} else {
		c.Debug("Unhandled Command: %s\nSeq: %d\nThis message can be ignored.", req.CommandName, req.SequenceID)
	}
}
