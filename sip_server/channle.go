package sip_server

import (
	"github.com/yushimeng/rock/sdp"
	"github.com/yushimeng/rock/sip"
)

type SipChannel struct {
	session              *SipSession
	channelId            string
	invitePlayState      sip.InviteState
	invitePlayRequest    *sip.Request
	invitePlayResponse   *sip.Response
	channelPlayState     sip.InviteState
	channelPlaybackState sip.InviteState
	channelDownloadState sip.InviteState
}

func (channel *SipChannel) SendInvite() {
	sess := channel.session
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

	req := sip.NewInviteRequest(sender, recipment, sess.registerRequest.Transport(), s.Builder())
	req.SetDestination(sess.source)
	req.SetTransport(sess.transport)

	sess.conn.WriteMsg(req)

	channel.invitePlayRequest = req
}
