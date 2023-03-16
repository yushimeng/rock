package sip

type SessionState int

const (
	SessionStateInit SessionState = iota
	SessionStateRegister
	SessionStateAlive
	SessionStateDestroy
)

type InviteState int

const (
	InviteStateInit InviteState = iota
	InviteStateStart
	InviteStateTrying
	InviteStateOk
	InviteStateDone
)

func (state *SessionState) OnTimeout() {
	*state = SessionStateDestroy
}

func (state *SessionState) OnRegister() {
	switch *state {
	case SessionStateInit:
		*state = SessionStateRegister
	default:
		return
	}
}

func (state *SessionState) OnUnregister() {
	*state = SessionStateDestroy
}

func (state *SessionState) OnKeepalive() {
	switch *state {
	case SessionStateRegister:
		*state = SessionStateAlive
	default:
		return
	}
}

func (state *InviteState) OnInvite() {
	*state = InviteStateStart
}
func (state *InviteState) OnInviteTrying() {
	switch *state {
	case InviteStateStart:
		*state = InviteStateTrying
	default:
		return
	}
}

func (state *InviteState) OnInviteOK() {
	switch *state {
	case InviteStateStart:
		*state = InviteStateOk
	case InviteStateTrying:
		*state = InviteStateOk
	default:
		return
	}
}

func (state *InviteState) OnInviteDone() {
	switch *state {
	// case InviteStateTrying:
	// 	*state = InviteStateDone
	case InviteStateOk:
		*state = InviteStateDone
	default:
		return
	}
}
