package sip_server

import (
	"context"
	"errors"

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
		log:       ctx.Value("log").(*logrus.Logger).WithFields(logrus.Fields{"caller": "SipServer"}),
	}

	srv.tp.OnMessage(srv.handleMessage)
	return srv
}

// handleMessage is entry for handling requests and responses from transport
func (srv *SipServer) handleMessage(msg sip.Message) {
	switch msg := msg.(type) {
	case *sip.Request:
		// TODO Consider making goroutine here already?
		srv.handleRequest(msg)
	case *sip.Response:
		// TODO Consider making goroutine here already?
		srv.handleResponse(msg)
	default:
		srv.log.Error("unsupported message, skip it")
		// todo pass up error?
	}
}

func (srv *SipServer) handleRequest(req *sip.Request) {
	srv.log.Println("get Request ", req.String())
	sess, err := srv.fetchOrCreateSession(req)
	if err == nil {
		sess.RequestChan() <- req
	}
}

func (srv *SipServer) handleResponse(res *sip.Response) {
	srv.log.Println("get reseponse ", res.String())
	sess, err := srv.featchSessionByResponse(res)
	if err != nil {

		sess.ResponseChan() <- res
	}
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
		sess = NewSipSession(srv, clientId, req.Transport(), req.Source())
		if sess == nil {
			return nil, errors.New("new sip session failed")
		}
		go sess.Serve(srv.ctx)
	}

	return sess, nil
}

func (srv *SipServer) featchSessionByResponse(res *sip.Response) (*SipSession, error) {
	/*
		From: "Bob" <sips:bob@biloxi.com> ;tag=a48s
		From: sip:+12125551212@phone2net.com;tag=887s
		From: Anonymous <sip:c8oqz84zk7z@privacy.org>;tag=hyh8
	*/
	to, bret := res.To()
	if !bret {
		return nil, errors.New("request to is nil")
	}
	srv.log.Println("to: ", to)
	key := to.Address.User
	srv.log.Debugf("res.to.address.user=%v", key)

	// TODO: mutex?
	sess, ok := srv.sessions[key]
	if !ok {
		sess = &SipSession{}
		go sess.Serve(srv.ctx)
	}

	return sess, nil
}

func (srv *SipServer) ListenAndServe(ctx context.Context, network string, addr string) {
	srv.tp.ListenAndServe(ctx, network, addr)
}
