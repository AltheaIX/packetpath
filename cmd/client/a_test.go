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
