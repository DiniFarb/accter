package accter

type JsonPacket struct {
	Id            string          `json:"id"`
	Authenticator string          `json:"authenticator"`
	Code          string          `json:"code"`
	Key           string          `json:"key"`
	RemoteAddr    string          `json:"remote_addr"`
	Attributes    []JsonAttribute `json:"attributes"`
}

type JsonAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func NewRadiusJsonPacket() *JsonPacket {
	return &JsonPacket{}
}
