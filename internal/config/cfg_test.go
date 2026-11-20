package config

import "testing"

func TestParsePort_ValidPort(t *testing.T) {
	port := "65535"
	get, err := ParsePort(port)
	if err != nil {
		t.Fatal(err)
	}

	if get != uint16(65535) {
		t.Fatalf("Port %s is not valid", port)
	}
}

func TestParsePort_InvalidPort(t *testing.T) {
	port := "70000"
	get, err := ParsePort(port)
	if err == nil {
		t.Fatal(err)
	}

	if get != 0 {
		t.Fatal("port out of range")
	}
}
