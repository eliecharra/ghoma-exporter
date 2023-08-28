package protocol

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMessage(t *testing.T) {
	tests := []struct {
		name          string
		reader        io.Reader
		want          *Message
		expectedError string
	}{
		{
			name:          "invalid message",
			reader:        bytes.NewReader([]byte{0x5a}),
			expectedError: "unable to read message header",
		},
		{
			name:          "invalid payload",
			reader:        bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0x02}),
			expectedError: "unable to read payload",
		},
		{
			name:          "no checksum",
			reader:        bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0x02, 0x05, 0x0d, 0x07, 0x05, 0x08, 0x12}),
			expectedError: "unable to read checksum byte",
		},
		{
			name:          "invalid checksum",
			reader:        bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0x02, 0x05, 0x0d, 0x07, 0x05, 0x08, 0x12, 0xc6, 0x5b, 0xb5}),
			expectedError: "invalid checksum",
		},
		{
			name:          "no postfix",
			reader:        bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0x02, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12, 0xc6}),
			expectedError: "unable to read message postfix",
		},
		{
			name:          "invalid command message",
			reader:        bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0xFF, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12, 0xc9, 0x5b, 0xb5}),
			expectedError: "unknown command",
			want:          &Message{Payload: []byte{0xFF, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12}, Command: 0xFF},
		},
		{
			name:   "valid message",
			reader: bytes.NewReader([]byte{0x5a, 0xa5, 0x00, 0x07, 0x02, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12, 0xc6, 0x5b, 0xb5}),
			want:   &Message{Payload: []byte{0x02, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12}, Command: 0x02},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ReadMessage(tt.reader)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.expectedError)
			}
			require.Equal(t, tt.want, msg)
		})
	}
}
