package ghoma

import (
	"net"

	"go.uber.org/zap"

	"github.com/eliecharra/ghoma/protocol"
)

type Device struct {
	logger *zap.Logger

	ID              string
	FirmwareVersion string
	conn            net.Conn
}

func (d *Device) read() (*protocol.Message, error) {
	msg, err := protocol.ReadMessage(d.conn)
	if err != nil {
		return msg, err
	}
	d.logger.Debug("read", zap.Any("msg", msg))
	return msg, nil
}

func (d *Device) write(msg protocol.Message) error {
	if _, err := d.conn.Write(msg.ToBytes()); err != nil {
		return err
	}
	d.logger.Debug("write", zap.Any("msg", msg))
	return nil
}
