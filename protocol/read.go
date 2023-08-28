package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

func ReadMessage(r io.Reader) (*Message, error) {
	reader := bufio.NewReader(r)

	// ReadMessage header (prefix + payload length)
	buffer := make([]byte, 4)
	if n, err := reader.Read(buffer); n != 4 || err != nil {
		return nil, errors.Join(errors.New("unable to read message header"), err)
	}
	length := uint16(buffer[2])<<8 | uint16(buffer[3])

	// ReadMessage payload based on length declared in header
	payload := make([]byte, length)
	if n, err := reader.Read(payload); n != int(length) || err != nil {
		return nil, errors.Join(errors.New("unable to read payload"), err)
	}

	// The next byte should be the checksum byte, check it against the payload
	checksumByte, err := reader.ReadByte()
	if err != nil {
		return nil, errors.Join(errors.New("unable to read checksum byte"), err)
	}
	if Checksum(payload) != checksumByte {
		return nil, errors.New("invalid checksum")
	}

	// For consistency, check that the payload ends with the postfix
	buffer = make([]byte, 2)
	if n, err := reader.Read(buffer); n != 2 || err != nil || bytes.Compare(buffer, postfix) != 0 {
		return nil, errors.Join(errors.New("unable to read message postfix"), err)
	}

	msg, err := Parse(payload)
	if err != nil {
		return msg, err
	}

	return msg, nil
}
