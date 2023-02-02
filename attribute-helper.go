package accter

import (
	"encoding/binary"
	"fmt"
	"net"
)

type AttributeHelper struct {
	Name   string
	Parser func([]byte) string
}

var Attributes = map[int]AttributeHelper{
	1: {"User-Name", parseString},
	//2: 	{"User-Password",parseparseString}, NOT USED IN ACCOUNTING
	//3: 	{"CHAP-Password",parseparseString}, NOT USED IN ACCOUNTING
	4:  {"NAS-IP-Address", parseIPAddr},
	5:  {"NAS-Port", parseInteger},
	6:  {"Service-Type", parseInteger},
	7:  {"Framed-Protocol", parseInteger},
	8:  {"Framed-IP-Address", parseIPAddr},
	9:  {"Framed-IP-Netmask", parseIPAddr},
	10: {"Framed-Routing", parseInteger},
	11: {"Filter-Id", parseString},
	12: {"Framed-MTU", parseInteger},
	13: {"Framed-Compression", parseInteger},
	14: {"Login-IP-Host", parseIPAddr},
	15: {"Login-Service", parseInteger},
	16: {"Login-TCP-Port", parseInteger},
	//	18: {"Reply-Message",parseString}, NOT USED IN ACCOUNTING
	19: {"Callback-Number", parseString},
	20: {"Callback-Id", parseString},
	22: {"Framed-Route", parseString},
	23: {"Framed-IPX-Network", parseInteger},
	//	24: {"State",parseString}, NOT USED IN ACCOUNTING
	25: {"Class", parseString},
	26: {"Vendor-Specific", parseString},
	27: {"Session-Timeout", parseInteger},
	28: {"Idle-Timeout", parseInteger},
	29: {"Termination-Action", parseInteger},
	30: {"Called-Station-Id", parseString},
	31: {"Calling-Station-Id", parseString},
	32: {"NAS-Identifier", parseString},
	33: {"Proxy-State", parseString},
	34: {"Login-LAT-Service", parseString},
	35: {"Login-LAT-Node", parseString},
	36: {"Login-LAT-Group", parseString},
	37: {"Framed-AppleTalk-Link", parseInteger},
	38: {"Framed-AppleTalk-Network", parseInteger},
	39: {"Framed-AppleTalk-Zone", parseString},
	40: {"Acct-Status-Type", parseInteger},
	41: {"Acct-Delay-Time", parseInteger},
	42: {"Acct-Input-Octets", parseInteger},
	43: {"Acct-Output-Octets", parseInteger},
	44: {"Acct-Session-Id", parseString},
	45: {"Acct-Authentic", parseInteger},
	46: {"Acct-Session-Time", parseInteger},
	47: {"Acct-Input-Packets", parseInteger},
	48: {"Acct-Output-Packets", parseInteger},
	49: {"Acct-Terminate-Cause", parseInteger},
	50: {"Acct-Multi-Session-Id", parseString},
	51: {"Acct-Link-Count", parseInteger},
	52: {"Acct-Input-Gigawords", parseInteger},
	53: {"Acct-Output-Gigawords", parseInteger},
	//  60: {"CHAP-Challenge", parseString}, NOT USED IN ACCOUNTING
	61: {"NAS-Port-Type", parseInteger},
	62: {"Port-Limit", parseInteger},
	63: {"Login-LAT-Port", parseString},
}

func parseInteger(b []byte) string {
	if len(b) != 4 {
		return fmt.Sprintf("Wrong length %d", len(b))
	}
	return fmt.Sprintf("%d", binary.BigEndian.Uint32(b))
}

func parseString(b []byte) string {
	return string(b)
}

func parseIPAddr(b []byte) string {
	if len(b) != net.IPv4len {
		return fmt.Sprintf("Wrong length %d", len(b))
	}
	bs := make([]byte, net.IPv4len)
	copy(bs, b)
	return net.IP(bs).String()
}
