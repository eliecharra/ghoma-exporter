package ghoma

import (
	"encoding/hex"
	"fmt"
	"net"

	"github.com/eliecharra/ghoma/protocol"
)

type device struct {
	ID              string
	FirmwareVersion string
	conn            net.Conn
}

func (d *device) Read() (*protocol.Message, error) {
	msg, err := protocol.ReadMessage(d.conn)
	if err != nil {
		return msg, err
	}
	fmt.Printf("READ: %s\n", msg)
	return msg, nil
}

func (d *device) Write(msg protocol.Message) error {
	if _, err := d.conn.Write(msg.ToBytes()); err != nil {
		return err
	}
	fmt.Printf("WRITE: %s\n", msg)
	return nil
}

type handler interface {
	HandleStatus(device device, msg protocol.Message)
}

type Server struct {
	devices  map[string]*device
	handlers []handler
}

func NewServer() *Server {
	return &Server{
		devices: make(map[string]*device),
	}
}

func (s *Server) Start() (err error) {
	srv, err := net.Listen("tcp", ":4196")
	if err != nil {
		return err
	}
	defer func(srv net.Listener) {
		err = srv.Close()
		return
	}(srv)

	for {
		c, err := srv.Accept()
		if err != nil {
			fmt.Printf("Error connecting: %s", err)
			continue
		}
		fmt.Printf("Plug connected: %s\n", c.RemoteAddr().String())

		dev, err := s.register(c)
		if err != nil {
			fmt.Printf("Error to register device: %s", err)
			continue
		}

		fmt.Printf("Device registrered, ID=%s, Version=%s\n", dev.ID, dev.FirmwareVersion)

		go func() {
			for {
				msg, err := dev.Read()
				if err != nil {
					fmt.Printf("error reading: %s\n", err)
					if msg != nil {
						fmt.Printf("\tmessage: %s", msg)
					}
					// TODO delete device?
					return
				}
				s.handle(dev, msg)
			}
		}()
	}
}

func (s *Server) register(c net.Conn) (*device, error) {

	dev := &device{
		conn: c,
	}

	if err := dev.Write(*protocol.MustParse(protocol.Init1)); err != nil {
		return nil, err
	}

	msg, err := dev.Read()
	if err != nil {
		return nil, err
	}

	if err := dev.Write(*protocol.MustParse(protocol.Init1ACK)); err != nil {
		return nil, err
	}

	dev.ID = hex.EncodeToString(msg.Payload[6:9])

	if err := dev.Write(*protocol.MustParse(protocol.Init2)); err != nil {
		return nil, err
	}

	_, err = dev.Read()
	if err != nil {
		return nil, err
	}

	msg, err = dev.Read()
	if err != nil {
		return nil, err
	}

	dev.FirmwareVersion = fmt.Sprintf(
		"%d.%d.%d",
		msg.Payload[len(msg.Payload)-3],
		msg.Payload[len(msg.Payload)-2],
		msg.Payload[len(msg.Payload)-1],
	)

	s.devices[dev.ID] = dev

	return dev, nil
}

func (s *Server) handle(dev *device, msg *protocol.Message) {
	switch msg.Command {
	case protocol.CmdHeartBeat:
		if err := dev.Write(*protocol.MustParse(protocol.HeartBeatReply)); err != nil {
			// TODO log error unable to send heartbeat
		}
	case protocol.CmdStatus:
		for _, h := range s.handlers {
			h.HandleStatus(*dev, *msg)
		}
	}
}
