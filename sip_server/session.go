package sip_server

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/sip"
	"github.com/yushimeng/rock/transport"
)

type SipSession struct {
	// client info
	clientId  string
	transport sip.Transport
	peerIp    string
	peerPort  int
	conn      transport.Connection

	// server info
	registerRequest *sip.Request
	timerRegister   *time.Timer
	timerKeepalive  *time.Timer
	request_chan    chan *sip.Request
	response_chan   chan *sip.Response
	log             *logrus.Entry

	// conf
	// pull_immediate bool
	conf *SipServerConf
}

// for futher make sip session by response, we set params so many, and simple.
func NewSipSession(srv *SipServer, clientId, transport, addr string) *SipSession {
	ctx := srv.ctx
	logger := ctx.Value("log").(*logrus.Logger).WithFields(logrus.Fields{"caller": clientId})
	conn, err := srv.tp.GetConnection(transport, addr)
	if err != nil {
		logger.Errorf("failed to get connection, transport=%s addr=%s", transport, addr)
		return nil
	}

	sess := &SipSession{
		clientId:      clientId,
		conn:          conn,
		request_chan:  make(chan *sip.Request, 1),
		response_chan: make(chan *sip.Response, 1),
		log:           logger,
		// pull_immediate: true,
		conf: srv.conf,
	}

	return sess
}

// return a recv-only response chan.
func (sess *SipSession) ResponseChan() chan<- *sip.Response {
	return sess.response_chan
}

// return a recv-only request chan.
func (sess *SipSession) RequestChan() chan *sip.Request {
	return sess.request_chan
}

func (sess *SipSession) Serve(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sess.timerRegister = time.NewTimer(60 * time.Second)
	sess.timerKeepalive = time.NewTimer(3 * 60 * time.Second)

	sess.log.Info("sipSession entry loop...")
	for {
		select {
		case <-sess.timerRegister.C:
			sess.log.Warn("timerRegister timeout")
			return
		case <-sess.timerKeepalive.C:
			sess.log.Warn("timerKeepalive timeout")
			return
		// case invite<- sess.invite_chan:
		case req := <-sess.request_chan:
			sess.processRequest(req)
		case res := <-sess.response_chan:
			sess.processResponse(res)
		case <-ctx.Done():
			sess.log.Println("session recv ctx done.")
			return
		}
	}
}

/*
MESSAGE sip:34020000002000000001@3402000000 SIP/2.0
Via: SIP/2.0/UDP 192.168.10.8:60719;rport=60719;branch=z9hG4bK430967578
Max-Forwards: 70
To: <sip:34020000002000000001@3402000000>
From: <sip:34020000002000000719@3402000000>;tag=1555083253
Call-ID: 1910992793
CSeq: 20 MESSAGE
Content-Type: Application/MANSCDP+xml
User-Agent: IP Camera
Content-Length: 177

<?xml version="1.0" encoding="GB2312"?>
<Notify>
<CmdType>Keepalive</CmdType>
<SN>42</SN>
<DeviceID>34020000002000000719</DeviceID>
<Status>OK</Status>
<Info>
</Info>
</Notify>
*/
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

func (sess *SipSession) processRequest(req *sip.Request) {

	if req.Method == sip.REGISTER {
		expires := req.GetHeader("expires")
		if expires == nil {
			sess.log.Error("recv register without cseq")
			return
		}
		expiresInt, err := strconv.Atoi(expires.Value())
		if err != nil {
			sess.log.Error("recv register with invalied expires")
		}

		sess.timerRegister.Reset(time.Duration(expiresInt) * (time.Second))

		res := sip.NewResponseFromRequest(req, 200, "OK", nil)
		sess.conn.WriteMsg(res)

		sess.registerRequest = req
		//
		sess.log.Println("pull_immediate: ", sess.conf.pull_immediate)

		if sess.conf.pull_immediate {
			sess.sendInvite()
		}
		return
	}

	if req.Method == sip.MESSAGE {
		v := RequestBody{}
		err := xml.Unmarshal(req.Body(), &v)
		if err != nil {
			fmt.Printf("error: %v", err)
			return
		}
		fmt.Println(v)
	}

}
func (sess *SipSession) sendInvite() {
	sender := sip.Uri{
		User: sess.conf.serverId,
		Host: sess.conf.serverIp,
		Port: sess.conf.serverPort,
	}

	recipment := sip.Uri{
		User: sess.clientId,
		Host: sess.peerIp,
		Port: sess.peerPort,
	}

	req := sip.NewInviteRequest(sender, recipment, sess.registerRequest.Transport(), nil)

	// sdp := "v=0 \
	// o=34020000002000000719 0 0 IN IP4 192.168.10.10 \
	// s=Play
	// c=IN IP4 192.168.10.10
	// t=0 0
	// m=video 5060 RTP/AVP 96 97 98
	// a=recvonly
	// a=rtpmap:96 PS/90000
	// a=rtpmap:97 H264/90000
	// a=rtpmap:98 MPEG4/90000
	// y=1760090202"
	sess.log.Println("send invite", req)
	req.SetDestination(sess.registerRequest.Source())
	req.SetTransport(sess.registerRequest.Transport())
	sess.conn.WriteMsg(req)
}
func (sess *SipSession) processResponse(req *sip.Response) {

}
