package protocol

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessage_ToBytes(t *testing.T) {
	tests := []struct {
		name string
		m    Message
		want []byte
	}{
		{
			name: "init1 message",
			m:    Message{Payload: []byte{0x02, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12}},
			want: []byte{
				0x5A, 0xA5, // Prefix
				0x00, 0x07, // Length
				0x02, 0x05, 0x0d, 0x07, 0x05, 0x07, 0x12, // Payload
				0xC6,       // Checksum
				0x5B, 0xB5, // Postfix
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.m.ToBytes())
		})
	}
}
