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

func GetTestPacket(id byte) []byte {
	decodedByteArray, _ := hex.DecodeString(fmt.Sprintf(testHexStr, hex.EncodeToString([]byte{id})))
	return decodedByteArray
}

func ExchangePacket(ctx context.Context, b []byte, addr string) (Code, error) {
	d := &net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", addr)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	_, err = conn.Write(b)
	if err != nil {
		return 0, err
	}
	done := make(chan struct{})
	var packetError error
	var code Code
	go func(e error, c Code) {
		defer close(done)
		p := make([]byte, 4096)
		_, err = bufio.NewReader(conn).Read(p)
		if err != nil {
			packetError = err
			return
		}
		if err := CheckAuthenticator(b[4:20], p); err != nil {
			packetError = err
			return
		}
		code = Code(p[0])
	}(packetError, code)
	select {
	case <-done:
		return code, packetError
	case <-ctx.Done():
		return 0, ctx.Err()
	}
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

func StartTestServer(f func(packet *JsonPacket) error, r bool) {
	server := &PacketServer{
		Port:                1813,
		Secret:              "secret",
		HandleRequest:       f,
		AllowRetransmission: r,
	}
	server.Logger.level = 1
	go func(server *PacketServer) {
		if err := server.Serve(); err != nil {
			panic("error starting server: " + err.Error())
		}
	}(server)
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
	StartTestServer(handler, false)
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

func TestRetransRequest(t *testing.T) {
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
	StartTestServer(handler, false)
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
	c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	e, err := ExchangePacket(c, testPacket, "localhost:1813")
	fmt.Println(e)
	if err.Error() != "context deadline exceeded" {
		t.Errorf("timeout.res = %q, want (context deadline exceeded)", err.Error())
	}
	defer cancel()
}

func TestManyRequest(t *testing.T) {
	handler := func(packet *JsonPacket) error {
		return nil
	}
	StartTestServer(handler, false)
	wg := new(sync.WaitGroup)
	wg.Add(50)
	for i := 0; i < 50; i++ {
		go func(i int) {
			defer wg.Done()
			id := byte(i)
			testPacket := GetTestPacket(id)
			response, err := ExchangePacket(context.Background(), testPacket, "localhost:1813")
			if err != nil {
				t.Errorf(err.Error())
			}
			want := CodeAccountingResponse
			if response != want {
				t.Errorf("response.Code = %q, want %q", response, want)
			}
		}(i)
	}
	wg.Wait()
}

func TestAllowRetrans(t *testing.T) {
	handler := func(packet *JsonPacket) error {
		return nil
	}
	StartTestServer(handler, true)
	wg := new(sync.WaitGroup)
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			id := byte(1)
			testPacket := GetTestPacket(id)
			response, err := ExchangePacket(context.Background(), testPacket, "localhost:1813")
			if err != nil {
				t.Errorf(err.Error())
			}
			want := CodeAccountingResponse
			fmt.Println(i, response)
			if response != want {
				t.Errorf("response.Code = %q, want %q", response, want)
			}
		}(i)
	}
	wg.Wait()
}
