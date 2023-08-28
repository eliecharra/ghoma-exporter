package protocol

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
)

type Status struct {
	Switch *bool
	Energy *Energy
}

type Energy struct {
	Kind  uint8
	Value int64
}

type Message struct {
	Payload []byte
	Command Command
	Status  *Status
}

var ErrCmdUnknown = errors.New("unknown command")

func Parse(payload []byte) (*Message, error) {
	msg := &Message{Payload: payload}
	if Command(payload[0]).String() == "UNKNOWN" {
		return msg, ErrCmdUnknown
	}
	msg.Command = Command(payload[0])

	if msg.Command == CmdStatus {
		msg.Status = &Status{}
		if bytes.Compare(msg.Payload[9:9+len(Measure)], Measure) == 0 {
			msg.Status.Energy = &Energy{}

			msg.Status.Energy.Kind = msg.Payload[len(msg.Payload)-5]
			msg.Status.Energy.Value = int64(uint32(msg.Payload[len(msg.Payload)-3])<<16 +
				uint32(msg.Payload[len(msg.Payload)-2])<<8 +
				uint32(msg.Payload[len(msg.Payload)-1]))
		} else {
			state := msg.Payload[len(msg.Payload)-1]
			if state == 0xFF {
				state := true
				msg.Status.Switch = &state
			}
			if state == 0x00 {
				state := false
				msg.Status.Switch = &state
			}
			//TODO log unknown state reported
		}
	}

	return msg, nil
}

func MustParse(payload []byte) *Message {
	msg, err := Parse(payload)
	if err != nil {
		panic(err)
	}
	return msg
}

func (m Message) ToBytes() []byte {
	length := []byte{
		byte(uint(len(m.Payload)) >> 8),
		byte(len(m.Payload) & 255),
	}

	var msg []byte
	msg = append(msg, prefix...)
	msg = append(msg, length...)
	msg = append(msg, m.Payload...)
	msg = append(msg, Checksum(m.Payload))
	msg = append(msg, postfix...)

	return msg
}

func (m Message) String() string {
	str := fmt.Sprintf("len: %d, cmd: %s, data: %s", len(m.Payload), m.Command, hex.EncodeToString(m.Payload))

	if m.Status != nil {
		str += "\n\tStatus = "
		if m.Status.Switch != nil {
			str += fmt.Sprintf("switch: %t", *m.Status.Switch)
		}
		if m.Status.Energy != nil {
			str += fmt.Sprintf("energy: %+v", *m.Status.Energy)
		}
	}

	return str
}
