package sip_server

import (
	"context"
	"encoding/xml"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/sdp"
	"github.com/yushimeng/rock/sip"
	"github.com/yushimeng/rock/transport"
	"github.com/yushimeng/rock/util"
)

type SipSession struct {
	clientId    string
	transport   string
	source      string
	peerIp      string
	peerPort    int
	channelList map[string]*SipChannel
	conn        transport.Connection
	sipState    sip.SessionState

	registerRequest  *sip.Request
	registerTimer    *sip.Timer
	keepaliveTimer   *sip.Timer
	catalogTimer     *sip.Timer
	requestChan      chan *sip.Request
	responseChan     chan *sip.Response
	keepaliveTimeout int
	catalogTimeout   int
	onDestroyed      func(string)
	conf             *SipServerConf
	log              *logrus.Entry
}

// for futher make sip session by response, we set params so many, and simple.
func NewSipSession(srv *SipServer, clientId, transport, addr string) *SipSession {
	ctx := srv.ctx
	logger := ctx.Value(util.IdentifyLog).(*logrus.Logger).WithFields(logrus.Fields{string(util.IdentifyCaller): clientId})
	conn, err := srv.tp.GetConnection(transport, addr)
	if err != nil {
		logger.Errorf("failed to get connection, transport=%s addr=%s", transport, addr)
		return nil
	}

	sess := &SipSession{
		sipState:         sip.SessionStateInit,
		clientId:         clientId,
		transport:        transport,
		conn:             conn,
		requestChan:      make(chan *sip.Request, 1),
		responseChan:     make(chan *sip.Response, 1),
		conf:             srv.conf,
		keepaliveTimeout: 60,
		catalogTimeout:   120,
		log:              logger,
		onDestroyed:      srv.OnDeviceDestroyed,
	}
	srv.sessions[clientId] = sess
	return sess
}

// return a recv-only response chan.
func (sess *SipSession) ResponseChan() chan<- *sip.Response {
	return sess.responseChan
}

// return a recv-only request chan.
func (sess *SipSession) RequestChan() chan *sip.Request {
	return sess.requestChan
}

func (sess *SipSession) Destroy() {
	if sess.registerTimer != nil {
		sess.registerTimer.Stop()
	}

	if sess.keepaliveTimer != nil {
		sess.registerTimer.Stop()
	}

	if sess.catalogTimer != nil {
		sess.catalogTimer.Stop()
	}

	sess.onDestroyed(sess.clientId)
}

func (sess *SipSession) Serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer sess.Destroy()

	sess.registerTimer = sip.NewTimer(sess.keepaliveTimeout, 1)
	sess.keepaliveTimer = sip.NewTimer(sess.keepaliveTimeout, 3)
	sess.catalogTimer = sip.NewTimer(sess.catalogTimeout, 3)

	sess.log.Info("sipSession entry loop...")
	for {
		select {
		case <-sess.registerTimer.Timer.C:
			sess.registerTimer.TimeoutCnt++
			sess.sipState.OnTimeout()
			sess.log.Error("timerRegister timeout")
			return
		case <-sess.keepaliveTimer.Timer.C:
			sess.keepaliveTimer.TimeoutCnt++
			if sess.keepaliveTimer.TimeoutCnt > sess.keepaliveTimer.MaxTimeoutCnt {
				sess.sipState.OnTimeout()
				sess.log.Error("timerKeepalive timeout")
				return
			}
			sess.keepaliveTimer.Reset(sess.keepaliveTimeout)
		case <-sess.catalogTimer.Timer.C:
			// send catalog
			sess.SendCatalog()
			sess.catalogTimer.Reset(sess.catalogTimeout)
		case req := <-sess.requestChan:
			// TODO: seperate req/res here. merge req&res chan.
			sess.processRequest(req)
		case res := <-sess.responseChan:
			sess.processResponse(res)
		case <-ctx.Done():
			sess.log.Println("session recv ctx done.")
			return
		}

		if sess.conf.pull_immediate {
			for _, ch := range sess.channelList {
				if ch.invitePlayState == sip.InviteStateInit {
					ch.SendInvite()
					ch.invitePlayState.OnInvite()
				}
			}
		}
	}
}

type Info struct {
}

type RequestBody struct {
	XMLName  xml.Name `xml:"Notify"`
	CmdType  string   `xml:"CmdType"`
	Sn       int      `xml:"SN"`
	DeviceID string   `xml:"DeviceID"`
	Status   string   `xml:"Status"`
	Info     Info     `xml:"Info"`
}

func (sess *SipSession) refreshSessionInfo(req *sip.Request) {
	sess.source = req.Source()
	if via, exist := req.Via(); exist {
		sess.peerIp = via.Host
		sess.peerPort = via.Port
	}
}

func (sess *SipSession) processRequest(req *sip.Request) {
	if req.Method == sip.REGISTER {
		sess.processRegister(req)
		return
	}

	if req.Method == sip.MESSAGE {
		v := RequestBody{}
		if err := xml.Unmarshal(req.Body(), &v); err != nil {
			sess.log.Errorf("message body Unmarshal error: %v", err)
			return
		}
		if v.CmdType == "Keepalive" {
			sess.processKeepalive(req, &v)
		} else {
			sess.log.Warnf("recv unrecognized CmdType:%s", v.CmdType)
		}
	}

}

func (sess *SipSession) processRegister(req *sip.Request) {
	expires := req.GetHeader("expires")
	if expires == nil {
		sess.log.Error("recv register without cseq")
		return
	}
	expiresInt, err := strconv.Atoi(expires.Value())
	if err != nil {
		sess.log.Error("recv register with invalied expires")
	}
	sess.registerTimer.Reset(expiresInt)

	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	sess.conn.WriteMsg(res)

	sess.registerRequest = req
	sess.refreshSessionInfo(req)
	if expiresInt == 0 {
		sess.sipState.OnUnregister()
		return
	}

	sess.sipState.OnRegister()
	if len(sess.channelList) == 0 {
		sess.catalogTimer.Reset(0)
	}
}

func (sess *SipSession) processKeepalive(req *sip.Request, body *RequestBody) {
	res := sip.NewResponseFromRequest(req, 200, "OK", nil)
	sess.conn.WriteMsg(res)

	sess.sipState.OnKeepalive()
	sess.keepaliveTimer.Reset(sess.keepaliveTimeout)
}

func (sess *SipSession) processResponse(res *sip.Response) {
	if res.IsProvisional() {
		return
	}
	if res.IsInviteOk() {
		sess.invitePlayResponse = res
		ack := sip.NewAckRequest(sess.invitePlayRequest, sess.invitePlayResponse, nil)
		sess.conn.WriteMsg(ack)
	}
}

func (sess *SipSession) SendCatalog() {

	sender := &sip.Uri{
		User: sess.conf.serverId,
		Host: sess.conf.serverIp,
		Port: sess.conf.serverPort,
	}

	recipment := &sip.Uri{
		User: sess.clientId,
		Host: sess.peerIp,
		Port: sess.peerPort,
	}

	ssrc := sip.RandInt()
	s := &sdp.SDP{
		Version:   0,
		OwnerId:   sess.conf.serverId,
		OwnerHost: sess.conf.serverIp,
		OwnerPort: sess.conf.serverPort,
		Ssrc:      ssrc,
		SendOnly:  sdp.RecvOnly,
		Session:   sdp.Play,
	}
	req := sip.NewCatalogRequest(sender, recipment, sess.registerRequest.Transport(), s.Builder())
	// req := sip.NewInviteRequest()
	req.SetDestination(sess.source)
	req.SetTransport(sess.transport)

	sess.conn.WriteMsg(req)

	// channel.invitePlayRequest = req
}
