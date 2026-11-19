package main

import (
	"testing"
)

func TestIpv4ToString(t *testing.T) {
	got := Ipv4ToString(3251269830)
	want := "193.202.112.198"

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIHLConversion(t *testing.T) {
	data := []byte{69} // 0100 0101, Version: 4, Header Length: 20 bytes (5)
	got := int(data[0]&0x0f) * 4
	want := 20

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
