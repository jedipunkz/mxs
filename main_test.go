package main

import "testing"

func TestUUIDFormat(t *testing.T) {
	if got, want := floodgateUUID(2535428654845283), "00000000-0000-0000-0009-01f57c534d63"; got != want {
		t.Errorf("floodgateUUID = %s, want %s", got, want)
	}
	if got, err := parseFloodgateUUID("00000000-0000-0000-0009-01f57c534d63"); err != nil || got != 2535428654845283 {
		t.Errorf("parseFloodgateUUID = %d, %v", got, err)
	}
	if _, err := parseFloodgateUUID("069a79f4-44e9-4726-a5be-fca90e38aaf5"); err == nil {
		t.Error("parseFloodgateUUID accepted a non-floodgate uuid")
	}
	if got, want := dashed("069a79f444e94726a5befca90e38aaf5"), "069a79f4-44e9-4726-a5be-fca90e38aaf5"; got != want {
		t.Errorf("dashed = %s, want %s", got, want)
	}
}
