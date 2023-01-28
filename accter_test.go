package accter

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

const (
	testHexStr = "04%s005883aa46a746f4a9cc25e9037492695d7c0114746573747573657240646f6d61696e2e63682c" +
		"07742d3830302806000000030506000000050406010203042a06000004d22b060000ddd5190654455354200570616e"
	testAuthenticator = "83aa46a746f4a9cc25e9037492695d7c"
	testSessionId     = "t-800"
)

var once sync.Once

func GetTestPacket(id byte) []byte {
	decodedByteArray, _ := hex.DecodeString(fmt.Sprintf(testHexStr, hex.EncodeToString([]byte{id})))
	return decodedByteArray
}

func ExchangePacket(ctx context.Context, b []byte, addr string) (Code, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	conn.Write(b)
	p := make([]byte, 2048)
	_, err = bufio.NewReader(conn).Read(p)
	if err != nil {
		return 0, err
	}
	if err := CheckAuthenticator(b[4:20], p); err != nil {
		return 0, err
	}
	return Code(p[0]), nil
}

/*
Important: this is not a real authenticator check, this is bacause we assume that the
length of the answer is always 20 bytes long. This is only true cos accter does not send
any attributes in the response and therfore the length is always 20 bytes long.
*/
func CheckAuthenticator(sendAuthenticator, p []byte) error {
	got := make([]byte, 16)
	copy(got, p[4:20])
	want := make([]byte, 20)
	want[0] = p[0]
	want[1] = p[1]
	binary.BigEndian.PutUint16(want[2:4], 20)
	hash := md5.New()
	hash.Write(want[:4])
	hash.Write(sendAuthenticator)
	hash.Write([]byte("secret"))
	hash.Sum(want[4:4:20])
	if !bytes.Equal(got, want[4:20]) {
		return fmt.Errorf("authenticator mismatch want: %x got: %x", want[4:20], got)
	}
	return nil
}

func StartTestServer(f func(packet *JsonPacket) error) {
	once.Do(func() {
		server := PacketServer{
			Port:          1813,
			Secret:        "secret",
			HandleRequest: f,
		}
		server.Logger.level = 1
		go func(server PacketServer) {
			if err := server.Serve(); err != nil {
				panic("error starting server: " + err.Error())
			}
		}(server)
	})
}

func TestOneRequest(t *testing.T) {
	var gotAuthenticator string
	var gotId string
	var gotSessionId string
	handler := func(packet *JsonPacket) error {
		gotAuthenticator = packet.Authenticator
		gotId = packet.Id
		for _, attr := range packet.Attributes {
			if attr.Name == "Acct-Session-Id" {
				gotSessionId = attr.Value
			}
		}
		return nil
	}
	StartTestServer(handler)
	time.Sleep(1 * time.Second)
	id := byte(1)
	testPacket := GetTestPacket(id)
	response, err := ExchangePacket(context.Background(), testPacket, "localhost:1813")
	if err != nil {
		t.Errorf(err.Error())
	}
	wantResp := CodeAccountingResponse
	if response != wantResp {
		t.Errorf("response.Code = %q, want %q", response, wantResp)
	}
	if gotAuthenticator != testAuthenticator {
		t.Errorf("Authenticator = %q, want %q", gotAuthenticator, testAuthenticator)
	}
	if gotId != fmt.Sprintf("%#x", id) {
		t.Errorf("Id = %q, want %q", gotId, string(id))
	}
	if gotSessionId != testSessionId {
		t.Errorf("Acct-Session-Id = %q, want %q", gotSessionId, testSessionId)
	}
}

func TestManyRequest(t *testing.T) {
	handler := func(packet *JsonPacket) error {
		return nil
	}
	StartTestServer(handler)
	time.Sleep(1 * time.Second)
	for i := 0; i < 1000; i++ {
		id := byte(i)
		testPacket := GetTestPacket(id)
		go func() {
			response, err := ExchangePacket(context.Background(), testPacket, "localhost:1813")
			if err != nil {
				t.Errorf(err.Error())
			}
			want := CodeAccountingResponse
			if response != want {
				t.Errorf("response.Code = %q, want %q", response, want)
			}
		}()
	}
}
