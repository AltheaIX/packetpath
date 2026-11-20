package forwarder

import "testing"

func TestIHLConversion(t *testing.T) {
	data := []byte{69} // 0100 0101, Version: 4, Header Length: 20 bytes (5)
	got := int(data[0]&0x0f) * 4
	want := 20

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
