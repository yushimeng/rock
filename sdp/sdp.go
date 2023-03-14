package sdp

import (
	"fmt"
	"strings"
)

const (
	// 编码格式
	CODEC_MPEG4 int = iota + 1
	CODEC_H264
	CODEC_SVAC
	CODEC_3GP
)

const (
	RESOLUTION_QCIF int = iota + 1
	RESOLUTION_CIF
	RESOLUTION_4CIF
	RESOLUTION_D1
	RESOLUTION_720P
	RESOLUTION_1080P
)

const (
	CBR int = iota + 1
	VBR
)
const (
	G711 int = iota + 1
	G723
	G729
	G722
)

const (
	// G.723.1 中 使 用
	AUDIO_BIT_RATE_5300 int = iota + 1
	// G.723.1 中 使 用
	AUDIO_BIT_RATE_6300
	// G .7 2 9 中 使 用
	AUDIO_BIT_RATE_8000
	// G.722.1 中 使 用
	AUDIO_BIT_RATE_16000
	// G.722.1 中 使 用
	AUDIO_BIT_RATE_24000
	// G.722.1 中 使 用
	AUDIO_BIT_RATE_32000
	// G.722.1 中 使 用
	AUDIO_BIT_RATE_48000
	// G.711 中 使 用
	AUDIO_BIT_RATE_64000
)

const (
	// G.711/ G.723.1/ G.729 中 使 用
	AUDIO_SAMPLE_8000 int = iota + 1
	// G .7 2 2 .1 中 使 用
	AUDIO_SAMPLE_14000
	// G .7 2 2 .1 中 使 用
	AUDIO_SAMPLE_16000
	// G .7 2 2 .1 中 使 用
	AUDIO_SAMPLE_32000
)

var (
	// 0~99
	FRAME_RATE int = 25
	// kbps
	BIT_RATE int = 20
	// AUDIO_BIT_RATE int = 8000
)

type MLINE struct {
	Audio bool
	Rport int
	// tcp or udp
	Transport      string
	RecognizablePT []int
}
type CLINE struct {
	Proto   string
	Address string
}

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
type SDP struct {
	// 	v= (protocolversion)
	V int
	// o= (owner/creatorandsesionidentifier)
	O string
	// s= (sesionname)
	// “Play”代 表 实 时 点 播 ;“Playback”代 表 历 史 回 放 ;“Download”代 表 文 件 下 载 ;“Talk” 代表语音对讲。
	S string
	// u=* (URIofdescription)
	U string
	// 	c=* (connectioninformation-notrequiredifincludedinalmedia)
	C string
	// Timedescription:
	// t= (timethesesionisactive)
	T string
	// Mediadescription
	// m= (medianameandtransportaddres)
	M MLINE
	// a=* (zeroormoremediaatributelines)
	// b=* (bandwidthinformation)
	B string
	// y = * (S S R C )
	Y string
	// f= * (媒 体 描 述 )
	// f= v/编码格式/分辨率/帧率/码率类型/码率大小a/编码格式/码率大小/采样率
	F string
}

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
func (sdp *SDP) Builder() []byte {
	sb := new(strings.Builder)
	sb.Grow(512)
	ownerId := "34020000002000000719"
	ownerHost := "192.168.10.10"
	ownerPort := 5060
	ssrc := "1234567890"
	sb.WriteString(fmt.Sprintf("v=%d", sdp.V))
	sb.WriteString(fmt.Sprintf("o=%s 0 0 IN IP4 %s", ownerId, ownerHost))
	sb.WriteString(fmt.Sprintf("s=%s", sdp.S))

	// sb.WriteString(fmt.Sprintf("u=%d", sdp.U))
	sb.WriteString(fmt.Sprintf("c=IN IP4 %s", ownerHost))
	sb.WriteString(fmt.Sprintf("t=%d %d", 0, 0))
	sb.WriteString(fmt.Sprintf("m=video %d RTP/AVP 96 97 98", ownerPort))
	sb.WriteString(fmt.Sprintf("a=recvonly"))
	sb.WriteString(fmt.Sprintf("a=rtpmap:96 PS/90000"))
	sb.WriteString(fmt.Sprintf("a=rtpmap:97 H264/90000"))
	sb.WriteString(fmt.Sprintf("a=rtpmap:98 MPEG4/90000"))

	sb.WriteString(fmt.Sprintf("y=%s", ssrc))
	sb.WriteString(fmt.Sprintf("f=v/%d/%d/%d/%d/%da/%d/%d/%d",
		CODEC_H264, RESOLUTION_1080P, FRAME_RATE, CBR, BIT_RATE,
		G711, AUDIO_BIT_RATE_64000, AUDIO_SAMPLE_8000))
	return []byte(sb.String())
}
