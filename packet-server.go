package accter

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const NETWORK_TYPE = "udp"

type PacketServer struct {
	Port                int
	Secret              string
	AllowRetransmission bool
	Retransmission      RetransmissionHandler
	HandleRequest       func(*JsonPacket) error
	LogLevel            Level
	Routines            Routines
	shutdown            chan bool
}

type Routines struct {
	mu    sync.RWMutex
	count int
}

func (s *PacketServer) Serve() error {
	s.shutdown = make(chan bool)
	if s.HandleRequest == nil {
		return errors.New("server has no handler")
	}
	if s.Secret == "" {
		return errors.New("server has no secret source")
	}
	if s.Retransmission == nil && !s.AllowRetransmission {
		s.Retransmission = CreateLocalRetransmissionHandler()
	}
	if s.Port == 0 {
		s.Port = 1813
	}
	if s.LogLevel == 0 {
		CreateLogger(Info)
	} else {
		CreateLogger(s.LogLevel)
	}
	logger.info("starting server on :%d", s.Port)
	conn, err := net.ListenUDP(NETWORK_TYPE, &net.UDPAddr{Port: s.Port})
	if err != nil {
		return fmt.Errorf("listen to UDP failed with: %v", err)
	}
	defer conn.Close()
	for {
		var buff = make([]byte, MaxPacketLength)
		n, remoteAddr, err := conn.ReadFromUDP(buff[:])
		if err != nil {
			logger.error("error reading from connection: %v", err)
			continue
		}
		// breake if we got shutdown signal
		select {
		case <-s.shutdown:
			logger.info("shutdown signal received, shutting down")
			return nil
		default:
			go func(buff []byte, remoteAddr net.Addr) {
				s.IncrementRoutineCount()
				defer s.DecreaseRoutineCount()
				res, err := s.handlePacket(buff, remoteAddr)
				if err != nil {
					logger.warn("processing packet faild with: %v", err)
					return
				}
				if res == nil {
					// this a retransmission => no response
					return
				}
				if _, err := conn.WriteTo(res, remoteAddr); err != nil {
					logger.error("error sending response to [%s]: %v", remoteAddr, err)
					return
				}
			}(append([]byte(nil), buff[:n]...), remoteAddr)
		}
	}
}

func (s *PacketServer) Shutdown() {
	s.shutdown <- true
	for s.GetRoutineCount() > 0 {
		time.Sleep(1 * time.Second)
	}
	logger.info("all routines finished, server shutdown")
}

func (s *PacketServer) IncrementRoutineCount() {
	if s.Routines == (Routines{}) {
		s.Routines = Routines{
			mu:    sync.RWMutex{},
			count: 0,
		}
	}
	s.Routines.mu.Lock()
	defer s.Routines.mu.Unlock()
	s.Routines.count++
	logger.trace("new routine started, now we have %d active routines", s.Routines.count)
}

func (s *PacketServer) DecreaseRoutineCount() {
	s.Routines.mu.Lock()
	defer s.Routines.mu.Unlock()
	if s.Routines.count == 0 {
		logger.warn("routine count is already 0, something is wrong")
		return
	}
	logger.trace("routine finished, now we have %d active routines", s.Routines.count)
}

func (s *PacketServer) GetRoutineCount() int {
	s.Routines.mu.RLock()
	defer s.Routines.mu.RUnlock()
	return s.Routines.count
}

func (s *PacketServer) handlePacket(b []byte, remoteAddr net.Addr) ([]byte, error) {
	jsonPacket := NewRadiusJsonPacket()
	logger.debug(fmt.Sprintf("received packet from: %s with id: %#x", remoteAddr, b[1]))
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
	if !s.AllowRetransmission {
		logger.trace(fmt.Sprintf("[packet-%#x] checking retransmission key: %s", b[1], jsonPacket.Key))
		isRetry := s.Retransmission.IsRetransmission(jsonPacket.Key)
		if isRetry {
			logger.debug(fmt.Sprintf("[packet-%#x] retransmission detected with key: %s", b[0], jsonPacket.Key))
			return nil, nil
		} else {
			s.Retransmission.AddToCache(jsonPacket.Key)
		}
	}
	logger.trace(fmt.Sprintf("[packet-%#x] transform attributes as strings", b[1]))
	for _, v := range packet.Attributes {
		if attr, ok := Attributes[int(v.Type)]; ok {
			val := attr.Parser(v.Value)
			a := JsonAttribute{Name: attr.Name, Value: val}
			jsonPacket.Attributes = append(jsonPacket.Attributes, a)
			logger.trace(fmt.Sprintf("[packet-%#x] add attr: %s='%s',", b[1], attr.Name, val))
		} else {
			a := JsonAttribute{Name: fmt.Sprintf("(UNSUPPORTED) %d", v.Type), Value: string(v.Value)}
			jsonPacket.Attributes = append(jsonPacket.Attributes, a)
			logger.trace(fmt.Sprintf("[packet-%#x] %d='%v (UNSUPPORTED)'", b[1], v.Type, string(v.Value)))
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
		logger.trace(fmt.Sprintf("[packet-%#x] send response", b[1]))
		return writebytes, nil
	}
}
