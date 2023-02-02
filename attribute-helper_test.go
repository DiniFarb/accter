package accter

import "testing"

func TestParseIPAddr(t *testing.T) {
	want := "10.10.10.1"
	got := parseIPAddr([]byte{10, 10, 10, 1})
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseIPAddrFail(t *testing.T) {
	want := "Wrong length 5"
	got := parseIPAddr([]byte{10, 10, 10, 1, 1})
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseInteger(t *testing.T) {
	want := "10"
	got := parseInteger([]byte{0, 0, 0, 10})
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseIntegerFail(t *testing.T) {
	want := "Wrong length 5"
	got := parseInteger([]byte{0, 0, 0, 0, 10})
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseString(t *testing.T) {
	want := "test"
	got := parseString([]byte{116, 101, 115, 116})
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
