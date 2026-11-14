package main

import (
	"fmt"
	"testing"
)

func TestIpv4ToString(t *testing.T) {
	got := Ipv4ToString(3251269830)
	want := "193.202.112.198"

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAsd(t *testing.T) {
	var x uint32 = 0xFF000000

	fmt.Printf("a        = %032b  (%d)\n", x, x)
	fmt.Printf("a        = %032b  (%d)\n", x<<7, x<<7)
}
