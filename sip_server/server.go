package sip_server

import (
	"context"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/yushimeng/rock/sip"
)

type SipServerConf struct {
	serverId       string
	serverPort     int
	serverIp       string
	ServerRealm    string
	pull_immediate bool
}

// // RequestHandler is a callback that will be called on the incoming request
// type RequestHandler func(req *Request, tx ServerTransaction)

type SipServer struct {
	*UserAgent

	sessions map[string]*SipSession
	// requestHandlers map of all registered request handlers
	// requestHandlers  map[RequestMethod]RequestHandler
	// unhandledHandler RequestHandler
	ctx  context.Context
	log  *logrus.Entry
	conf *SipServerConf
}

// func (srv *SipServer) defaultUnhandledHandler(req *Request, tx ServerTransaction) {
// 	srv.log.Info("SIP request handler not found")
// 	res := NewResponseFromRequest(req, 405, "Method Not Allowed", nil)
// 	// Send response directly and let transaction terminate
// 	if err := srv.WriteResponse(res); err != nil {
// 		srv.log.Info("respond '405 Method Not Allowed' failed")
// 	}
// }

// WriteResponse will proxy message to transport layer. Use it in stateless mode
func (srv *SipServer) WriteResponse(r *sip.Response) error {
	return srv.tp.WriteMsg(r)
}

func NewSipServer(ctx context.Context) *SipServer {
	ua, err := NewUA()
	if err != nil {
		return nil
	}

	serverip, err := sip.GetLocalIp()
	if err != nil {
		return nil
	}

	cfg := &SipServerConf{
		serverId:       "34020000002000000001",
		serverPort:     5060,
		serverIp:       serverip,
		ServerRealm:    "3402000000",
		pull_immediate: true,
	}
	srv := &SipServer{
		UserAgent: ua,
		ctx:       ctx,
		conf:      cfg,
		sessions:  make(map[string]*SipSession),
		log:       ctx.Value("log").(*logrus.Logger).WithFields(logrus.Fields{"caller": "SipServer"}),
	}

	srv.tp.OnMessage(srv.handleMessage)
	return srv
}

// handleMessage is entry for handling requests and responses from transport
func (srv *SipServer) handleMessage(msg sip.Message) {
	var err error
	switch msg := msg.(type) {
	case *sip.Request:
		// TODO Consider making goroutine here already?
		err = srv.handleRequest(msg)
	case *sip.Response:
		// TODO Consider making goroutine here already?
		err = srv.handleResponse(msg)
	default:
		srv.log.Error("unsupported message, skip it")
		// todo pass up error?
	}
	if err != nil {
		srv.log.Errorf("handle message failed, err:%v", err)
	}
}

func (srv *SipServer) handleRequest(req *sip.Request) error {
	srv.log.Println("get Request ", req.String())
	sess, err := srv.fetchOrCreateSession(req)
	if err == nil {
		sess.RequestChan() <- req
	}
	return nil
}

func (srv *SipServer) handleResponse(res *sip.Response) (err error) {
	srv.log.Println("get reseponse ", res.String())
	sess, err := srv.fetchSessionByResponse(res)
	if err != nil {
		return err
	}
	sess.ResponseChan() <- res
	return
}

func (srv *SipServer) fetchOrCreateSession(req *sip.Request) (*SipSession, error) {
	/*
		From: "Bob" <sips:bob@biloxi.com> ;tag=a48s
		From: sip:+12125551212@phone2net.com;tag=887s
		From: Anonymous <sip:c8oqz84zk7z@privacy.org>;tag=hyh8
	*/
	from, bret := req.From()
	if !bret {
		return nil, errors.New("request frome is nil")
	}
	clientId := from.Address.User

	// TODO: mutex?
	sess, ok := srv.sessions[clientId]
	if !ok {
		srv.log.Infof("fetch session by req clientId<%s> failed, create it.", clientId)
		sess = NewSipSession(srv, clientId, req.Transport(), req.Source())
		if sess == nil {
			return nil, errors.New("new sip session failed")
		}
		go sess.Serve(srv.ctx)
	}

	return sess, nil
}

func (srv *SipServer) fetchSessionByResponse(res *sip.Response) (*SipSession, error) {
	to, bret := res.To()
	if !bret {
		return nil, errors.New("request to is nil")
	}
	srv.log.Println("to: ", to)
	clientId := to.Address.User
	srv.log.Debugf("res.to.address.user=%v", clientId)

	// TODO: mutex?
	// TODO: response should not create session.
	sess, ok := srv.sessions[clientId]
	if !ok {
		srv.log.Warnf("fetech session by user failed. clientId:%s", clientId)
		return nil, fmt.Errorf("session not exist,%s", clientId)
		// sess = NewSipSession(srv, clientId, res.Transport(), res.Source())
		// if sess == nil {
		// 	return nil, errors.New("new sip session failed")
		// }
		// go sess.Serve(srv.ctx)

		// sess = &SipSession{}
		// go sess.Serve(srv.ctx)
	}

	return sess, nil
}

func (srv *SipServer) ListenAndServe(ctx context.Context, network string, addr string) {
	srv.tp.ListenAndServe(ctx, network, addr)
}
