package accter

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const NETWORK_TYPE = "udp"

type PacketServer struct {
	Port           int
	Secret         string
	Retransmission RetransmissionHandler
	HandleRequest  func(*JsonPacket) error
	Logger         Logger
}

func (s *PacketServer) Serve() error {
	if s.HandleRequest == nil {
		return errors.New("server has no handler")
	}
	if s.Secret == "" {
		return errors.New("server has no secret source")
	}
	if s.Logger == (Logger{}) {
		s.Logger = Logger{
			timeformat: "2006-01-02 15:04:05",
			appName:    "ACCTER",
			level:      Info,
		}
	}
	if s.Retransmission == nil {
		s.Retransmission = CreateLocalRetransmissionHandler()
	}
	if s.Port == 0 {
		s.Port = 1813
	}
	s.Logger.info("starting server on :%d", s.Port)
	conn, err := net.ListenUDP(NETWORK_TYPE, &net.UDPAddr{Port: s.Port})
	if err != nil {
		return fmt.Errorf("listen to UDP failed with: %v", err)
	}
	defer conn.Close()
	for {
		var buff = make([]byte, MaxPacketLength)
		n, remoteAddr, err := conn.ReadFromUDP(buff[:])
		if err != nil {
			s.Logger.error("error reading from connection: %v", err)
			continue
		}
		go func(buff []byte, remoteAddr net.Addr) {
			res, err := s.handlePacket(buff, remoteAddr)
			if err != nil {
				s.Logger.warn("processing packet faild with: %v", err)
				return
			}
			if res == nil {
				// this a retransmission => no response
				return
			}
			if _, err := conn.WriteTo(res, remoteAddr); err != nil {
				s.Logger.error("error sending response to [%s]: %v", remoteAddr, err)
				return
			}
		}(append([]byte(nil), buff[:n]...), remoteAddr)
	}
}

// TODO: add graceful shutdown
func (s *PacketServer) Shutdown() error {
	return errors.New("not implemented yet")
}

func (s *PacketServer) handlePacket(b []byte, remoteAddr net.Addr) ([]byte, error) {
	jsonPacket := NewRadiusJsonPacket()
	s.Logger.debug(fmt.Sprintf("received packet from: %s with id: %#x", remoteAddr, b[1]))
	jsonPacket.RemoteAddr = remoteAddr.String()
	packet, err := ParsePacket(b, []byte(s.Secret))
	if err != nil {
		err := fmt.Errorf(fmt.Sprintf("[packet-%#x] unable to parse bytes: %v", err, b[1]))
		return nil, err
	}
	jsonPacket.Id = fmt.Sprintf("%#x", b[1])
	jsonPacket.Authenticator = fmt.Sprintf("%x", b[4:20])
	jsonPacket.Code = Code(b[0]).String()
	jsonPacket.Key = jsonPacket.Id + "_" + jsonPacket.Authenticator
	if packet.Code != CodeAccountingRequest {
		err := fmt.Errorf("[packet-%#x] only accounting request is supported", b[1])
		return nil, err
	}
	s.Logger.trace(fmt.Sprintf("[packet-%#x] checking retransmission key: %s", b[1], jsonPacket.Key))
	isRetry := s.Retransmission.IsRetransmission(jsonPacket.Key)
	if isRetry {
		// TODO: Add Option to allow retransmission
		s.Logger.debug(fmt.Sprintf("[packet-%#x] retransmission detected with key: %s", b[0], jsonPacket.Key))
		return nil, nil
	} else {
		s.Retransmission.AddToCache(jsonPacket.Key)
	}
	s.Logger.trace(fmt.Sprintf("[packet-%#x] transform attributes as strings", b[1]))
	for _, v := range packet.Attributes {
		if attr, ok := Attributes[int(v.Type)]; ok {
			val := attr.Parser(v.Value)
			a := JsonAttribute{Name: attr.Name, Value: val}
			jsonPacket.Attributes = append(jsonPacket.Attributes, a)
			s.Logger.trace(fmt.Sprintf("[packet-%#x] add attr: %s='%s',", b[1], attr.Name, val))
		} else {
			a := JsonAttribute{Name: fmt.Sprintf("(UNSUPPORTED) %d", v.Type), Value: string(v.Value)}
			jsonPacket.Attributes = append(jsonPacket.Attributes, a)
			s.Logger.trace(fmt.Sprintf("[packet-%#x] %d='%v (UNSUPPORTED)'", b[1], v.Type, string(v.Value)))
		}
	}
	err = s.HandleRequest(jsonPacket)
	if err != nil {
		return nil, err
	} else {
		writebytes := make([]byte, 20)
		writebytes[0] = byte(CodeAccountingResponse)
		writebytes[1] = packet.Identifier
		binary.BigEndian.PutUint16(writebytes[2:4], 20)
		hash := md5.New()
		hash.Write(writebytes[:4])
		hash.Write(packet.Authenticator[:])
		hash.Write(packet.Secret)
		hash.Sum(writebytes[4:4:20])
		s.Logger.trace(fmt.Sprintf("[packet-%#x] send response", b[1]))
		return writebytes, nil
	}
}
