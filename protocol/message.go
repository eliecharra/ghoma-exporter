package protocol

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type Status struct {
	Switch *bool
	Energy *Energy
}

type Energy struct {
	kind  uint8
	value int64
}

func (e *Energy) Kind() string {
	switch e.kind {
	case 1:
		return "POWER"
	case 2:
		return "ENERGY"
	case 3:
		return "VOLTAGE"
	case 4:
		return "CURRENT"
	case 5:
		return "FREQUENCY"
	case 7:
		return "MAX_POWER"
	case 8:
		return "COSPHI"
	default:
		return "UNKNOWN"
	}
}

func (e *Energy) Value() int64 {
	return e.value
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
		if bytes.Equal(msg.Payload[9:9+len(Measure)], Measure) {
			msg.Status.Energy = &Energy{}

			msg.Status.Energy.kind = msg.Payload[len(msg.Payload)-5]
			msg.Status.Energy.value = int64(uint32(msg.Payload[len(msg.Payload)-3])<<16 +
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

func (m Message) MarshalJSON() ([]byte, error) {
	type energy struct {
		Kind  string `json:"kind"`
		Value int64  `json:"value"`
	}

	type status struct {
		Switch string  `json:"switch,omitempty"`
		Energy *energy `json:"energy,omitempty"`
	}

	data := &struct {
		Payload string  `json:"payload,omitempty"`
		Command string  `json:"command"`
		Status  *status `json:"status,omitempty"`
	}{
		Command: m.Command.String(),
	}

	if data.Command == "UNKNOWN" {
		data.Payload = hex.EncodeToString(m.Payload)
	}

	if m.Status != nil {
		data.Status = &status{}
		if m.Status.Switch != nil {
			switch *m.Status.Switch {
			case true:
				data.Status.Switch = "ON"
			case false:
				data.Status.Switch = "OFF"
			}
		}
		if m.Status.Energy != nil {
			e := &energy{}
			e.Kind = m.Status.Energy.Kind()
			e.Value = m.Status.Energy.Value()
			data.Status.Energy = e
		}
	}

	return json.Marshal(data)
}
