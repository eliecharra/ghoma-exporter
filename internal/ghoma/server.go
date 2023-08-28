package ghoma

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/eliecharra/ghoma/protocol"
	"go.uber.org/zap"
)

type device struct {
	logger *zap.Logger

	ID              string
	FirmwareVersion string
	conn            net.Conn
}

func (d *device) read() (*protocol.Message, error) {
	msg, err := protocol.ReadMessage(d.conn)
	if err != nil {
		return msg, err
	}
	d.logger.Debug("read", zap.Any("msg", msg))
	return msg, nil
}

func (d *device) write(msg protocol.Message) error {
	if _, err := d.conn.Write(msg.ToBytes()); err != nil {
		return err
	}
	d.logger.Debug("write", zap.Any("msg", msg))
	return nil
}

type handler interface {
	HandleStatus(device device, msg protocol.Message)
}

type ServerOptions struct {
	ListenAddr string
}

type Server struct {
	listener net.Listener
	quit     chan interface{}
	wg       sync.WaitGroup

	options  ServerOptions
	devices  sync.Map
	handlers []handler
}

func NewServer(options ServerOptions) *Server {
	return &Server{
		quit:    make(chan interface{}),
		options: options,
	}
}

func (s *Server) stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) Start(ctx context.Context) (err error) {
	logger := zap.L()
	listener, err := net.Listen("tcp", s.options.ListenAddr)
	if err != nil {
		return err
	}
	s.listener = listener
	s.wg.Add(1)
	logger.Info("Server listening", zap.String("address", listener.Addr().String()))

	go func() {
		<-ctx.Done()
		logger.Info("Stopping server")
		s.stop()
	}()

	s.serve()
	return nil
}

func (s *Server) serve() {
	defer s.wg.Done()

	for {
		c, err := s.listener.Accept()
		logger := zap.L()
		if c != nil {
			logger = zap.L().With(zap.String("remote", c.RemoteAddr().String()))
		}
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				logger.Error("error accepting new connection", zap.Error(err))
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			logger.Info("Device connected")
			s.handleDevice(c)
			s.wg.Done()
		}()
	}
}

func (s *Server) handleDevice(c net.Conn) {
	defer c.Close()
	logger := zap.L()

	dev, err := s.register(logger, c)
	if err != nil {
		logger.Error("unable to register device", zap.Error(err))
		return
	}

	logger = zap.L().With(zap.String("device_id", dev.ID))
	dev.logger = logger
	logger.Info("Device registered", zap.String("firmware_version", dev.FirmwareVersion))

	for {
		msg, err := dev.read()
		if err != nil {
			if msg != nil {
				logger = logger.With(zap.Any("msg", msg))
			}
			if errors.Is(err, protocol.ErrCmdUnknown) {
				logger.Debug("unknown command")
				continue
			}
			logger.Error("read error", zap.Error(err))
			return
		}
		s.handle(dev, msg)
	}
}

func (s *Server) register(logger *zap.Logger, c net.Conn) (*device, error) {
	dev := &device{
		logger: logger,
		conn:   c,
	}

	if err := dev.write(*protocol.MustParse(protocol.Init1)); err != nil {
		return nil, err
	}

	msg, err := dev.read()
	if err != nil {
		return nil, err
	}

	if err := dev.write(*protocol.MustParse(protocol.Init1ACK)); err != nil {
		return nil, err
	}

	dev.ID = hex.EncodeToString(msg.Payload[6:9])

	if err := dev.write(*protocol.MustParse(protocol.Init2)); err != nil {
		return nil, err
	}

	_, err = dev.read()
	if err != nil {
		return nil, err
	}

	msg, err = dev.read()
	if err != nil {
		return nil, err
	}

	dev.FirmwareVersion = fmt.Sprintf(
		"%d.%d.%d",
		msg.Payload[len(msg.Payload)-3],
		msg.Payload[len(msg.Payload)-2],
		msg.Payload[len(msg.Payload)-1],
	)

	s.devices.Store(dev.ID, dev)

	return dev, nil
}

func (s *Server) handle(dev *device, msg *protocol.Message) {
	switch msg.Command {
	case protocol.CmdHeartBeat:
		if err := dev.write(*protocol.MustParse(protocol.HeartBeatReply)); err != nil {
			// TODO log error unable to send heartbeat
			_ = ""
		}
	case protocol.CmdStatus:
		for _, h := range s.handlers {
			h.HandleStatus(*dev, *msg)
		}
	}
}
