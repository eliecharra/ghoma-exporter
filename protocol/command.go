package protocol

type Command byte

const (
	CmdInit1          Command = 0x02
	CmdInit1Reply     Command = 0x03
	CmdHeartBeat      Command = 0x04
	CmdInit2          Command = 0x05
	CmdHeartBeatReply Command = 0x06
	CmdInit2Reply     Command = 0x07
	CmdStatus         Command = 0x90
)

func (c Command) String() string {
	switch c {
	case CmdInit1:
		return "INIT1"
	case CmdInit1Reply:
		return "INIT1_REPLY"
	case CmdHeartBeat:
		return "HEARTHBEAT"
	case CmdInit2:
		return "INIT2"
	case CmdHeartBeatReply:
		return "HEARTHBEAT_REPLY"
	case CmdInit2Reply:
		return "INIT2_REPLY"
	case CmdStatus:
		return "STATUS"
	default:
		return "UNKNOWN"
	}
}
