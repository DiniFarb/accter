package accter

import (
	"encoding/binary"
	"errors"
	"strconv"
)

const (

	// MaxPacketLength is the maximum length of a RADIUS packet.
	MaxPacketLength = 4096

	// Standard RADIUS ACCOUNTING packet codes.
	CodeAccountingRequest  Code = 4
	CodeAccountingResponse Code = 5
)

// Code defines the RADIUS packet type.
type Code int

type RequestPacket struct {
	Code          Code
	Length        uint16
	Identifier    uint8
	Secret        []byte
	Authenticator [16]byte
	Attributes    []RadiusAttribute
}

type RadiusAttribute struct {
	Type   uint8
	Length uint8
	Value  []byte
}

/*
 * The byte array b is the raw packet data and should be at least 20 bytes long. While:
 *  - b[0] is the packet code
 *  - b[1] is the packet identifier
 *  - b[2:4] is the packet length
 *  - b[4:20] is the packet authenticator
 *  - b[20:rest] are the packet attributes
 */
func ParsePacket(b, secret []byte) (*RequestPacket, error) {
	if len(b) < 20 {
		return nil, errors.New("radius packet not at least 20 bytes long")
	}

	length := int(binary.BigEndian.Uint16(b[2:4]))
	if length < 20 || length > MaxPacketLength || len(b) < length {
		return nil, errors.New("radius invalid packet length")
	}

	attrs, err := ParseAttributes(b[20:length])
	if err != nil {
		return nil, err
	}

	packet := &RequestPacket{
		Code:       Code(b[0]),
		Identifier: uint8(b[1]),
		Length:     uint16(length),
		Secret:     secret,
		Attributes: attrs,
	}
	copy(packet.Authenticator[:], b[4:20])
	return packet, nil
}

func ParseAttributes(b []byte) ([]RadiusAttribute, error) {
	attrs := make([]RadiusAttribute, 0)
	for len(b) > 0 {
		if len(b) < 2 {
			return nil, errors.New("invalid attribute: short buffer")
		}
		length := int(b[1])
		if length > len(b) || length < 2 || length > 255 {
			return nil, errors.New("invalid attribute length")
		}
		avp := RadiusAttribute{
			Type:   b[0],
			Length: b[1],
			Value:  b[2:length],
		}
		attrs = append(attrs, avp)
		b = b[length:]
	}
	return attrs, nil
}

func (c Code) String() string {
	switch c {
	case CodeAccountingRequest:
		return `Accounting-Request`
	case CodeAccountingResponse:
		return `Accounting-Response`
	default:
		return "Unsupported(" + strconv.Itoa(int(c)) + ")"
	}
}
